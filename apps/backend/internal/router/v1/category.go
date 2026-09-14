package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/handler"
)

func registerCategoryRoutes(r *gin.RouterGroup, h *handler.CategoryHandler) {
	categories := r.Group("/categories")
	{
		categories.GET("", h.List)
		categories.POST("", h.Create)
		categories.GET("/:id", h.GetByID)
		categories.PATCH("/:id", h.Update)
		categories.DELETE("/:id", h.Delete)
	}
}
