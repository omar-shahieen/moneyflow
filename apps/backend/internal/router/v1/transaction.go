package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/handler"
)

func registerTransactionRoutes(r *gin.RouterGroup, h *handler.TransactionHandler, receipt *handler.ReceiptHandler) {
	transactions := r.Group("/transactions")
	{
		transactions.GET("", h.List)
		transactions.GET("/summary", h.Summary)
		transactions.GET("/:id", h.GetByID)
		transactions.POST("", h.Create)
		transactions.PATCH("/:id", h.Update)
		transactions.DELETE("/:id", h.Delete)
		transactions.POST("/:id/receipt", receipt.GetUploadURL)
	}
}
