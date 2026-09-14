package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/handler"
)

func registerImportRoutes(r *gin.RouterGroup, h *handler.ImportHandler) {
	imports := r.Group("/imports")
	{
		imports.POST("", h.Create)
		imports.GET("/:id", h.GetByID)
	}
}
