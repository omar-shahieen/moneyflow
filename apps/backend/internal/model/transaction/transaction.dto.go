package transaction

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/model"
)

var validate = validator.New()

type GetTransactionRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r GetTransactionRequest) Validate() error {
	return validate.Struct(r)
}

type CreateTransactionRequest struct {
	CategoryID  string `json:"category_id" binding:"required,uuid"`
	AmountMinor int64  `json:"amount_minor" binding:"required"`
	Note        string `json:"note" binding:"omitempty,max=500"`
	ReceiptKey  string `json:"receipt_key" binding:"omitempty"`
	OccurredAt  string `json:"occurred_at" binding:"required"`
}

func (r CreateTransactionRequest) Validate() error {
	return validate.Struct(r)
}

func (r *CreateTransactionRequest) Normalize() {
	r.CategoryID = strings.TrimSpace(r.CategoryID)
	r.Note = strings.TrimSpace(r.Note)
	r.ReceiptKey = strings.TrimSpace(r.ReceiptKey)
	r.OccurredAt = strings.TrimSpace(r.OccurredAt)
}

type UpdateTransactionRequest struct {
	ID          uuid.UUID `uri:"id" binding:"required,uuid"`
	CategoryID  string    `json:"category_id" binding:"required,uuid"`
	AmountMinor int64     `json:"amount_minor" binding:"required"`
	Note        string    `json:"note" binding:"omitempty,max=500"`
	ReceiptKey  string    `json:"receipt_key" binding:"omitempty"`
	OccurredAt  string    `json:"occurred_at" binding:"required"`
}

func (r UpdateTransactionRequest) Validate() error {
	return validate.Struct(r)
}

func (r *UpdateTransactionRequest) Normalize() {
	r.CategoryID = strings.TrimSpace(r.CategoryID)
	r.Note = strings.TrimSpace(r.Note)
	r.ReceiptKey = strings.TrimSpace(r.ReceiptKey)
	r.OccurredAt = strings.TrimSpace(r.OccurredAt)
}

type GetReceiptUploadURLRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r GetReceiptUploadURLRequest) Validate() error { return validate.Struct(r) }

type DeleteTransactionRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r DeleteTransactionRequest) Validate() error {
	return validate.Struct(r)
}

type SummaryRequest struct {
	Month time.Time `form:"month" validate:"omitempty,datetime=2006-01"`
}

func (r SummaryRequest) Validate() error {
	return validate.Struct(r)
}

type TransactionSummary struct {
	TotalIncome  int64           `json:"total_income"`
	TotalExpense int64           `json:"total_expense"`
	ByCategory   []CategoryTotal `json:"by_category"`
}

type CategoryTotal struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Total        int64     `json:"total"`
	Type         string    `json:"type"`
}

type ListTransactionsRequest struct {
	model.PaginationRequest
	CategoryID    *uuid.UUID `form:"category_id" filter:"category_id,eq"`
	MinAmount     *int64     `form:"min_amount" filter:"amount_minor,gte"`
	MaxAmount     *int64     `form:"max_amount" filter:"amount_minor,lte"`
	From          *time.Time `form:"from" filter:"occurred_at,gte"`
	To            *time.Time `form:"to" filter:"occurred_at,lte"`
	Type          *string    `form:"type"`
	BillingStatus *string    `form:"billing_status" filter:"billing_status,eq"`
}

func (r ListTransactionsRequest) ApplyCustomFilters(args pgx.NamedArgs) string {
	if r.Type == nil {
		return ""
	}
	args["cat_type"] = *r.Type
	return " AND category_id IN (SELECT id FROM categories WHERE user_id = @user_id AND type = @cat_type)"
}
