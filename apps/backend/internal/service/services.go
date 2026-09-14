package service

import (
	"github.com/omar-shahieen/moneyflow/internal/lib/job"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/storage"
)

type Services struct {
	Auth                 *AuthService
	Job                  *job.JobService
	CategoryService      *CategoryService
	BudgetService        *BudgetService
	ImportService        *ImportService
	ReportService        *ReportService
	TransactionService   *TransactionService
	RecurringRuleService *RecurringRuleService
	Storage              *storage.LocalStorage
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)
	localStorage := storage.NewLocalStorage("uploads")

	categoryService := NewCategoryService(s, repos.Category, repos.User)
	budgetService := NewBudgetService(s, repos.Budget, repos.BudgetMember, repos.Category)
	transactionService := NewTransactionService(s, repos.Transaction, repos.Category)
	importService := NewImportService(s, repos.Import, repos.Transaction, repos.Category)
	reportService := NewReportService(s, repos.Report, repos.Transaction, localStorage)
	recurringRuleService := NewRecurringRuleService(s, repos.Recurring, repos.Category, repos.Transaction)

	return &Services{
		Auth:                 authService,
		Job:                  s.Job,
		CategoryService:      categoryService,
		BudgetService:        budgetService,
		ImportService:        importService,
		ReportService:        reportService,
		TransactionService:   transactionService,
		RecurringRuleService: recurringRuleService,
		Storage:              localStorage,
	}, nil
}
