package budget

import (
	"time"

	"github.com/google/uuid"
)

type Budget struct {
	ID                uuid.UUID `json:"id" db:"id"`
	CategoryID        uuid.UUID `json:"category_id" db:"category_id"`
	MonthlyLimitMinor int64     `json:"monthly_limit_minor" db:"monthly_limit_minor"`
	Currency          string    `json:"currency" db:"currency"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

type BudgetMemberRole string

const (
	BudgetMemberRoleOwner  BudgetMemberRole = "owner"
	BudgetMemberRoleMember BudgetMemberRole = "member"
)

type BudgetMember struct {
	BudgetID uuid.UUID        `json:"budget_id" db:"budget_id"`
	UserID   string           `json:"user_id" db:"user_id"`
	Role     BudgetMemberRole `json:"role" db:"role"`
}

type BudgetWithMembers struct {
	Budget
	Members []BudgetMember `json:"members"`
}
