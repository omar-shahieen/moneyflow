package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/handler"
)

func registerBudgetRoutes(r *gin.RouterGroup, h *handler.BudgetHandler) {
	budgets := r.Group("/budgets")
	{
		budgets.GET("", h.List)
		budgets.GET("/:id", h.GetByID)
		budgets.POST("", h.Create)
		budgets.PATCH("/:id", h.Update)
		budgets.POST("/:id/members", h.AddMember)
		budgets.DELETE("/:id/members/:userId", h.RemoveMember)
	}
}
