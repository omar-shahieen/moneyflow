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
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
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
	CreatedAt              time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time          `json:"updated_at" db:"updated_at"`
}
