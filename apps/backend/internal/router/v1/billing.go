package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/handler"
)

func registerBillingRoutes(r *gin.RouterGroup, h *handler.BillingHandler) {
	billing := r.Group("/billing")
	{
		billing.POST("/checkout", h.CreateCheckout)
		billing.GET("/status/:ref_number", h.GetStatus)
		billing.GET("/history", h.GetHistory)
	}
}
