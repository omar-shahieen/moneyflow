package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/handler"
)

func RegisterV1Routes(r *gin.RouterGroup, h *handler.Handlers) {
	registerCategoryRoutes(r, h.Category)
	registerTransactionRoutes(r, h.Transaction, h.Receipt)
	registerBudgetRoutes(r, h.Budget)
	registerSubscriptionRoutes(r, h.Subscription)
	registerRecurringRoutes(r, h.Recurring)
	registerImportRoutes(r, h.Import)
	registerReportRoutes(r, h.Report)
}
