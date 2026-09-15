# Fawry Billing Integration Plan

## Overview

Integrate Fawry (Egyptian payment gateway) as the sole billing provider. Transactions auto-create from Fawry webhooks. Subscriptions get checkout/upgrade flow via Fawry. Remove multi-currency support entirely — hardcode EGP as the default and only currency. All billing centralized in `internal/lib/billing/`.

Fawry has no native subscription object, so the recurring-charge loop, idempotency, pending expiry, cancellation semantics, and dunning are built and owned by this codebase — Fawry only executes individual charges and reports their outcome via webhook.

---

## Decisions

| Decision | Choice |
|----------|--------|
| Billing provider | Fawry only (no multi-provider abstraction) |
| Transaction source | Auto-create from webhook |
| Billing endpoints | Full CRUD (checkout, webhook, portal, history/status) |
| Plan pricing | Configurable via env vars |
| Webhook middleware | Skip ALL middleware (auth, rate limit, CORS) |
| Currency | EGP only — remove all multi-currency code |
| Webhook idempotency | Every processed event recorded in a `billing_events` ledger, keyed by `fawry_ref_number` (unique) — duplicates are dropped, not reprocessed |
| Pending charge expiry | Every charge gets `expires_at`; unresolved pending charges are auto-expired by the reconciliation job, never left open indefinitely |
| Cancellation semantics | End-of-period — subscription stays `active` until `current_period_end`, no new renewal charge is created after cancellation |
| Failed renewal handling | Dunning: grace period + bounded retries before downgrade, no immediate cutoff on a single failed charge |
| Recurring billing | Own cron/worker charges the stored card token (`customerProfileId`) each cycle — Fawry has no built-in recurring billing |
| Data retention | `billing_events` is append-only and never deleted; archival/rotation of terminal records is a documented future phase, not built in this pass |

---

## Step 1: Add Billing Config

**File:** `internal/config/config.go`

Add `BillingConfig` struct and `Billing` field to `Config`:

```go
type BillingConfig struct {
    Provider         string             `koanf:"provider" validate:"required"`          // "fawry"
    MerchantCode     string             `koanf:"merchant_code" validate:"required"`
    SecureKey        string             `koanf:"secure_key" validate:"required"`
    BaseURL          string             `koanf:"base_url" validate:"required"`          // staging or production URL
    WebhookSecret    string             `koanf:"webhook_secret" validate:"required"`
    SuccessReturnURL string             `koanf:"success_return_url" validate:"required"`
    CancelReturnURL  string             `koanf:"cancel_return_url" validate:"required"`
    WebhookURL       string             `koanf:"webhook_url" validate:"required"`       // public URL for Fawry callbacks
    PlanPrices       map[string]float64 `koanf:"plan_prices" validate:"required"`       // {"pro": 9.99, "vip": 19.99}

    // Pending-charge / dunning policy (see best-practices doc: pending must expire, failed renewals get a grace window)
    ChargeExpiryMinutes    int `koanf:"charge_expiry_minutes" validate:"required"`       // TTL for a pending charge, e.g. 30
    DunningGraceDays       int `koanf:"dunning_grace_days" validate:"required"`          // days past period end before downgrade
    DunningMaxRetries      int `koanf:"dunning_max_retries" validate:"required"`         // retry attempts within the grace period
    DunningRetryIntervalHr int `koanf:"dunning_retry_interval_hours" validate:"required"`
}
```

Add to `Config` struct:
```go
Billing BillingConfig `koanf:"billing" validate:"required"`
```

