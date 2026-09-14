package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/omar-shahieen/moneyflow/internal/server"

	"github.com/gin-gonic/gin"
)

type OpenAPIHandler struct {
	Handler
}

func NewOpenAPIHandler(s *server.Server) *OpenAPIHandler {
	return &OpenAPIHandler{
		Handler: NewHandler(s),
	}
}

func (h *OpenAPIHandler) ServeOpenAPIUI(c *gin.Context) {
	templateBytes, err := os.ReadFile("static/openapi.html")
	c.Header("Cache-Control", "no-cache")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to read OpenAPI UI template: %v", err),
		})
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", templateBytes)
}
