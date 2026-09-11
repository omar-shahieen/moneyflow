package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*model.UserAccount, error)
	Create(ctx context.Context, user *model.UserAccount) error
	Update(ctx context.Context, user *model.UserAccount) error
}

type CategoryRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.Category, error)
	List(ctx context.Context, userID string) ([]model.Category, error)
	Create(ctx context.Context, category *model.Category) error
	Update(ctx context.Context, category *model.Category) error
	Delete(ctx context.Context, id uuid.UUID, userID string) error
	CountByUser(ctx context.Context, userID string) (int, error)
}

type TransactionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.Transaction, error)
	List(ctx context.Context, userID string, filter TransactionFilter) ([]model.Transaction, int, error)
	Create(ctx context.Context, transaction *model.Transaction) error
	Update(ctx context.Context, transaction *model.Transaction) error
	Delete(ctx context.Context, id uuid.UUID, userID string) error
	CountByUser(ctx context.Context, userID string) (int, error)
	Summary(ctx context.Context, userID string, month time.Time) (*TransactionSummary, error)
}

type TransactionFilter struct {
	CategoryID *uuid.UUID
	Type       *string
	From       *time.Time
	To         *time.Time
	Page       int
	Limit      int
}

type TransactionSummary struct {
	TotalIncome  int64           `json:"total_income"`
	TotalExpense int64           `json:"total_expense"`
	ByCategory   []CategoryTotal `json:"by_category"`
	ByCurrency   []CurrencyTotal `json:"by_currency"`
}

type CategoryTotal struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Total        int64     `json:"total"`
	Type         string    `json:"type"`
}

type CurrencyTotal struct {
	Currency     string `json:"currency"`
	TotalIncome  int64  `json:"total_income"`
	TotalExpense int64  `json:"total_expense"`
}

type BudgetRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Budget, error)
	ListByUser(ctx context.Context, userID string) ([]model.BudgetWithMembers, error)
	Create(ctx context.Context, budget *model.Budget) error
	Update(ctx context.Context, budget *model.Budget) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByUser(ctx context.Context, userID string) (int, error)
	GetUsage(ctx context.Context, budgetID uuid.UUID, month time.Time) (int64, error)
}

type BudgetMemberRepository interface {
	Get(ctx context.Context, budgetID uuid.UUID, userID string) (*model.BudgetMember, error)
	ListByBudget(ctx context.Context, budgetID uuid.UUID) ([]model.BudgetMember, error)
	Add(ctx context.Context, member *model.BudgetMember) error
	Remove(ctx context.Context, budgetID uuid.UUID, userID string) error
	CountByBudget(ctx context.Context, budgetID uuid.UUID) (int, error)
}

type RecurringRuleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.RecurringRule, error)
	List(ctx context.Context, userID string) ([]model.RecurringRule, error)
	Create(ctx context.Context, rule *model.RecurringRule) error
	Update(ctx context.Context, rule *model.RecurringRule) error
	Delete(ctx context.Context, id uuid.UUID, userID string) error
	GetDueRules(ctx context.Context, before time.Time) ([]model.RecurringRule, error)
}

type ReportRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.Report, error)
	List(ctx context.Context, userID string, limit int) ([]model.Report, error)
	Create(ctx context.Context, report *model.Report) error
	Update(ctx context.Context, report *model.Report) error
	CountByUser(ctx context.Context, userID string) (int, error)
}

type ImportRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.Import, error)
	Create(ctx context.Context, imp *model.Import) error
	Update(ctx context.Context, imp *model.Import) error
}

type SubscriptionRepository interface {
	GetByUserID(ctx context.Context, userID string) (*model.Subscription, error)
	Create(ctx context.Context, sub *model.Subscription) error
	Update(ctx context.Context, sub *model.Subscription) error
}
