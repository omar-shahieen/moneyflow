package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/handler"
)

func registerReportRoutes(r *gin.RouterGroup, h *handler.ReportHandler) {
	reports := r.Group("/reports")
	{
		reports.GET("", h.List)
		reports.GET("/:id", h.GetByID)
		reports.POST("", h.Create)
	}
}
