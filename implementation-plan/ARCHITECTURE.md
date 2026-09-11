# Expense Tracker — Architecture & System Design

**Nx monorepo** Architecture

---

## 1. Purpose

A production-grade personal and shared finance tracker, built as a vehicle to practice real backend engineering discipline:

* Clean architecture and dependency inversion
* Domain-driven service boundaries
* Relational data modeling and integrity
* Async job processing and idempotency
* Multi-user authorization
* Shared/family financial resources
* Observability and distributed tracing
* API contract design
* Object storage and presigned uploads
* Subscription and plan enforcement
* Integration testing against real infrastructure

The project builds on infrastructure that is already wired up so the primary focus remains on **domain logic, reliability, scalability, and production engineering practices rather than basic infrastructure plumbing**.

---

## 2. Requirements

### Functional Requirements

| Area                   | Requirement                                                                                            |
| ---------------------- | ------------------------------------------------------------------------------------------------------ |
| Categories              | CRUD; type = income/expense; cannot delete if transactions reference it                                |
| Transactions            | CRUD; filter by date range, category, type; paginated list; multi-currency amounts                     |
| Budgets                 | Monthly limit per category; percentage used; exceeded state; shared/family budgets with member roles   |
| Summary                 | Totals by category and month; month-over-month trend; per-currency grouping                            |
| Recurring transactions  | Rule-based auto-generation based on frequency and next-run date; materializes real transactions        |
| Reports                 | Generate PDF and/or CSV; on-demand and monthly scheduled; downloadable via expiring link                |
| CSV import               | Bulk transaction upload; per-row validation; partial success with per-row error reporting              |
| Receipts                 | Attach a photo to a transaction; direct-to-bucket upload                                               |
| Notifications            | Monthly summary email; budget-exceeded alert email                                                     |
| Auth                     | Clerk-based signup/login; row-level ownership; shared-budget membership roles                          |
| Pricing / Plans          | Three tiers — **Free**, **Pro**, **VIP**; enforced at service layer; external provider handles billing |

### Non-Functional Requirements

**Reliability** — Background jobs must be safe under Asynq's at-least-once delivery model. Jobs must be idempotent so retries cannot create duplicate transactions, reports, imports, notifications, or billing effects.

**Security** — No cross-user data leakage; every resource access is authorized; private object storage bucket; presigned URLs for uploads/downloads; least-privilege storage credentials; provider webhook signatures verified; no sensitive provider credentials exposed to the frontend.

**Observability** — Every HTTP request and background job must be traceable through correlation/request/job IDs. The system exposes request rate, error rate, request latency, job execution metrics, queue depth, job failures/retries, and database performance metrics.

**Performance** — Aggregation performed in SQL rather than application memory; pagination on all collection endpoints; appropriate database indexes; background processing for expensive operations; direct-to-object-storage uploads; no large CSV/report processing inside HTTP request handlers.

**Data Integrity** — Money stored as integer minor units; currency explicitly stored with monetary values; database foreign keys and constraints act as a backstop; unique constraints prevent duplicate logical records; database transactions used where multiple writes must succeed atomically.

---

## Pricing Tiers & Feature Limits

| Feature                     | Free       | Pro               | VIP                  |
| --------------------------- | ---------- | ----------------- | -------------------- |
| **Price**                   | $0 / month | $9 / month        | $29 / month          |
| **Categories**              | 5 max      | 50 max            | Unlimited            |
| **Transactions / month**    | 50 max     | 500 max           | Unlimited            |
| **Budgets**                 | 1 max      | 10 max            | Unlimited            |
| **Recurring rules**         | ✗          | ✓                 | ✓                    |
| **Shared / family budgets** | ✗          | ✓ up to 3 members | ✓ unlimited members  |
| **CSV import**              | ✗          | ✓ 100 rows/upload | ✓ 10,000 rows/upload |
| **Receipt attachments**     | ✗          | ✓                 | ✓                    |
| **PDF + CSV reports**       | CSV only   | PDF + CSV         | PDF + CSV            |
| **Report history**          | 3 reports  | 30 reports        | Unlimited            |
| **Multi-currency support**  | ✗          | ✓                 | ✓                    |
| **Priority email support**  | ✗          | ✗                 | ✓                    |
| **Full account export**     | ✗          | ✗                 | ✓                    |

Limits are enforced **in the service layer**, not only in the frontend, so limits remain enforced regardless of which client accesses the API.

When a user exceeds a plan limit:

```text
HTTP 402 Payment Required
code: plan_limit_exceeded
```

---

## 3. Domain Models

### UserAccount

Clerk remains the identity provider. The application maintains a local user record for application-owned data and relationships.

```text
USER_ACCOUNT
----------------------------
id                TEXT PK
email             TEXT UNIQUE
display_name      TEXT
avatar_url        TEXT NULL
default_currency  CHAR(3)
timezone          TEXT
created_at        TIMESTAMPTZ
updated_at        TIMESTAMPTZ
```

