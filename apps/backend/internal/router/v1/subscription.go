package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/handler"
)

func registerSubscriptionRoutes(r *gin.RouterGroup, h *handler.SubscriptionHandler) {
	subscription := r.Group("/subscription")
	{
		subscription.GET("", h.Get)
	}
}
