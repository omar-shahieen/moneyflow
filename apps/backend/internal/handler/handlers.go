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
	Recurring    *RecurringRuleHandler
	Receipt      *ReceiptHandler
	Import       *ImportHandler
	Report       *ReportHandler
}

func NewHandlers(s *server.Server, svc *service.Services) *Handlers {
	return &Handlers{
		Health:       NewHealthHandler(s),
		OpenAPI:      NewOpenAPIHandler(s),
		Category:     NewCategoryHandler(s, svc.CategoryService),
		Transaction:  NewTransactionHandler(s, svc.TransactionService),
		Budget:       NewBudgetHandler(s, svc.BudgetService),
		Subscription: NewSubscriptionHandler(s),
		Recurring:    NewRecurringRuleHandler(s, svc.RecurringRuleService),
		Receipt:      NewReceiptHandler(s, svc.Storage),
		Import:       NewImportHandler(s, svc.ImportService),
		Report:       NewReportHandler(s, svc.ReportService),
	}
}