`id` is the Clerk User ID.

### Category

```text
CATEGORY
----------------------------
id          UUID PK
user_id     TEXT FK -> user_account.id
name        TEXT
type        TEXT
created_at  TIMESTAMPTZ

UNIQUE(user_id, name)
```

Valid types: `income`, `expense`. A category cannot be deleted while transactions reference it.

### Transaction

```text
TRANSACTION
----------------------------
id            UUID PK
user_id       TEXT FK -> user_account.id
category_id   UUID FK -> category.id
amount_minor  BIGINT
currency      CHAR(3)
note          TEXT
receipt_key   TEXT NULL
occurred_at   TIMESTAMPTZ
created_at    TIMESTAMPTZ
```

Money is represented using minor currency units plus the currency code (`amount_minor + currency`). Examples: EGP 150.50 → `amount_minor = 15050, currency = EGP`; USD 25.99 → `amount_minor = 2599, currency = USD`. Never use floating-point values for financial amounts.

### Budget

```text
BUDGET
----------------------------
id                  UUID PK
category_id         UUID FK -> category.id
monthly_limit_minor BIGINT
currency            CHAR(3)
created_at          TIMESTAMPTZ
```

The budget does not contain an `owner_user_id`; ownership and membership are represented through `BUDGET_MEMBER`.

### BudgetMember

```text
BUDGET_MEMBER
----------------------------
budget_id  UUID FK -> budget.id
user_id    TEXT FK -> user_account.id
role       TEXT

PRIMARY KEY (budget_id, user_id)
```

Valid roles: `owner`, `member`. Database-level enforcement should ensure a budget has exactly one owner:

```sql
CREATE UNIQUE INDEX one_owner_per_budget
ON budget_members (budget_id)
WHERE role = 'owner';
```

Service layer rules: only the owner can add/remove members; a member cannot promote themselves to owner; the owner cannot be removed without transferring ownership; only members can access a shared budget; budget limits are checked against the budget owner's subscription plan.

### RecurringRule

```text
RECURRING_RULE
----------------------------
id                  UUID PK
user_id             TEXT FK -> user_account.id
category_id         UUID FK -> category.id
amount_minor        BIGINT
currency            CHAR(3)
frequency           TEXT
next_run_date       DATE
last_generated_date DATE NULL
end_date            DATE NULL
```

Valid frequencies: `weekly`, `monthly`. Recurring rules create real `TRANSACTION` records; generation must be idempotent, using `recurring_rule_id + occurrence_date` as a deterministic deduplication key.

### Report

```text
REPORT
----------------------------
id            UUID PK
user_id       TEXT FK -> user_account.id
format        TEXT
period_start  DATE
period_end    DATE
status        TEXT
storage_key   TEXT NULL
created_at    TIMESTAMPTZ
completed_at  TIMESTAMPTZ NULL
```

Valid formats: `pdf`, `csv`. Valid statuses: `pending`, `processing`, `ready`, `failed`. Reports are generated asynchronously; the API returns `202 Accepted`, and once ready it generates an expiring presigned download URL.

### Import

```text
IMPORT
----------------------------
id           UUID PK
user_id      TEXT FK -> user_account.id
status       TEXT
total_rows   INT
success_rows INT
failed_rows  JSONB
created_at   TIMESTAMPTZ
```

Valid statuses: `pending`, `processing`, `completed`, `failed`. Example `failed_rows`:

```json
[
  { "row": 12, "reason": "invalid date" },
  { "row": 18, "reason": "unknown category" }
]
```

CSV processing happens asynchronously; a large CSV must never be processed entirely inside the HTTP request lifecycle.

### Subscription

```text
SUBSCRIPTION
----------------------------
id                        UUID PK
user_id                   TEXT UNIQUE FK -> user_account.id
plan                      TEXT
status                    TEXT
payment_provider          TEXT NULL
provider_customer_id      TEXT NULL
provider_subscription_id  TEXT NULL
current_period_end        TIMESTAMPTZ NULL
created_at                TIMESTAMPTZ
updated_at                TIMESTAMPTZ
```

Valid plans: `free`, `pro`, `vip`. Valid statuses: `active`, `past_due`, `canceled`. Supported providers can include `fawry`. The application should hide provider-specific implementation behind a billing interface.

### Billing Provider Abstraction

```text
BillingProvider
├── CreateCheckoutSession()
├── CreateCustomerPortal()
├── VerifyWebhook()
└── ParseWebhookEvent()
```

Provider-specific implementations live in infrastructure/platform code, so the billing provider can be replaced without changing domain services.

### Database Indexes

```text
transactions(user_id, occurred_at)
transactions(user_id, category_id)
transactions(user_id, occurred_at, category_id)
categories(user_id)
budgets(category_id)
budget_members(user_id)
budget_members(budget_id)
recurring_rules(user_id, next_run_date)
reports(user_id, created_at)
imports(user_id, created_at)
subscriptions(user_id) UNIQUE
```

