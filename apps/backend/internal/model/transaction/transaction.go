package transaction

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	CategoryID  uuid.UUID `json:"category_id" db:"category_id"`
	AmountMinor int64     `json:"amount_minor" db:"amount_minor"`
	Note        string    `json:"note" db:"note"`
	ReceiptKey      string    `json:"receipt_key,omitempty" db:"receipt_key"`
	BillingProvider string    `json:"billing_provider,omitempty" db:"billing_provider"`
	BillingRefID    string    `json:"billing_ref_id,omitempty" db:"billing_ref_id"`
	BillingStatus   string    `json:"billing_status" db:"billing_status"`
	OccurredAt      time.Time `json:"occurred_at" db:"occurred_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}
