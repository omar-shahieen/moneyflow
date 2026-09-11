package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type RecurringRuleHandler struct {
	service *service.RecurringRuleService
}

func NewRecurringRuleHandler(s *service.RecurringRuleService) *RecurringRuleHandler {
	return &RecurringRuleHandler{service: s}
}

type CreateRecurringRuleRequest struct {
	CategoryID  string `json:"category_id" binding:"required"`
	AmountMinor int64  `json:"amount_minor" binding:"required"`
	Currency    string `json:"currency"`
	Frequency   string `json:"frequency" binding:"required"`
	NextRunDate string `json:"next_run_date" binding:"required"`
	EndDate     string `json:"end_date"`
}

type UpdateRecurringRuleRequest struct {
	CategoryID  string `json:"category_id" binding:"required"`
	AmountMinor int64  `json:"amount_minor" binding:"required"`
	Currency    string `json:"currency"`
	Frequency   string `json:"frequency" binding:"required"`
	NextRunDate string `json:"next_run_date" binding:"required"`
	EndDate     string `json:"end_date"`
}

func (h *RecurringRuleHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	rules, err := h.service.List(c.Request.Context(), userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	if rules == nil {
		rules = []model.RecurringRule{}
	}

	RespondJSON(c, http.StatusOK, rules)
}

func (h *RecurringRuleHandler) GetByID(c *gin.Context) {
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

	rule, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusOK, rule)
}

func (h *RecurringRuleHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	var req CreateRecurringRuleRequest
	if err := BindAndValidate(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	nextRunDate, err := time.Parse("2006-01-02", req.NextRunDate)
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	var endDate *time.Time
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = &t
		}
	}

	rule, err := h.service.Create(c.Request.Context(), userID, service.CreateRecurringRuleInput{
		CategoryID:  categoryID,
		AmountMinor: req.AmountMinor,
		Currency:    req.Currency,
		Frequency:   model.RecurringFrequency(req.Frequency),
		NextRunDate: nextRunDate,
		EndDate:     endDate,
	})
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusCreated, rule)
}

func (h *RecurringRuleHandler) Update(c *gin.Context) {
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

	var req UpdateRecurringRuleRequest
	if err := BindAndValidate(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	nextRunDate, err := time.Parse("2006-01-02", req.NextRunDate)
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	var endDate *time.Time
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = &t
		}
	}

	rule, err := h.service.Update(c.Request.Context(), id, userID, service.UpdateRecurringRuleInput{
		CategoryID:  categoryID,
		AmountMinor: req.AmountMinor,
		Currency:    req.Currency,
		Frequency:   model.RecurringFrequency(req.Frequency),
		NextRunDate: nextRunDate,
		EndDate:     endDate,
	})
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusOK, rule)
}

func (h *RecurringRuleHandler) Delete(c *gin.Context) {
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
