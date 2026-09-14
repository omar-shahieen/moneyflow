package repository

import "github.com/omar-shahieen/moneyflow/internal/server"

type Repositories struct {
	Category     *CategoryRepo
	Transaction  *TransactionRepo
	Budget       *BudgetRepo
	BudgetMember *BudgetMemberRepo
	Recurring    *RecurringRuleRepo
	Report       *ReportRepo
	Import       *ImportRepo
	Subscription *SubscriptionRepo
	User         *UserRepo
}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{
		Category:     NewCategoryRepository(s),
		Transaction:  NewTransactionRepository(s),
		Budget:       NewBudgetRepository(s),
		BudgetMember: NewBudgetMemberRepository(s),
		Recurring:    NewRecurringRuleRepository(s),
		Report:       NewReportRepository(s),
		Import:       NewImportRepository(s),
		Subscription: NewSubscriptionRepository(s),
		User:         NewUserRepository(s),
	}
}