Additional indexes should be introduced based on actual query patterns rather than indexing every column automatically.

### Money Representation

Financial amounts use `BIGINT amount_minor` + `CHAR(3) currency`. "Minor unit" is used instead of "cents" because subdivisions differ by currency (USD/EUR → cents, EGP → piastres, JPY → no fractional minor unit). The database stores the integer amount together with its currency rather than assuming every currency uses cents.

---

## 4. API Design & Contract

**Versioning** — path-based: `/v1` (e.g. `GET /v1/transactions`).

**Error Envelope**

```json
{
  "error": {
    "code": "validation_failed",
    "message": "amount_cents must be non-zero",
    "field": "amount_cents"
  }
}
```

Fixed error vocabulary: `validation_failed`, `not_found`, `forbidden`, `conflict`, `plan_limit_exceeded`, `rate_limited`, `internal_error`.

**HTTP Status Codes**

| Status | Meaning                                    |
| ------ | ------------------------------------------ |
| 400    | Malformed request                          |
| 401    | Unauthenticated                            |
| 402    | Subscription/plan limit exceeded           |
| 403    | Authenticated but forbidden                |
| 404    | Resource not found or intentionally hidden |
| 409    | Resource conflict                          |
| 422    | Request validation failed                  |
| 429    | Rate limited                               |
| 500    | Internal server error                      |

For protected resources, the service may return `404` instead of `403` when revealing the existence of another user's resource would itself leak information.

**Pagination**

```json
{
  "data": [],
  "pagination": { "page": 1, "limit": 20, "total": 143 }
}
```

Default `page = 1`, `limit = 20`. The maximum allowed page size is enforced server-side.

**Endpoints**

```text
# Categories
GET     /v1/categories
POST    /v1/categories
PATCH   /v1/categories/{id}
DELETE  /v1/categories/{id}

# Transactions
GET     /v1/transactions?category_id=&from=&to=&type=&page=&limit=
POST    /v1/transactions
PATCH   /v1/transactions/{id}
DELETE  /v1/transactions/{id}
POST    /v1/transactions/{id}/receipt        # returns presigned upload URL
GET     /v1/transactions/summary?month=YYYY-MM

# Budgets
GET     /v1/budgets
POST    /v1/budgets
PATCH   /v1/budgets/{id}
POST    /v1/budgets/{id}/members
DELETE  /v1/budgets/{id}/members/{userId}

# Recurring Rules
POST    /v1/recurring-rules
GET     /v1/recurring-rules
PATCH   /v1/recurring-rules/{id}
DELETE  /v1/recurring-rules/{id}

# Reports
POST    /v1/reports          # 202 Accepted
GET     /v1/reports/{id}
GET     /v1/reports

# CSV Imports
POST    /v1/imports          # 202 Accepted
GET     /v1/imports/{id}

# Subscription / Billing
GET     /v1/subscription
POST    /v1/subscription/checkout    # returns provider Checkout Session
POST    /v1/subscription/portal
POST    /v1/webhooks/provider
```

The provider webhook endpoint: verifies the provider signature → extracts the provider event ID → checks whether the event was already processed → stores the event → updates the subscription → acknowledges duplicate events safely.

---

## 5. Tools & Purpose

| Tool               | Purpose                                                                                 |
| ------------------- | ---------------------------------------------------------------------------------------- |
| Nx + Bun            | Monorepo task orchestration, dependency graph, builds, tests, and development workflows |
| Go                  | Backend language                                                                        |
| Gin                 | HTTP framework — routing, middleware chain, request binding                             |
| Clean Architecture  | Separation of handlers, services, repositories, and domain models                       |
| Postgres            | Primary relational datastore                                                            |
| Clerk               | Authentication, signup/login, JWT verification, React SDK                               |
| Redis               | Caching, distributed coordination, and Asynq backend                                    |
| Asynq               | Background job queue                                                                    |
| Cloudflare R2       | Object storage for reports and receipt photos                                           |
| Resend              | Transactional email                                                                     |
| Billing Provider    | Subscription checkout, customer portal, and billing webhooks                            |
| New Relic           | Logs, metrics, distributed traces, and application monitoring                           |
| Testcontainers      | Integration/API tests using real Postgres and Redis                                     |
| OpenAPI/Swagger     | API contract and interactive documentation                                              |
| React               | Frontend application                                                                    |
| Zod                 | Frontend/runtime schema validation where appropriate                                    |

---

## 6. Background Jobs

Asynchronous work is handled through Asynq. Recommended queues: `critical`, `default`, `low`.

* **Critical** — subscription synchronization, important notifications
* **Default** — CSV imports, report generation, recurring transaction generation, budget alerts
* **Low** — monthly summaries, analytics, cleanup jobs

**Job Idempotency** — every job must be safe to retry:

* Recurring Transactions: dedup key = `recurring_rule_id + occurrence_date`
* Reports: a report ID represents one generation request; retries update the same report instead of creating another
* Imports: each import has a unique import ID; processing resumes from its current state instead of starting another independent import
* Notifications: dedup key = `user_id + notification_type + period`