Env vars: `MONEYFLOW_BILLING_PROVIDER`, `MONEYFLOW_BILLING_MERCHANT_CODE`, `MONEYFLOW_BILLING_SECURE_KEY`, `MONEYFLOW_BILLING_BASE_URL`, `MONEYFLOW_BILLING_WEBHOOK_SECRET`, `MONEYFLOW_BILLING_SUCCESS_RETURN_URL`, `MONEYFLOW_BILLING_CANCEL_RETURN_URL`, `MONEYFLOW_BILLING_WEBHOOK_URL`, `MONEYFLOW_BILLING_PLAN_PRICES_PRO`, `MONEYFLOW_BILLING_PLAN_PRICES_VIP`, `MONEYFLOW_BILLING_CHARGE_EXPIRY_MINUTES`, `MONEYFLOW_BILLING_DUNNING_GRACE_DAYS`, `MONEYFLOW_BILLING_DUNNING_MAX_RETRIES`, `MONEYFLOW_BILLING_DUNNING_RETRY_INTERVAL_HOURS`.

---

## Step 2: Create Billing Lib

**Directory:** `internal/lib/billing/`

### `internal/lib/billing/client.go`

Fawry API client:

```go
type Client struct {
    merchantCode string
    secureKey    string
    baseURL      string
    logger       *zerolog.Logger
}

func NewClient(cfg *config.Config, logger *zerolog.Logger) *Client
```

Methods:
- `CreateCharge(userID, email, mobile, plan string, amount float64, refNum string) (*ChargeResponse, error)` — `POST /ECommerceWeb/Fawry/payments/charge`; sets `PaymentExpiry` from `cfg.Billing.ChargeExpiryMinutes` so every charge has a hard TTL, never left open indefinitely
- `VerifyWebhookSignature(body []byte, signature string) bool` — SHA-256 verification against `fawrySecureKey`; reject on mismatch, do not process
- `ParseWebhookEvent(body []byte) (*WebhookEvent, error)` — parse Fawry notification JSON
- `GetPaymentStatus(refNum string) (*ChargeResponse, error)` — `GET` status-query endpoint; used exclusively by the reconciliation job (Step 16) to resolve charges whose webhook never arrived
- `generateSignature(merchantRefNum, amount string) string` — SHA-256 of `merchantCode + merchantRefNum + amount + secureKey`

> Confirm the exact field concatenation order for both the request signature and the webhook `messageSignature` against the current Fawry Server APIs docs before implementing — do not assume the order above is authoritative.

### `internal/lib/billing/types.go`

Fawry API types:

```go
type ChargeRequest struct {
    MerchantCode      string       `json:"merchantCode"`
    MerchantRefNum    string       `json:"merchantRefNum"`
    CustomerProfileID string       `json:"customerProfileId,omitempty"`
    PaymentMethod     string       `json:"paymentMethod"`
    CustomerName      string       `json:"customerName,omitempty"`
    CustomerMobile    string       `json:"customerMobile"`
    CustomerEmail     string       `json:"customerEmail"`
    Amount            float64      `json:"amount"`
    CurrencyCode      string       `json:"currencyCode"`
    PaymentExpiry     int64        `json:"paymentExpiry,omitempty"`
    Description       string       `json:"description"`
    Language          string       `json:"language"`
    ChargeItems       []ChargeItem `json:"chargeItems"`
    OrderWebHookUrl   string       `json:"orderWebHookUrl"`
    Signature         string       `json:"signature"`
}

type ChargeItem struct {
    ItemID      string  `json:"itemId"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
    Quantity    int     `json:"quantity"`
}

type ChargeResponse struct {
    Type              string  `json:"type"`
    ReferenceNumber   string  `json:"referenceNumber"`
    MerchantRefNumber string  `json:"merchantRefNumber"`
    OrderAmount       float64 `json:"orderAmount"`
    PaymentAmount     float64 `json:"paymentAmount"`
    FawryFees         float64 `json:"fawryFees"`
    PaymentMethod     string  `json:"paymentMethod"`
    OrderStatus       string  `json:"orderStatus"`
    PaymentTime       int64   `json:"paymentTime"`
    CustomerMobile    string  `json:"customerMobile"`
    CustomerMail      string  `json:"customerMail"`
    CustomerProfileID string  `json:"customerProfileId"`
    Signature         string  `json:"signature"`
    StatusCode        int     `json:"statusCode"`
    StatusDescription string  `json:"statusDescription"`
}

