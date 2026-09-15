package recurring

import (
	"time"

	"github.com/google/uuid"
)

type RecurringFrequency string

const (
	RecurringFrequencyWeekly  RecurringFrequency = "weekly"
	RecurringFrequencyMonthly RecurringFrequency = "monthly"
)

type RecurringRule struct {
	ID                uuid.UUID          `json:"id" db:"id"`
	UserID            string             `json:"user_id" db:"user_id"`
	CategoryID        uuid.UUID          `json:"category_id" db:"category_id"`
	AmountMinor       int64              `json:"amount_minor" db:"amount_minor"`
	Frequency         RecurringFrequency `json:"frequency" db:"frequency"`
	NextRunDate       time.Time          `json:"next_run_date" db:"next_run_date"`
	LastGeneratedDate *time.Time         `json:"last_generated_date,omitempty" db:"last_generated_date"`
	EndDate           *time.Time         `json:"end_date,omitempty" db:"end_date"`
}
