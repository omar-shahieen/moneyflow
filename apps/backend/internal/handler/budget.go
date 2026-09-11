package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type BudgetHandler struct {
	service *service.BudgetService
}

func NewBudgetHandler(s *service.BudgetService) *BudgetHandler {
	return &BudgetHandler{service: s}
}

type CreateBudgetRequest struct {
	CategoryID        string `json:"category_id" binding:"required"`
	MonthlyLimitMinor int64  `json:"monthly_limit_minor" binding:"required"`
	Currency          string `json:"currency"`
}

type UpdateBudgetRequest struct {
	CategoryID        string `json:"category_id" binding:"required"`
	MonthlyLimitMinor int64  `json:"monthly_limit_minor" binding:"required"`
	Currency          string `json:"currency"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type BudgetResponse struct {
	model.BudgetWithMembers
	Usage        int64   `json:"usage"`
	UsagePercent float64 `json:"usage_percent"`
	Exceeded     bool    `json:"exceeded"`
}

func (h *BudgetHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	budgets, err := h.service.List(c.Request.Context(), userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	if budgets == nil {
		budgets = []model.BudgetWithMembers{}
	}

	responses := make([]BudgetResponse, 0, len(budgets))
	now := time.Now()
	for _, b := range budgets {
		usage, _ := h.service.GetUsage(c.Request.Context(), b.ID, now)
		var usagePercent float64
		if b.MonthlyLimitMinor > 0 {
			usagePercent = float64(usage) / float64(b.MonthlyLimitMinor) * 100
		}
		responses = append(responses, BudgetResponse{
			BudgetWithMembers: b,
			Usage:             usage,
			UsagePercent:      usagePercent,
			Exceeded:          usage > b.MonthlyLimitMinor,
		})
	}

	RespondJSON(c, http.StatusOK, responses)
}

func (h *BudgetHandler) GetByID(c *gin.Context) {
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

	budget, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	usage, _ := h.service.GetUsage(c.Request.Context(), id, time.Now())
	var usagePercent float64
	if budget.MonthlyLimitMinor > 0 {
		usagePercent = float64(usage) / float64(budget.MonthlyLimitMinor) * 100
	}

	RespondJSON(c, http.StatusOK, BudgetResponse{
		BudgetWithMembers: *budget,
		Usage:             usage,
		UsagePercent:      usagePercent,
		Exceeded:          usage > budget.MonthlyLimitMinor,
	})
}

func (h *BudgetHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	var req CreateBudgetRequest
	if err := BindAndValidate(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	budget, err := h.service.Create(c.Request.Context(), userID, service.CreateBudgetInput{
		CategoryID:        categoryID,
		MonthlyLimitMinor: req.MonthlyLimitMinor,
		Currency:          req.Currency,
	})
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusCreated, budget)
}

func (h *BudgetHandler) Update(c *gin.Context) {
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

	var req UpdateBudgetRequest
	if err := BindAndValidate(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	budget, err := h.service.Update(c.Request.Context(), id, userID, service.UpdateBudgetInput{
		CategoryID:        categoryID,
		MonthlyLimitMinor: req.MonthlyLimitMinor,
		Currency:          req.Currency,
	})
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusOK, budget)
}

func (h *BudgetHandler) AddMember(c *gin.Context) {
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

	var req AddMemberRequest
	if err := BindAndValidate(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	if err := h.service.AddMember(c.Request.Context(), id, userID, service.AddMemberInput{
		UserID: req.UserID,
	}); err != nil {
		RespondError(c, err)
		return
	}

	RespondNoContent(c)
}

func (h *BudgetHandler) RemoveMember(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	budgetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		RespondError(c, ErrInvalidID)
		return
	}

	if err := h.service.RemoveMember(c.Request.Context(), budgetID, userID, targetUserID); err != nil {
		RespondError(c, err)
		return
	}

	RespondNoContent(c)
}
