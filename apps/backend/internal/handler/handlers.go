package handler

import (
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type Handlers struct {
	Health       *HealthHandler
	OpenAPI      *OpenAPIHandler
	Category     *CategoryHandler
	Transaction  *TransactionHandler
	Budget       *BudgetHandler
	Subscription *SubscriptionHandler
	Billing      *BillingHandler
	Recurring    *RecurringRuleHandler
	Receipt      *ReceiptHandler
	Import       *ImportHandler
	Report       *ReportHandler
	User         *UserHandler
}

func NewHandlers(s *server.Server, svc *service.Services) *Handlers {
	return &Handlers{
		Health:       NewHealthHandler(s),
		OpenAPI:      NewOpenAPIHandler(s),
		Category:     NewCategoryHandler(s, svc.CategoryService),
		Transaction:  NewTransactionHandler(s, svc.TransactionService),
		Budget:       NewBudgetHandler(s, svc.BudgetService),
		Subscription: NewSubscriptionHandler(s, svc.SubscriptionService),
		Billing:      NewBillingHandler(s, svc.BillingService),
		Recurring:    NewRecurringRuleHandler(s, svc.RecurringRuleService),
		Receipt:      NewReceiptHandler(s, svc.Storage),
		Import:       NewImportHandler(s, svc.ImportService, svc.Storage),
		Report:       NewReportHandler(s, svc.ReportService),
		User:         NewUserHandler(s, svc.UserService),
	}
}
