package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type CategoryHandler struct {
	service *service.CategoryService
}

func NewCategoryHandler(s *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: s}
}

type CreateCategoryRequest struct {
	Name string             `json:"name" binding:"required"`
	Type model.CategoryType `json:"type" binding:"required"`
}

type UpdateCategoryRequest struct {
	Name string             `json:"name" binding:"required"`
	Type model.CategoryType `json:"type" binding:"required"`
}

func (h *CategoryHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	categories, err := h.service.List(c.Request.Context(), userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	if categories == nil {
		categories = []model.Category{}
	}

	RespondJSON(c, http.StatusOK, categories)
}

func (h *CategoryHandler) GetByID(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	category, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusOK, category)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	var req CreateCategoryRequest
	if err := BindAndValidate(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	category, err := h.service.Create(c.Request.Context(), userID, service.CreateCategoryInput{
		Name: req.Name,
		Type: req.Type,
	})
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusCreated, category)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	var req UpdateCategoryRequest
	if err := BindAndValidate(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	category, err := h.service.Update(c.Request.Context(), id, userID, service.UpdateCategoryInput{
		Name: req.Name,
		Type: req.Type,
	})
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusOK, category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	if err := h.service.Delete(c.Request.Context(), id, userID); err != nil {
		RespondError(c, err)
		return
	}

	RespondNoContent(c)
}