type WebhookEvent struct {
    FawryRefNumber    string  `json:"fawryRefNumber"`
    MerchantRefNumber string  `json:"merchantRefNumber"`
    OrderAmount       float64 `json:"orderAmount"`
    PaymentAmount     float64 `json:"paymentAmount"`
    OrderStatus       string  `json:"orderStatus"`
    PaymentMethod     string  `json:"paymentMethod"`
    PaymentTime       int64   `json:"paymentTime"`
    CustomerMobile    string  `json:"customerMobile"`
    CustomerMail      string  `json:"customerMail"`
    CustomerProfileID string  `json:"customerProfileId"`
    StatusCode        int     `json:"statusCode"`
    StatusDescription string  `json:"statusDescription"`
    Signature         string  `json:"signature"`
}
```

Fawry status constants:
```go
const (
    OrderStatusPaid     = "PAID"
    OrderStatusNew      = "NEW"
    OrderStatusCanceled = "CANCELED"
    OrderStatusRefunded = "REFUNDED"
    OrderStatusExpired  = "EXPIRED"
    OrderStatusFailed   = "FAILED"
)
```

### `internal/lib/billing/plans.go`

Plan pricing helper (reads from config):
```go
func GetPlanAmount(plan string, prices map[string]float64) (float64, error)
```

---

## Step 3: Migration — Billing Tables & Columns

**New migration:** `internal/database/migrations/010_add_billing_tables.sql`

```sql
-- Transactions: link back to the billing event that created them
ALTER TABLE transactions
    ADD COLUMN billing_provider TEXT DEFAULT 'manual',
    ADD COLUMN billing_ref_id TEXT,
    ADD COLUMN billing_status TEXT DEFAULT 'manual';

CREATE INDEX idx_transactions_billing_ref ON transactions(billing_ref_id);