---

## 7. Architecture

```text
HTTP Handler
     │
     ▼
Service / Use Case
     │
     ├──────────────► Repository
     ├──────────────► Storage
     ├──────────────► Queue
     ├──────────────► Email
     └──────────────► Billing
```

The service layer owns business rules. Repositories handle persistence. Handlers translate HTTP requests into service calls and service results into HTTP responses. Infrastructure implementations remain replaceable.

**Dependency Direction** — dependencies point inward:

```text
Handler → Service → Interfaces ← Infrastructure implementations
```

Example: `Service` depends on a `Storage` interface; `R2Storage` implements `Storage`. The service must not directly import Cloudflare R2 SDK code.

---

## 8. Best Practices Applied

* **Dependency Inversion** — repository interfaces are defined by the service/domain layer rather than dictated by infrastructure implementations.
* **Three Struct Separation** — separate `Model`, `Request DTO`, `Response DTO`; never reuse a database model directly as an HTTP request or response.
* **Context-Carried Identity** — the authenticated user ID is extracted at the handler/middleware boundary and passed through `context.Context`.
* **Row-Level Authorization** — every user-owned query scopes by the authenticated user (`WHERE id = $1 AND user_id = $2`); shared budgets use membership authorization via `BUDGET → BUDGET_MEMBER → authenticated user`.
* **Money Representation** — `BIGINT amount_minor` + `CHAR(3) currency`; never `FLOAT`/`DOUBLE` for financial amounts.
* **Idempotent Background Jobs** — unique constraints, status columns, checkpoints, deterministic dedup keys, database transactions.
* **Queue Separation** — `critical` / `default` / `low`.
* **Storage Abstraction** — `Storage` interface (`GenerateUploadURL`, `GenerateDownloadURL`, `Delete`, `Exists`); implementations can include R2, S3, MinIO without changing domain services.
* **Structured Logging** — JSON logs with `timestamp`, `level`, `service`, `environment`, `request_id`, `correlation_id`, `job_id`, `user_id`, `operation`, `duration`, `error`.
* **RED Metrics** — HTTP: Rate, Errors, Duration per route. Jobs: `jobs_started_total`, `jobs_completed_total`, `jobs_failed_total`, `job_duration_seconds`, `job_retries_total`.
* **Fail-Fast Configuration** — required configuration validated at startup; the app never silently starts with missing critical configuration.
* **Sentinel Errors** — `ErrNotFound`, `ErrForbidden`, `ErrConflict`, `ErrPlanLimitExceeded`, `ErrValidation` via `errors.Is`/`errors.As`; handlers map domain errors to HTTP responses without leaking internals.
* **Consistent API Contracts** — `/v1`, consistent error envelope, consistent pagination, consistent HTTP status semantics, OpenAPI documentation.
* **Direct-to-Bucket Uploads** — receipts upload directly to object storage via presigned URL; the API never proxies file contents.
* **Plan Enforcement** — every operation subject to a subscription limit calls `PlanGuard` before performing the protected operation (allowed → `Repository.Create()`; exceeded → `402 plan_limit_exceeded`).
* **Provider Webhook Idempotency** — each billing webhook has a unique `provider_event_id`, stored before/while processing; retried events are acknowledged without repeating side effects.

---

## 9. High-Level System Architecture

```text
                         ┌─────────────────┐
                         │     React       │
                         │    Frontend     │
                         └────────┬────────┘
                                  │
                              HTTPS / JSON
                                  │
                                  ▼
                         ┌─────────────────┐
                         │      API        │
                         │   Handlers      │
                         └────────┬────────┘
                                  │
                                  ▼
                         ┌─────────────────┐
                         │    Services     │
                         │  Business Logic │
                         └────┬───┬───┬────┘
                              │   │   │
                ┌─────────────┘   │   └─────────────┐
                ▼                 ▼                 ▼
          ┌───────────┐     ┌───────────┐     ┌───────────┐
          │ Postgres  │     │   Redis   │     │    R2     │
          │           │     │  + Asynq  │     │  Storage  │
          └───────────┘     └─────┬─────┘     └───────────┘
                                  │
                                  ▼
                           ┌─────────────┐
                           │   Workers   │
                           │             │
                           │ Reports     │
                           │ Imports     │
                           │ Recurring   │
                           │ Email       │
                           │ Billing     │
                           └──────┬──────┘
                                  │
                    ┌─────────────┼──────────────┐
                    ▼             ▼              ▼
                ┌───────┐    ┌──────────┐   ┌──────────┐
                │Resend │    │ Billing  │   │New Relic │
                │ Email │    │ Provider │   │Observ.   │
                └───────┘    └──────────┘   └──────────┘
```

---

## 10. Core Architectural Principles

