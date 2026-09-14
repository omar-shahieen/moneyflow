package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/handler"
)

func registerRecurringRoutes(r *gin.RouterGroup, h *handler.RecurringRuleHandler) {
	recurringRules := r.Group("/recurring-rules")
	{
		recurringRules.GET("", h.List)
		recurringRules.GET("/:id", h.GetByID)
		recurringRules.POST("", h.Create)
		recurringRules.PATCH("/:id", h.Update)
		recurringRules.DELETE("/:id", h.Delete)
	}
}
