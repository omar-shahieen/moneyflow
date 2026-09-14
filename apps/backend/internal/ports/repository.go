package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/budget"
	"github.com/omar-shahieen/moneyflow/internal/model/category"
	"github.com/omar-shahieen/moneyflow/internal/model/imports"
	"github.com/omar-shahieen/moneyflow/internal/model/recurring"
	"github.com/omar-shahieen/moneyflow/internal/model/report"
	"github.com/omar-shahieen/moneyflow/internal/model/subscription"
	"github.com/omar-shahieen/moneyflow/internal/model/transaction"
	"github.com/omar-shahieen/moneyflow/internal/model/user"
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*user.UserAccount, error)
	Create(ctx context.Context, u *user.UserAccount) error
	Update(ctx context.Context, u *user.UserAccount) error
}

type CategoryRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*category.Category, error)
	List(ctx context.Context, userID string, query *model.ListQuery) (*model.PaginatedResponse[category.Category], error)
	Create(ctx context.Context, c *category.Category) error
	Update(ctx context.Context, c *category.Category) error
	Delete(ctx context.Context, id uuid.UUID, userID string) error
	CountByUser(ctx context.Context, userID string) (int, error)
}

type TransactionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*transaction.Transaction, error)
	List(ctx context.Context, userID string, query *model.ListQuery, filters TransactionFilters) (*model.PaginatedResponse[transaction.Transaction], error)
	Create(ctx context.Context, t *transaction.Transaction) error
	Update(ctx context.Context, t *transaction.Transaction) error
	Delete(ctx context.Context, id uuid.UUID, userID string) error
	CountByUser(ctx context.Context, userID string) (int, error)
	Summary(ctx context.Context, userID string, month time.Time) (*TransactionSummary, error)
}

type TransactionFilters struct {
	CategoryID *uuid.UUID
	Type       *string
	From       *time.Time
	To         *time.Time
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
	GetByID(ctx context.Context, id uuid.UUID) (*budget.Budget, error)
	ListByUser(ctx context.Context, userID string, query *model.ListQuery) (*model.PaginatedResponse[budget.BudgetWithMembers], error)
	Create(ctx context.Context, b *budget.Budget) error
	Update(ctx context.Context, b *budget.Budget) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByUser(ctx context.Context, userID string) (int, error)
	GetUsage(ctx context.Context, budgetID uuid.UUID, month time.Time) (int64, error)
}

type BudgetMemberRepository interface {
	Get(ctx context.Context, budgetID uuid.UUID, userID string) (*budget.BudgetMember, error)
	ListByBudget(ctx context.Context, budgetID uuid.UUID) ([]budget.BudgetMember, error)
	Add(ctx context.Context, m *budget.BudgetMember) error
	Remove(ctx context.Context, budgetID uuid.UUID, userID string) error
	CountByBudget(ctx context.Context, budgetID uuid.UUID) (int, error)
}

type RecurringRuleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*recurring.RecurringRule, error)
	List(ctx context.Context, userID string, query *model.ListQuery) (*model.PaginatedResponse[recurring.RecurringRule], error)
	Create(ctx context.Context, r *recurring.RecurringRule) error
	Update(ctx context.Context, r *recurring.RecurringRule) error
	Delete(ctx context.Context, id uuid.UUID, userID string) error
	GetDueRules(ctx context.Context, before time.Time) ([]recurring.RecurringRule, error)
	UpdateNextRunDate(ctx context.Context, ruleID uuid.UUID, nextRun time.Time) error
	IsOccurrenceGenerated(ctx context.Context, ruleID uuid.UUID, occurrenceDate time.Time) (bool, error)
	MarkOccurrenceGenerated(ctx context.Context, ruleID uuid.UUID, occurrenceDate time.Time) error
}

type ReportRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*report.Report, error)
	List(ctx context.Context, userID string, query *model.ListQuery) (*model.PaginatedResponse[report.Report], error)
	Create(ctx context.Context, r *report.Report) error
	Update(ctx context.Context, r *report.Report) error
	CountByUser(ctx context.Context, userID string) (int, error)
}

type ImportRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*imports.Import, error)
	Create(ctx context.Context, i *imports.Import) error
	Update(ctx context.Context, i *imports.Import) error
	EnsureNoActiveImport(ctx context.Context, userID string) error
}

type SubscriptionRepository interface {
	GetByUserID(ctx context.Context, userID string) (*subscription.Subscription, error)
	Create(ctx context.Context, s *subscription.Subscription) error
	Update(ctx context.Context, s *subscription.Subscription) error
}