1. Business logic lives in services, not HTTP handlers.
2. Infrastructure dependencies are hidden behind interfaces.
3. Every user-owned resource is authorized server-side.
4. Shared resources use explicit membership authorization.
5. Money is represented using integer minor units plus currency.
6. Long-running operations are asynchronous.
7. Every background job is retry-safe and idempotent.
8. Database constraints provide a final integrity boundary.
9. Subscription limits are enforced independently of the frontend.
10. External providers are accessed through abstractions.
11. Object storage is private and accessed through presigned URLs.
12. Observability is part of the application architecture, not an afterthought.
13. API behavior is explicitly documented through OpenAPI.
14. Integration tests use real infrastructure through Testcontainers.
15. The Nx monorepo provides a unified development and CI/CD workflow.

---

## 11. MoneyFlow Backend Target — What to Build From Scratch (Gin)

The backend is being built from scratch on **Gin**. This section is the target checklist for that foundation layer — everything domain-specific (categories, transactions, budgets, plan enforcement, etc.) gets built on top of it once it's in place, so it's worth getting this layer right before touching domain code.

### Architecture

* Infrastructure dependencies are isolated behind `server` / `config` / `database` packages.
* Generic handler helpers centralize: request binding, validation, JSON responses, error handling, logging, and New Relic tracing.

### Configuration

* Environment-based configuration using `MONEYFLOW_*` variables.
* Configuration validation at startup; required values fail fast instead of silently falling back to invalid defaults.
* Separate configuration blocks for server, database, Redis, authentication, integrations, and observability.

### Database

* PostgreSQL connection pooling.
* Database ping during startup.
* Configurable pool limits and connection lifetimes.
* Password URL-encoding for safer DSN construction.
* Embedded, versioned migrations.
* Database query tracing with New Relic.
* SQL logging available in local development.

### Security

* Clerk JWT/session authentication.
* Authenticated user identity stored in request context.
* CORS configuration.
* Secure HTTP headers via Gin middleware (e.g. `gin-contrib/secure`).
* Global panic recovery.
* Rate limiting.
* User ID and permissions extracted from Clerk claims.

### Observability

* Structured logging, with environment and service fields included.
* Request IDs; user IDs attached to request logs.
* Request method, URI, status, latency, IP, and user agent logging.
* New Relic APM integration and distributed tracing.
* PostgreSQL instrumentation.
* Error stack traces in development; application log forwarding in production.

### Error Handling

* Centralized global error handler.
* Consistent HTTP error conversion.
* Database errors translated into application errors.
* Internal errors logged with stack traces; sensitive internal error details are not returned directly to clients.
* Support for validation errors and field-level errors.

### Background Jobs

* Redis-backed Asynq job processing.
* Separate priority queues: `critical`, `default`, `low`.
* Configurable worker concurrency.
* Graceful worker shutdown.
* Job client and worker server abstraction.

### Testing

* Testcontainers-based PostgreSQL integration tests.
* Migrations applied automatically during test setup.
* Test database cleanup through test lifecycle hooks.
* Database connection retry logic.
* Helpers for transactions, assertions, servers, and test setup.

### API Design

* Versioned API route structure beginning with `/api/v1`.
* OpenAPI specification and Swagger HTML documentation.
* Standardized response handlers.
* Request validation before business logic execution.
* Health/system routes.

### Email Integration

* Resend integration.
* HTML email templates.
* Welcome email background task.
* Email preview support.

---

## 12. Gin Foundation — Build Order

Building the MoneyFlow foundation from scratch means deciding, in order, what gets wired up before any domain code (categories, transactions, etc.) is written. Suggested build order for the Gin foundation:

1. **`go.mod` + project layout** — clean-architecture folders: `cmd/api` (entrypoint), `internal/handler`, `internal/service`, `internal/repository`, `internal/domain` (models + interfaces), `internal/config`, `internal/database`, `pkg/` for anything reusable outside this module.
2. **Config loading** — a `config` package that reads `MONEYFLOW_*` env vars into a typed struct, validates required fields, and fails fast (`log.Fatal`) on missing/invalid values before the server starts.
3. **Database layer** — `pgxpool` (or `database/sql` + `pgx` driver) connection pool, configurable pool limits and connection lifetimes, a startup ping, password URL-encoding for the DSN, and an embedded migration tool (e.g. `golang-migrate` with `embed.FS`).
4. **Gin engine + core middleware chain** — recovery (`gin.Recovery()` or a custom panic handler that logs instead of leaking a stack trace to the client), request ID middleware, structured request logging (method, path, status, latency, IP, user agent), CORS (`gin-contrib/cors`), secure headers, and rate limiting.
5. **Auth middleware** — verifies the Clerk JWT/session, extracts user ID and claims, and stores identity on the Gin context (`c.Set("userID", ...)`) so handlers/services can read it via `c.Get("userID")` or a typed context helper.
6. **Generic handler helpers** — shared functions for binding + validating request bodies (`c.ShouldBindJSON` + `go-playground/validator`, which Gin already builds on), writing consistent success/error JSON responses, and translating errors into the fixed error envelope.
7. **Centralized error handling** — sentinel domain errors (`ErrNotFound`, `ErrForbidden`, `ErrConflict`, `ErrPlanLimitExceeded`, `ErrValidation`) plus a single place (middleware or a handler wrapper) that maps them to HTTP status + the error envelope, so individual handlers never construct error JSON by hand.
8. **Observability** — New Relic Go agent wrapped around the Gin engine and the database driver for request and query tracing; structured JSON logging with the shared fields (timestamp, level, service, environment, request_id, user_id, operation, duration, error).
9. **Background jobs** — Redis connection + Asynq client/server setup, the three priority queues (`critical`/`default`/`low`), configurable worker concurrency, and graceful shutdown on SIGTERM/SIGINT.
10. **Email** — Resend client wrapper, HTML templates, and a background-task pattern (queued via Asynq rather than sent inline in a request).
11. **API scaffolding** — a `/api/v1` route group, health/system routes, and OpenAPI/Swagger generation (`swaggo/swag` is the common Gin-ecosystem choice) wired in early so every new endpoint gets documented as it's added rather than retrofitted later.
12. **Testing harness** — `testcontainers-go` spinning up a real Postgres per test run, migrations applied automatically in test setup, a teardown/cleanup hook per test, and shared helpers for building a test Gin router, making authenticated test requests, and asserting on the error envelope shape.

