package service

import (
	"github.com/omar-shahieen/moneyflow/internal/lib/billing"
	"github.com/omar-shahieen/moneyflow/internal/lib/job"
	"github.com/omar-shahieen/moneyflow/internal/ports"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/storage"
	"github.com/omar-shahieen/moneyflow/internal/worker"
	"github.com/rs/zerolog"
)

type Services struct {
	Auth                   *AuthService
	Job                    *job.JobService
	CategoryService        *CategoryService
	BudgetService          *BudgetService
	ImportService          *ImportService
	ReportService          *ReportService
	TransactionService     *TransactionService
	RecurringRuleService   *RecurringRuleService
	SubscriptionService    *SubscriptionService
	BillingService         *BillingService
	RecurringBillingWorker *worker.RecurringBillingWorker
	ReconciliationWorker   *worker.ReconciliationWorker
	Storage                ports.Storage
	UserService            *UserService
}

func NewServices(s *server.Server, repos *repository.Repositories, logger *zerolog.Logger) (*Services, error) {
	var store ports.Storage
	switch s.Config.Storage.Provider {
	case "r2":
		r2Store, err := storage.NewR2Storage(s.Config.Storage.R2, logger)
		if err != nil {
			return nil, err
		}
		store = r2Store
	default:
		basePath := s.Config.Storage.Local.BasePath
		if basePath == "" {
			basePath = "uploads"
		}
		store = storage.NewLocalStorage(basePath)
	}

	authService := NewAuthService(s)

	billingClient := billing.NewClient(s.Config, s.Logger)

	categoryService := NewCategoryService(s, repos.Category, repos.User)
	budgetService := NewBudgetService(s, repos.Budget, repos.BudgetMember, repos.Category)
	transactionService := NewTransactionService(s, repos.Transaction, repos.Category)
	importService := NewImportService(s, repos.Import, repos.Transaction, repos.Category, store)
	reportService := NewReportService(s, repos.Report, repos.Transaction, store)
	recurringRuleService := NewRecurringRuleService(s, repos.Recurring, repos.Category, repos.Transaction)
	userService := NewUserService(s, repos.User, repos.Notification)

	subscriptionService := NewSubscriptionService(s, repos.Subscription, repos.User, repos.Transaction, billingClient)
	billingService := NewBillingService(s, billingClient, repos.Transaction, repos.Subscription, repos.BillingEvent, repos.Category, subscriptionService)

	recurringBillingWorker := worker.NewRecurringBillingWorker(repos.Subscription, repos.User, billingClient, repos.BillingEvent, &s.Config.Billing, s.Logger)
	reconciliationWorker := worker.NewReconciliationWorker(repos.Subscription, repos.User, billingClient, &s.Config.Billing, s.Logger)

	return &Services{
		Auth:                   authService,
		Job:                    s.Job,
		CategoryService:        categoryService,
		BudgetService:          budgetService,
		ImportService:          importService,
		ReportService:          reportService,
		TransactionService:     transactionService,
		RecurringRuleService:   recurringRuleService,
		SubscriptionService:    subscriptionService,
		BillingService:         billingService,
		RecurringBillingWorker: recurringBillingWorker,
		ReconciliationWorker:   reconciliationWorker,
		Storage:                store,
		UserService:            userService,
	}, nil
}
