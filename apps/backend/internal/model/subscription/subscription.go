package subscription

import (
	"time"

	"github.com/google/uuid"
)

type Plan string

const (
	PlanFree Plan = "free"
	PlanPro  Plan = "pro"
	PlanVIP  Plan = "vip"
)

type SubscriptionStatus string

const (
	SubscriptionStatusIncomplete SubscriptionStatus = "incomplete"
	SubscriptionStatusPending    SubscriptionStatus = "pending"
	SubscriptionStatusActive     SubscriptionStatus = "active"
	SubscriptionStatusPastDue    SubscriptionStatus = "past_due"
	SubscriptionStatusCancelled  SubscriptionStatus = "cancelled"
	SubscriptionStatusCanceled   SubscriptionStatus = "canceled"
	SubscriptionStatusExpired    SubscriptionStatus = "expired"
)

type Subscription struct {
	ID                     uuid.UUID          `json:"id" db:"id"`
	UserID                 string             `json:"user_id" db:"user_id"`
	Plan                   Plan               `json:"plan" db:"plan"`
	Status                 SubscriptionStatus `json:"status" db:"status"`
	PaymentProvider        string             `json:"payment_provider,omitempty" db:"payment_provider"`
	ProviderCustomerID     string             `json:"provider_customer_id,omitempty" db:"provider_customer_id"`
	ProviderSubscriptionID string             `json:"provider_subscription_id,omitempty" db:"provider_subscription_id"`
	CurrentPeriodEnd       *time.Time         `json:"current_period_end,omitempty" db:"current_period_end"`
	ExpiresAt              *time.Time         `json:"expires_at,omitempty" db:"expires_at"`
	CancelAtPeriodEnd      bool               `json:"cancel_at_period_end" db:"cancel_at_period_end"`
	CancelledAt            *time.Time         `json:"cancelled_at,omitempty" db:"cancelled_at"`
	FailedAttempts         int                `json:"failed_attempts" db:"failed_attempts"`
	NextBillingAt          *time.Time         `json:"next_billing_at,omitempty" db:"next_billing_at"`
	CreatedAt              time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time          `json:"updated_at" db:"updated_at"`
}

type Entitlements struct {
	Plan             Plan `json:"plan"`
	IsEntitled       bool `json:"is_entitled"`
	MaxCategories    int  `json:"max_categories"`
	MaxTransactions  int  `json:"max_transactions"`
	MaxBudgets       int  `json:"max_budgets"`
	MaxCSVRows       int  `json:"max_csv_rows"`
	MaxReports       int  `json:"max_reports"`
	HasRecurring     bool `json:"has_recurring"`
	HasSharedBudgets bool `json:"has_shared_budgets"`
	HasCSVImport     bool `json:"has_csv_import"`
	HasReceipts      bool `json:"has_receipts"`
	HasPDFReports    bool `json:"has_pdf_reports"`
}