Once this foundation is in place and passing its own tests, domain features (Categories → Transactions → Budgets → …) get added as vertical slices through the same handler → service → repository layering, without touching the foundation again.

---

## 13. Core Architectural Principles (Full List, Updated)

1. Business logic lives in services, not HTTP handlers.
2. Infrastructure dependencies are hidden behind interfaces.
3. Every user-owned resource is authorized server-side.
4. Shared resources use explicit membership authorization.
5. Money is represented using integer minor units plus currency.
6. Long-running operations are asynchronous.
7. Every background job is retry-safe and idempotent.
8. Database constraints provide a final integrity boundary.
9. Subscription limits are enforced independently of the frontend.
10. External providers are accessed through abstractions.
11. Object storage is private and accessed through presigned URLs.
12. Observability is part of the application architecture, not an afterthought.
13. API behavior is explicitly documented through OpenAPI.
14. Integration tests use real infrastructure through Testcontainers.
15. The Nx monorepo provides a unified development and CI/CD workflow.
16. Configuration is validated and fails fast at startup — the app never silently runs with invalid defaults.
17. Errors are centrally handled and translated consistently from database/domain errors to HTTP responses without leaking internals.
18. Security is layered at the framework level (CORS, secure headers, panic recovery, rate limiting) in addition to domain-level authorization.

---

## 14. Implementation Plan — Nx + Gin Migration

This plan converts the former Turborepo/Echo starter into an Nx monorepo with a Gin backend while preserving the existing architectural goals and API behavior. The migration is intentionally incremental: each phase leaves the repository buildable and has a concrete exit check.

### 14.1 Current Baseline

The repository currently has:

* A Bun workspace with `apps/*` and `packages/*` plus Nx project targets.
* A Go module at `apps/backend`.
* Echo-specific routing, middleware, and handler abstractions in `internal/router`, `internal/middleware`, and `internal/handler`.
* Reusable backend infrastructure for configuration, PostgreSQL, Redis/Asynq, logging, New Relic, email, and testing.
* A React/Vite frontend and TypeScript packages for OpenAPI contracts and Zod schemas.

The migration must not mix framework conversion with new finance features. First establish the Gin foundation and contract tests; then implement the product as vertical slices.

### 14.2 Target Repository Layout

```text
.
├── apps/
│   ├── api/                         # Go module: Gin HTTP API
│   │   ├── cmd/api/main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── database/
│   │   │   ├── domain/              # models, ports, sentinel errors
│   │   │   ├── handler/              # Gin adapters and DTOs
│   │   │   ├── middleware/
│   │   │   ├── repository/
│   │   │   ├── service/
│   │   │   ├── worker/
│   │   │   └── router/
│   │   └── migrations/
│   ├── worker/                      # optional separate worker entrypoint
│   └── frontend/                    # React/Vite application
├── packages/
│   ├── openapi/                     # source contract and generated artifacts
│   ├── zod/                         # shared frontend/runtime schemas
│   └── emails/                      # React email templates
├── nx.json
├── project.json                     # root targets only when useful
└── package.json
```

`apps/backend` can be renamed to `apps/api` in a dedicated move commit, or retained temporarily while the Gin migration is in progress. Do not move files and rewrite framework code in the same commit; that makes failures difficult to localize.

### 14.3 Phase 0 — Freeze the Contract and Establish a Baseline