-- Subscriptions: pending-expiry + end-of-period cancellation + dunning state
-- (assumes a `subscriptions` table already exists with status/current_period_end;
--  adjust column names if they differ from what's already in the schema)
ALTER TABLE subscriptions
    ADD COLUMN expires_at TIMESTAMPTZ,              -- TTL for a pending charge/reference code
    ADD COLUMN cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN cancelled_at TIMESTAMPTZ,
    ADD COLUMN failed_attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN next_billing_at TIMESTAMPTZ;

-- Partial index: reconciliation/dunning jobs only ever scan non-terminal rows
CREATE INDEX idx_subscriptions_pending
    ON subscriptions(status)
    WHERE status IN ('pending', 'past_due');

-- Billing events: append-only idempotency ledger + audit trail.
-- Every processed Fawry notification lands here exactly once, deduped on
-- fawry_ref_number. Never deleted — this is the record for disputes/audits.
CREATE TABLE billing_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fawry_ref_number   TEXT NOT NULL UNIQUE,
    merchant_ref_number TEXT NOT NULL,
    order_status        TEXT NOT NULL,
    payload              JSONB NOT NULL,
    processed_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_billing_events_merchant_ref ON billing_events(merchant_ref_number);
```

Subscription `status` values used throughout this plan: `incomplete` (never activated) → `pending` (charge created, awaiting notification) → `active` → `past_due` (renewal failed, in dunning) → `cancelled` (end-of-period, no future charge) / `expired` (pending charge TTL passed unpaid).

---

## Step 4: Remove Multi-Currency Support

Remove all currency fields and the `CurrencyTotal` type. Hardcode `"EGP"` where needed.

### 4a. Transaction Model

**File:** `internal/model/transaction/transaction.go`
- Remove `Currency` field from `Transaction` struct

**File:** `internal/model/transaction/transaction.dto.go`
- Remove `Currency` from `CreateTransactionRequest`
- Remove `Currency` from `UpdateTransactionRequest`
- Remove `ByCurrency` field from `TransactionSummary`
- Remove entire `CurrencyTotal` struct

### 4b. Transaction Repository

**File:** `internal/repository/transaction.go`
- Remove `currency` from all SELECT, INSERT, UPDATE queries
- Remove `"currency": t.Currency` from all named args
- Remove the entire currency-totals query block in `Summary()`
- Remove `ByCurrency` initialization in `Summary()`

### 4c. Transaction Service

**File:** `internal/service/transaction.go`
- Remove currency defaulting logic (`if currency == "" { currency = "EGP" }`)
- Remove `Currency: currency` from transaction creation
- Remove `t.Currency = payload.Currency` from update

### 4d. Budget Model & DTOs

**File:** `internal/model/budget/budget.go`
- Remove `Currency` field

**File:** `internal/model/budget/budget.dto.go`
- Remove `Currency` from `CreateBudgetRequest` and `UpdateBudgetRequest`

### 4e. Budget Repository

**File:** `internal/repository/budget.go`
- Remove `currency` from all SELECT, INSERT, UPDATE queries
- Remove `"currency": b.Currency` from all named args

### 4f. Budget Service

**File:** `internal/service/budget.go`
- Remove `Currency: payload.Currency` from create and update

### 4g. Recurring Rule Model & DTOs

**File:** `internal/model/recurring/recurring.go`
- Remove `Currency` field

**File:** `internal/model/recurring/recurring.dto.go`
- Remove `Currency` from `CreateRecurringRuleRequest` and `UpdateRecurringRuleRequest`

### 4h. Recurring Rule Repository

**File:** `internal/repository/recurring.go`
- Remove `currency` from all SELECT, INSERT, UPDATE queries
- Remove `"currency": rule.Currency` from all named args

### 4i. Recurring Rule Service

**File:** `internal/service/recurring.go`
- Remove currency defaulting and assignment logic
- Remove `Currency: rule.Currency` in `processRule` transaction creation

### 4j. Import Service & Handler

**File:** `internal/service/import.go`
- Remove `Currency` from `CSVRow` struct
- Remove `Currency: row.Currency` from `processRow`

**File:** `internal/handler/import.go`
- Remove `Currency: record[2]` from CSV parsing

### 4k. User Model & Repository

**File:** `internal/model/user/user.go`
- Remove `DefaultCurrency` field

**File:** `internal/repository/user.go`
- Remove `default_currency` from all SELECT, INSERT, UPDATE queries
- Remove `"default_currency": u.DefaultCurrency` from all named args

### 4l. User Service

**File:** `internal/service/user.go`
- Remove `DefaultCurrency` update logic

### 4m. User Sync Middleware

**File:** `internal/middleware/user_sync.go`
- Remove `DefaultCurrency: "EGP"` from user creation

### 4n. Plan Guard

**File:** `internal/service/plan_guard.go`
- Remove `HasMultiCurrency` field from `PlanLimits` struct
- Remove `HasMultiCurrency` from all plan maps

---

## Step 5: Add Billing Fields to Transaction Model

**File:** `internal/model/transaction/transaction.go`

Add fields:
```go
BillingProvider string `json:"billing_provider,omitempty" db:"billing_provider"`
BillingRefID    string `json:"billing_ref_id,omitempty" db:"billing_ref_id"`
BillingStatus   string `json:"billing_status" db:"billing_status"` // "manual", "pending", "paid", "failed"
```

**File:** `internal/model/transaction/transaction.dto.go`

- Add `BillingStatus *string` filter to `ListTransactionsRequest`

---

## Step 6: Repository Updates for Billing

**File:** `internal/repository/transaction.go`

Add methods:
- `CreateFromBilling(ctx context.Context, tx *transaction.Transaction) error` — create transaction from billing event
- `GetByBillingRef(ctx context.Context, refID string) (*transaction.Transaction, error)` — lookup by billing reference

Update existing queries to include the new billing columns (`billing_provider`, `billing_ref_id`, `billing_status`).

---

## Step 7: Subscription DTOs

**New file:** `internal/model/subscription/subscription.dto.go`

```go
type CheckoutRequest struct {
    Plan       string `json:"plan" binding:"required,oneof=pro vip"`
    SuccessURL string `json:"success_url" binding:"required"`
    CancelURL  string `json:"cancel_url" binding:"required"`
}

type CheckoutResponse struct {
    CheckoutURL string `json:"checkout_url"`
    RefNumber   string `json:"ref_number"`
}

type UpgradePlanRequest struct {
    Plan string `json:"plan" binding:"required,oneof=pro vip"`
}

func (r CheckoutRequest) Validate() error { ... }
func (r UpgradePlanRequest) Validate() error { ... }
```

---

## Step 8: Subscription Service

**New file:** `internal/service/subscription.go`

```go
type SubscriptionService struct {
    server           *server.Server
    subscriptionRepo *repository.SubscriptionRepo
    billingClient    *billing.Client
}

func NewSubscriptionService(s *server.Server, subRepo *repository.SubscriptionRepo, billingClient *billing.Client) *SubscriptionService
```

Methods:
- `GetSubscription(ctx, userID) (*subscription.Subscription, error)` — ensure free sub exists, return it
- `CreateCheckout(ctx, userID, req) (*subscription.CheckoutResponse, error)` — generate merchantRefNum, call Fawry CreateCharge, save `pending` subscription with `expires_at` set from `ChargeExpiryMinutes`, return checkout URL
- `HandleWebhook(ctx, event *billing.WebhookEvent) error` — update subscription status based on payment event, auto-create transaction if PAID (idempotency is enforced one layer up, in `BillingService.HandleWebhook` — see Step 9)
- `UpgradePlan(ctx, userID, req) (*subscription.CheckoutResponse, error)` — create new charge for plan upgrade
- `CancelSubscription(ctx, userID) error` — end-of-period cancel: sets `cancel_at_period_end = true` and `cancelled_at = now()`; does **not** touch `status` or entitlements — those only change when `current_period_end` actually passes (handled by the recurring billing worker, which skips creating a renewal charge when `cancel_at_period_end` is true and instead flips status to `cancelled`)
- `Entitlements(ctx, userID) (*subscription.Entitlements, error)` — derives feature access purely from `status` (`active`/`past_due` during grace period → entitled; `cancelled`/`expired`/`incomplete` → not). This is the only place entitlement logic lives — handlers and other services call this rather than checking `status` themselves

---

## Step 9: Billing Service

**New file:** `internal/service/billing.go`

```go
type BillingService struct {
    server           *server.Server
    billingClient    *billing.Client
    transactionRepo  *repository.TransactionRepo
    subscriptionRepo *repository.SubscriptionRepo
}

func NewBillingService(s *server.Server, billingClient *billing.Client, txRepo *repository.TransactionRepo, subRepo *repository.SubscriptionRepo) *BillingService
```

Methods:
- `CreateCheckout(ctx, userID, plan, successURL, cancelURL) (*CheckoutResponse, error)` — generates merchantRefNum, calls Fawry, stores pending subscription
- `HandleWebhook(ctx, event *billing.WebhookEvent) error` — the only entry point that processes Fawry notifications. Order of operations, in this order, every time:
  1. Verify signature (`VerifyWebhookSignature`) — reject silently (log + 200 OK, no processing) on mismatch
  2. `billingEventRepo.RecordEvent(ctx, event)` — insert into `billing_events` keyed on `fawry_ref_number`; if it already existed (duplicate delivery), stop here and return — **do not reprocess**
  3. Map `event.OrderStatus` to a subscription transition: `PAID` → active (extend `current_period_end`, reset `failed_attempts`), `FAILED`/`CANCELED` on a renewal → `past_due` + increment `failed_attempts` (dunning, not immediate cancel), `EXPIRED` → `expired`, `REFUNDED` → handle as its own case, not folded into `FAILED`
  4. On `PAID`, auto-create the transaction (`CreateFromBilling`) — guarded by the same idempotency check, so a duplicate notification can never double-create
- `GetPaymentStatus(ctx, refNumber) (*billing.ChargeResponse, error)` — query Fawry for current status; used by the reconciliation job, not the webhook path
- `GetBillingHistory(ctx, userID, pagination) (*PaginatedResponse, error)` — list transactions filtered by billing_status

---

## Step 10: Billing Handler

**New file:** `internal/handler/billing.go`

```go
type BillingHandler struct {
    Handler
    billingService *service.BillingService
}

func NewBillingHandler(s *server.Server, billingService *service.BillingService) *BillingHandler
```

Endpoints:
- `POST /api/v1/billing/checkout` — create checkout session → returns Fawry charge URL
- `POST /api/v1/billing/webhook` — receive Fawry webhook (no auth middleware)
- `GET /api/v1/billing/status/:ref_number` — check payment status
- `GET /api/v1/billing/history` — user's billing history (paginated, filterable)

---

## Step 11: Refactor Subscription Handler

**File:** `internal/handler/subscription.go`

- Replace direct `repository.EnsureFreeSubscription` call with `SubscriptionService`
- Add `SubscriptionService` dependency
- Update `Get` to call `subscriptionService.GetSubscription`
- Add `Upgrade` endpoint

---

## Step 12: Wire Everything Together

### `internal/handler/handlers.go`

Add fields:
```go
Billing      *BillingHandler
```

Update `NewHandlers` to accept `*service.BillingService` and `*service.SubscriptionService`.

### `internal/service/services.go`

Add fields:
```go
BillingService      *BillingService
SubscriptionService *SubscriptionService
```

In `NewServices`:
1. Create `billing.Client` from config
2. Create `BillingService` with billing client + repos
3. Create `SubscriptionService` with sub repo + billing client

### `internal/router/v1/v1.go`

Add: `registerBillingRoutes(r, h.Billing)`

### `internal/router/v1/billing.go` (NEW)

```go
func registerBillingRoutes(r *gin.RouterGroup, h *handler.BillingHandler) {
    billing := r.Group("/billing")
    {
        billing.POST("/checkout", h.CreateCheckout)
        billing.POST("/webhook", h.Webhook)  // no auth — called by Fawry
        billing.GET("/status/:ref_number", h.GetStatus)
        billing.GET("/history", h.GetHistory)
    }
}
```

### `internal/router/v1/subscription.go`

Add upgrade and cancel routes:
```go
subscription.POST("/upgrade", h.Upgrade)
subscription.POST("/cancel", h.Cancel)  // end-of-period cancel — see SubscriptionService.CancelSubscription
```

---

## Step 13: Skip Middleware for Webhook

**File:** `internal/router/router.go`

The `/api/v1/billing/webhook` route must bypass auth, rate limit, CORS, and user sync middleware. Register it before the auth middleware group, or create a separate route group outside the protected middleware stack.

---

## Step 14: Billing Events Repository (Idempotency Ledger)

**New file:** `internal/repository/billing_event.go`

```go
type BillingEventRepo struct {
    db *sqlx.DB
}

func NewBillingEventRepo(db *sqlx.DB) *BillingEventRepo
```

Methods:
- `RecordEvent(ctx context.Context, e *billing.WebhookEvent) (created bool, err error)` — `INSERT ... ON CONFLICT (fawry_ref_number) DO NOTHING`, returns `created = false` when the row already existed. This return value is the idempotency gate `BillingService.HandleWebhook` checks before doing anything else.
- `GetByMerchantRef(ctx context.Context, merchantRefNum string) (*BillingEvent, error)` — lookup for support/debugging and for the recurring worker to check whether a given cycle's charge was already attempted.

This table is append-only. Nothing in the codebase should `UPDATE` or `DELETE` rows here — it is the audit trail for disputes and reconciliation.

---

## Step 15: Recurring Billing Worker

**New file:** `internal/worker/billing_recurring.go`

Fawry has no subscription object, so renewal charges are created here, not by Fawry. Runs on a schedule (e.g. every 15 min) via the app's existing scheduler.

```go
type RecurringBillingWorker struct {
    subscriptionRepo *repository.SubscriptionRepo
    billingClient    *billing.Client
    billingEventRepo *repository.BillingEventRepo
    cfg              *config.BillingConfig
    logger           *zerolog.Logger
}

func (w *RecurringBillingWorker) Run(ctx context.Context) error
```

Per run:
1. Query subscriptions where `status = 'active' AND next_billing_at <= now()`.
2. For each: if `cancel_at_period_end` is true, flip `status = 'cancelled'`, do **not** charge — this is the actual mechanism by which cancellation takes effect.
3. Otherwise, create a new charge via `billingClient.CreateCharge` using the stored `customerProfileId` (card-on-file token) — no user interaction required. Use a fresh `merchantRefNum` per attempt (never reuse one across cycles or retries).
4. Set subscription to `pending` for this cycle; wait for the webhook to resolve it (per Step 9's `HandleWebhook`) — do **not** mark the cycle paid synchronously here.
5. On the *next* run, if a `pending` renewal charge's `expires_at` has passed with no resolution, treat as failed and hand off to dunning (increment `failed_attempts`, set `past_due`) rather than leaving it pending forever.

---

## Step 16: Reconciliation Job

**New file:** `internal/worker/billing_reconciliation.go`

Safety net for notifications that never arrive (dropped delivery, user abandoned a Pay-at-Fawry kiosk code, etc.). Runs on a schedule (e.g. every 15–30 min).

```go
type ReconciliationWorker struct {
    subscriptionRepo *repository.SubscriptionRepo
    billingClient    *billing.Client
    cfg              *config.BillingConfig
    logger           *zerolog.Logger
}

func (w *ReconciliationWorker) Run(ctx context.Context) error
```

Per run:
1. **Expire stale pending charges** — `status = 'pending' AND expires_at < now()` → mark `expired`. This is what prevents a Pay-at-Fawry reference code or an abandoned 3DS flow from sitting unresolved forever.
2. **Resolve stuck past-due/pending renewals** — `status IN ('pending','past_due') AND current_period_end < now() - dunning_grace_days` → call `billingClient.GetPaymentStatus(refNum)` directly (don't just guess), and resolve to `active` or `cancelled` based on Fawry's actual answer, not the passage of time alone.
3. **Bounded dunning retries** — for `past_due` subscriptions with `failed_attempts < DunningMaxRetries` and enough time elapsed since the last attempt (`DunningRetryIntervalHr`), trigger another charge attempt via the same path as Step 15. Past `DunningMaxRetries` or past `DunningGraceDays`, downgrade to `cancelled` and stop retrying.

This job uses the partial index on `subscriptions(status)` from Step 3, so it only ever scans non-terminal rows.

---

## Execution Order

| # | Task | Files |
|---|------|-------|
| 1 | Add billing config (incl. charge expiry + dunning policy) | `config.go` |
| 2 | Create billing lib (client, types, plans, `GetPaymentStatus`) | `internal/lib/billing/client.go`, `types.go`, `plans.go` |
| 3 | New migration — billing_events table + subscription pending/cancel/dunning columns | `010_add_billing_tables.sql` |
| 4 | Remove multi-currency from transactions | model, dto, repo, service |
| 5 | Remove multi-currency from budgets | model, dto, repo, service |
| 6 | Remove multi-currency from recurring rules | model, dto, repo, service |
| 7 | Remove multi-currency from imports | service, handler |
| 8 | Remove multi-currency from users | model, repo, service, middleware |
| 9 | Remove HasMultiCurrency from plan guard | `plan_guard.go` |
| 10 | Add billing fields to transaction model | `transaction.go`, `transaction.dto.go` |
| 11 | Update transaction repository for billing | `transaction.go` repo |
| 12 | Create billing_events repository (idempotency ledger) | `billing_event.go` repo |
| 13 | Create subscription DTOs | `subscription.dto.go` |
| 14 | Create subscription service (state machine + `CancelSubscription` + `Entitlements`) | `subscription.go` service |
| 15 | Create billing service (checkout, webhook w/ idempotency, status, history) | `billing.go` service |
| 16 | Create recurring billing worker | `billing_recurring.go` worker |
| 17 | Create reconciliation job | `billing_reconciliation.go` worker |
| 18 | Create billing handler | `billing.go` handler |
| 19 | Refactor subscription handler + add cancel endpoint | `subscription.go` handler |
| 20 | Wire services, workers + handlers | `handlers.go`, `services.go` |
| 21 | Add billing routes | `v1.go`, `billing.go` router |
| 22 | Update subscription routes (upgrade + cancel) | `subscription.go` router |
| 23 | Skip middleware for webhook | `router.go` |
| 24 | Verify build | `go build ./...` |

---

## Files Summary

| File | Action |
|------|--------|
| `internal/config/config.go` | Modify — add `BillingConfig` |
| `internal/lib/billing/client.go` | **New** — Fawry API client |
| `internal/lib/billing/types.go` | **New** — Fawry request/response types |
| `internal/lib/billing/plans.go` | **New** — plan pricing helper |
| `internal/database/migrations/010_add_billing_tables.sql` | **New** — `billing_events` table + subscription pending/cancel/dunning columns |
| `internal/repository/billing_event.go` | **New** — idempotency ledger repo (`RecordEvent`, `GetByMerchantRef`) |
| `internal/model/transaction/transaction.go` | Modify — remove Currency, add billing fields |
| `internal/model/transaction/transaction.dto.go` | Modify — remove Currency, CurrencyTotal, add billing filter |
| `internal/repository/transaction.go` | Modify — remove currency queries, add billing methods |
| `internal/service/transaction.go` | Modify — remove currency logic |
| `internal/model/budget/budget.go` | Modify — remove Currency |
| `internal/model/budget/budget.dto.go` | Modify — remove Currency |
| `internal/repository/budget.go` | Modify — remove currency queries |
| `internal/service/budget.go` | Modify — remove currency logic |
| `internal/model/recurring/recurring.go` | Modify — remove Currency |
| `internal/model/recurring/recurring.dto.go` | Modify — remove Currency |
| `internal/repository/recurring.go` | Modify — remove currency queries |
| `internal/service/recurring.go` | Modify — remove currency logic |
| `internal/service/import.go` | Modify — remove Currency from CSVRow |
| `internal/handler/import.go` | Modify — remove Currency from CSV parsing |
| `internal/model/user/user.go` | Modify — remove DefaultCurrency |
| `internal/repository/user.go` | Modify — remove default_currency |
| `internal/service/user.go` | Modify — remove DefaultCurrency logic |
| `internal/middleware/user_sync.go` | Modify — remove DefaultCurrency |
| `internal/service/plan_guard.go` | Modify — remove HasMultiCurrency |
| `internal/model/subscription/subscription.dto.go` | **New** — checkout/upgrade DTOs |
| `internal/service/subscription.go` | **New** — subscription service |
| `internal/service/billing.go` | **New** — billing service (checkout, webhook w/ idempotency, status, history) |
| `internal/worker/billing_recurring.go` | **New** — recurring billing worker (charges saved card on renewal, honors `cancel_at_period_end`) |
| `internal/worker/billing_reconciliation.go` | **New** — reconciliation job (expires stale pending, resolves stuck past-due, drives dunning retries) |
| `internal/handler/billing.go` | **New** — billing handler |
| `internal/handler/handlers.go` | Modify — add Billing handler |
| `internal/handler/subscription.go` | Modify — use SubscriptionService, add Cancel endpoint |
| `internal/service/services.go` | Modify — add BillingService, SubscriptionService, register workers |
| `internal/router/v1/v1.go` | Modify — add billing routes |
| `internal/router/v1/billing.go` | **New** — billing route registration |
| `internal/router/v1/subscription.go` | Modify — add upgrade + cancel routes |
| `internal/router/router.go` | Modify — skip middleware for webhook |