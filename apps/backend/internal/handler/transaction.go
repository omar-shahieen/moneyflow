package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
	"github.com/omar-shahieen/moneyflow/internal/domain/ports"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type TransactionHandler struct {
	service *service.TransactionService
}

func NewTransactionHandler(s *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: s}
}

type CreateTransactionRequest struct {
	CategoryID  string `json:"category_id" binding:"required"`
	AmountMinor int64  `json:"amount_minor" binding:"required"`
	Currency    string `json:"currency"`
	Note        string `json:"note"`
	OccurredAt  string `json:"occurred_at" binding:"required"`
}

type UpdateTransactionRequest struct {
	CategoryID  string `json:"category_id" binding:"required"`
	AmountMinor int64  `json:"amount_minor" binding:"required"`
	Currency    string `json:"currency"`
	Note        string `json:"note"`
	OccurredAt  string `json:"occurred_at" binding:"required"`
}

type TransactionListResponse struct {
	Data       []model.Transaction `json:"data"`
	Pagination struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
		Total int `json:"total"`
	} `json:"pagination"`
}

func (h *TransactionHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	filter := ports.TransactionFilter{
		Page:  1,
		Limit: 20,
	}

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			filter.Page = v
		}
	}
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			filter.Limit = v
		}
	}
	if catID := c.Query("category_id"); catID != "" {
		if id, err := uuid.Parse(catID); err == nil {
			filter.CategoryID = &id
		}
	}
	if t := c.Query("type"); t != "" {
		filter.Type = &t
	}
	if from := c.Query("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.From = &t
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filter.To = &t
		}
	}

	transactions, total, err := h.service.List(c.Request.Context(), userID, filter)
	if err != nil {
		RespondError(c, err)
		return
	}

	if transactions == nil {
		transactions = []model.Transaction{}
	}

	response := TransactionListResponse{
		Data: transactions,
	}
	response.Pagination.Page = filter.Page
	response.Pagination.Limit = filter.Limit
	response.Pagination.Total = total

	RespondJSON(c, http.StatusOK, response)
}

func (h *TransactionHandler) GetByID(c *gin.Context) {
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

	transaction, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusOK, transaction)
}

func (h *TransactionHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	var req CreateTransactionRequest
	if err := BindAndValidate(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	occurredAt, err := time.Parse(time.RFC3339, req.OccurredAt)
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	transaction, err := h.service.Create(c.Request.Context(), userID, service.CreateTransactionInput{
		CategoryID:  categoryID,
		AmountMinor: req.AmountMinor,
		Currency:    req.Currency,
		Note:        req.Note,
		OccurredAt:  occurredAt,
	})
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusCreated, transaction)
}

func (h *TransactionHandler) Update(c *gin.Context) {
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

	var req UpdateTransactionRequest
	if err := BindAndValidate(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	occurredAt, err := time.Parse(time.RFC3339, req.OccurredAt)
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	transaction, err := h.service.Update(c.Request.Context(), id, userID, service.UpdateTransactionInput{
		CategoryID:  categoryID,
		AmountMinor: req.AmountMinor,
		Currency:    req.Currency,
		Note:        req.Note,
		OccurredAt:  occurredAt,
	})
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusOK, transaction)
}

func (h *TransactionHandler) Delete(c *gin.Context) {
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

func (h *TransactionHandler) Summary(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	monthStr := c.Query("month")
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}

	month, err := time.Parse("2006-01", monthStr)
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	summary, err := h.service.Summary(c.Request.Context(), userID, month)
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusOK, summary)
}