1. Record the current commands and supported local prerequisites: Bun, Node, Go, PostgreSQL, and Redis.
2. Run the existing backend tests, frontend checks, and package checks. Record known failures separately from migration failures.
3. Treat `packages/openapi/openapi.json` as the API compatibility reference. Resolve the existing `/v1` versus `/api/v1` discrepancy before adding routes; the target convention is `/api/v1` unless the contract is deliberately versioned otherwise.
4. Add or update smoke tests for health, authentication failure, validation failure, not-found, rate limiting, and a successful JSON response.
5. Define migration invariants: response envelopes, status codes, request IDs, auth claims, database schema, and graceful shutdown behavior must remain stable.

**Exit gate:** the baseline checks are reproducible and the contract tests fail if the error envelope or route prefix changes unexpectedly.

### 14.4 Phase 1 — Make Nx the Monorepo Orchestrator

1. Install Nx and add `nx.json` with named inputs, cacheable targets, and the default base branch configuration.
2. Generate or hand-author Nx project configuration for `frontend`, `openapi`, `zod`, `emails`, and `api`.
3. Use root Nx scripts for workspace orchestration:
  * `bun nx run-many -t build`
  * `bun nx run-many -t lint`
  * `bun nx run-many -t typecheck`
  * `bun nx affected -t lint test build`
  * `bun nx graph`
4. Keep Go commands explicit in the API project targets. Typical targets are `go test ./...`, `go vet ./...`, `gofmt -w`/check, `go build ./cmd/api`, and `go run ./cmd/api`.
5. Add project tags and dependencies so frontend packages depend on contract packages, while the Go API does not depend on frontend source packages.
6. Update CI to use `nx affected` for JavaScript/TypeScript projects and the API project target for Go. Retain a full verification job for release branches.
7. Keep the Nx configuration and project targets as the single workspace orchestration layer.

**Exit gate:** `nx graph` shows the intended project dependencies, affected checks pass on a small change, and a clean checkout can build every project through Nx.

### 14.5 Phase 2 — Extract Framework-Neutral Backend Boundaries

Before introducing Gin, reduce the Echo blast radius.

1. Move shared request/response DTOs, validation types, domain errors, and service ports away from Echo imports.
2. Keep `context.Context` as the service boundary. A Gin context may be used only in the handler adapter; services must not accept `*gin.Context`.
3. Define typed ports for repositories, storage, queue, email, billing, and clock/time where deterministic tests need them.
4. Make startup ownership explicit: configuration, database, Redis, job client, and worker lifecycle should be assembled in one application composition root.
5. Preserve existing database migrations and test helpers while adding framework-neutral service tests.

**Exit gate:** service and repository tests compile without importing Echo, and the existing HTTP tests still pass.

### 14.6 Phase 3 — Build the Gin Foundation

Implement the foundation in this order:

1. Create the Gin application composition root and route registration under `/api/v1`.
2. Replace Echo middleware with Gin adapters for recovery, request ID, structured request logging, CORS, secure headers, rate limiting, tracing, and context enrichment.
3. Implement Clerk authentication middleware that verifies the token, extracts the user ID and claims, and stores a typed identity in request context.
4. Replace Echo handler helpers with Gin helpers for binding, validation, JSON responses, file responses, and no-content responses.
5. Add one centralized error translator for domain, validation, database, and unknown errors. It must produce the fixed error envelope and never expose internal details.
6. Wrap the Gin engine and database operations with New Relic instrumentation, preserving request ID, user ID, operation, and duration fields.
7. Keep Asynq queues and worker lifecycle behind an internal queue package. Add graceful shutdown for HTTP, database, Redis, and workers.
8. Serve health/system routes and OpenAPI/Swagger documentation from the new router.
9. Use contract tests to compare Gin responses with the baseline behavior.

At this point the API should be able to boot, connect to required infrastructure, authenticate a request, return consistent errors, enqueue a test job, and shut down cleanly. Do not start domain feature work until this foundation is green.

**Exit gate:** Gin foundation tests pass with real Postgres/Redis integration where applicable, health and error contract tests pass, and no production package imports Echo.

### 14.7 Phase 4 — Migrate Existing Cross-Cutting Features

Migrate existing capabilities one at a time, preserving behavior:

1. Configuration and fail-fast startup validation.
2. Database pool, migrations, query tracing, and transaction helpers.
3. Logging, request IDs, New Relic tracing, and error reporting.
4. Asynq client, worker server, job handlers, and retry/idempotency behavior.
5. Resend email client, templates, preview endpoint, and welcome task.
6. OpenAPI generation and static documentation.
7. Testcontainers setup, server helpers, assertions, and authenticated request helpers.

Delete Echo-specific code only after its Gin replacement is covered and the old path is no longer referenced. The old and new implementations may coexist temporarily, but only one should own a route at a time.

**Exit gate:** all current non-domain capabilities run through Gin, and the old Echo dependency can be removed from `go.mod` without breaking tests or builds.

### 14.8 Phase 5 — Implement Product Features as Vertical Slices

Build each slice end to end: migration/schema, domain model, repository, service rules, Gin handlers/routes, OpenAPI contract, frontend API client, UI state, and tests.

