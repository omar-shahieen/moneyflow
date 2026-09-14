package ports

import "time"

type BillingProvider interface {
	CreateCheckoutSession(userID string, plan string, successURL string, cancelURL string) (string, error)
	CreateCustomerPortal(userID string, returnURL string) (string, error)
	VerifyWebhookSignature(payload []byte, signature string) (bool, error)
	ParseWebhookEvent(payload []byte) (*BillingEvent, error)
}

type BillingEvent struct {
	ProviderEventID string
	EventType       string
	CustomerID      string
	SubscriptionID  string
	Plan            string
	Status          string
	PeriodEnd       *time.Time
}
