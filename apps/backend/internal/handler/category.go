package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/category"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type CategoryHandler struct {
	Handler
	categoryService *service.CategoryService
}

func NewCategoryHandler(s *server.Server, categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		Handler:         NewHandler(s),
		categoryService: categoryService,
	}
}

func (h *CategoryHandler) List(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, query *category.ListCategoriesRequest) (*model.PaginatedResponse[category.Category], error) {
			userID := GetUserID(c)
			return h.categoryService.GetCategories(c.Request.Context(), userID, query)
		},
		http.StatusOK,
		&category.ListCategoriesRequest{},
	)(c)
}

func (h *CategoryHandler) GetByID(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *category.GetCategoryRequest) (*category.Category, error) {
			userID := GetUserID(c)
			return h.categoryService.GetCategoryByID(c.Request.Context(), userID, req.ID)
		},
		http.StatusOK,
		&category.GetCategoryRequest{},
	)(c)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, payload *category.CreateCategoryRequest) (*category.Category, error) {
			userID := GetUserID(c)
			return h.categoryService.CreateCategory(c.Request.Context(), userID, payload)
		},
		http.StatusCreated,
		&category.CreateCategoryRequest{},
	)(c)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, payload *category.UpdateCategoryRequest) (*category.Category, error) {
			userID := GetUserID(c)
			return h.categoryService.UpdateCategory(c.Request.Context(), userID, payload.ID, payload)
		},
		http.StatusOK,
		&category.UpdateCategoryRequest{},
	)(c)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	HandleNoContent(
		h.Handler,
		func(c *gin.Context, payload *category.DeleteCategoryRequest) error {
			userID := GetUserID(c)
			return h.categoryService.DeleteCategory(c.Request.Context(), userID, payload.ID)
		},
		http.StatusNoContent,
		&category.DeleteCategoryRequest{},
	)(c)
}