1. **Accounts and categories:** local Clerk user synchronization, category CRUD, ownership, deletion conflict rules, and plan limits.
2. **Transactions:** CRUD, minor-unit money validation, category ownership, filtering, pagination, summary queries, and multi-currency grouping.
3. **Budgets:** monthly limits, usage calculations, member roles, ownership transfer, and shared-resource authorization.
4. **Subscriptions and plan guard:** subscription state, plan limits, `402 plan_limit_exceeded`, checkout/portal ports, and signed webhook idempotency.
5. **Recurring transactions:** rules, occurrence keys, scheduled worker, deterministic generation, and retry tests.
6. **Receipts:** presigned upload URLs, private object keys, ownership checks, and expiry behavior.
7. **CSV imports:** asynchronous import lifecycle, row-level validation, checkpoints, partial success, and plan limits.
8. **Reports:** asynchronous PDF/CSV generation, storage keys, status polling, expiring download URLs, and report retention limits.
9. **Notifications:** budget alerts, monthly summaries, deduplication, templates, and delivery failure handling.

Each slice must enforce authorization and plan limits in the service layer; frontend guards are for user experience only.

### 14.9 Phase 6 — Frontend Migration and Delivery Workflow

1. Keep the React/Vite app as an Nx project before changing UI architecture.
2. Generate or validate the TypeScript API client from the OpenAPI contract. Keep Zod schemas at the boundary for response and form validation.
3. Add Clerk provider/auth state, protected routes, API error handling, request IDs for support diagnostics, and subscription-aware feature states.
4. Build screens in the same vertical-slice order as the backend: dashboard, categories, transactions, budgets, imports, reports, and subscription settings.
5. Add loading, empty, validation, permission, plan-limit, retry, and asynchronous job states for every workflow.
6. Add frontend unit/component tests and a small number of browser-level smoke tests for sign-in, transaction creation, import status, and report download.

**Exit gate:** an authenticated user can complete the core category, transaction, budget, and report workflows against the Gin API in a clean local environment.

### 14.10 Phase 7 — Hardening and Release Readiness

1. Run `nx affected` checks in pull requests and full builds on release branches.
2. Run Go race tests where practical, migration tests, contract tests, and Testcontainers integration tests in CI.
3. Verify authorization with cross-user and shared-budget cases, including intentionally hidden `404` responses.
4. Verify idempotency by retrying every Asynq job and billing webhook fixture.
5. Load-test list, summary, import enqueue, and report enqueue endpoints. Confirm expensive work never runs in the request handler.
6. Add database backup/restore documentation, object-storage lifecycle policies, secret rotation notes, and operational runbooks.
7. Add dashboards and alerts for HTTP RED metrics, queue depth, failed jobs, database health, and provider webhook failures.

**Release gate:** all acceptance checks pass, migrations are reversible or have a documented forward-only recovery path, and rollback can return the API to the last known-good contract.

### 14.11 Recommended Commit/PR Sequence

Keep the migration reviewable with small, independently verifiable changes:

1. `chore: add nx workspace configuration and project targets`
2. `refactor(api): isolate framework-neutral ports and errors`
3. `feat(api): add gin composition root and core middleware`
4. `refactor(api): migrate auth, observability, jobs, and email to gin`
5. `test(api): add gin contract and integration coverage`
6. `chore: remove echo and retire legacy orchestration configuration`
7. `feat: implement product vertical slices`
8. `ci: enable nx affected checks and release verification`

Avoid a single large rewrite PR. Each PR should include its focused tests and should leave `nx affected -t lint test build` passing.

### 14.12 Main Risks and Mitigations

| Risk | Mitigation |
| ---- | ---------- |
| Echo-to-Gin changes status/error behavior | Freeze contract tests before migration and compare response envelopes. |
| Nx targets hide failing Go commands | Use explicit Go targets and run `go test ./...` in the API project target. |
| Framework types leak into services | Enforce a rule that services accept standard context and domain DTOs only. |
| Old and new routers serve different behavior | Migrate one route group at a time and keep a single route owner. |
| Background retries duplicate effects | Add deterministic keys, unique constraints, checkpoints, and retry tests before enabling workers. |
| Frontend and backend contracts drift | Generate or validate clients from OpenAPI in CI. |
| Migration scope expands into product work | Do not add finance features until the Gin foundation exit gate is green. |

### 14.13 Definition of Done

The migration is complete when:

* Nx owns local development, affected checks, builds, tests, linting, and CI orchestration.
* The backend runs on Gin with no production Echo dependency.
* `/api/v1`, error envelopes, auth behavior, request IDs, and documented status codes are contract-tested.
* Postgres, Redis/Asynq, Clerk, storage, email, billing, and observability integrations remain behind replaceable interfaces.
* The frontend uses the shared OpenAPI/Zod contract and handles asynchronous and plan-limited workflows.
* Product vertical slices have service-level authorization, integration coverage, and idempotent background processing.
* A clean checkout can start the stack and run the focused Nx affected pipeline successfully.