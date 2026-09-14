package transaction

import (
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
	Currency    string `json:"currency" binding:"omitempty,len=3"`
	Note        string `json:"note" binding:"omitempty,max=500"`
	ReceiptKey  string `json:"receipt_key" binding:"omitempty"`
	OccurredAt  string `json:"occurred_at" binding:"required"`
}

func (r CreateTransactionRequest) Validate() error {
	return validate.Struct(r)
}

type UpdateTransactionRequest struct {
	ID          uuid.UUID `uri:"id" binding:"required,uuid"`
	CategoryID  string    `json:"category_id" binding:"required,uuid"`
	AmountMinor int64     `json:"amount_minor" binding:"required"`
	Currency    string    `json:"currency" binding:"omitempty,len=3"`
	Note        string    `json:"note" binding:"omitempty,max=500"`
	ReceiptKey  string    `json:"receipt_key" binding:"omitempty"`
	OccurredAt  string    `json:"occurred_at" binding:"required"`
}

func (r UpdateTransactionRequest) Validate() error {
	return validate.Struct(r)
}

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
	ByCurrency   []CurrencyTotal `json:"by_currency"`
}

type CategoryTotal struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Total        int64     `json:"total"`
	Type         string    `json:"type"`
}

type CurrencyTotal struct {
	Currency     string `json:"currency"`
	TotalIncome  int64  `json:"total_income"`
	TotalExpense int64  `json:"total_expense"`
}

type ListTransactionsRequest struct {
	model.PaginationRequest
	CategoryID *uuid.UUID `form:"category_id" filter:"category_id,eq"`
	MinAmount  *int64     `form:"min_amount" filter:"amount_minor,gte"`
	MaxAmount  *int64     `form:"max_amount" filter:"amount_minor,lte"`
	From       *time.Time `form:"from" filter:"occurred_at,gte"`
	To         *time.Time `form:"to" filter:"occurred_at,lte"`
	Type       *string    `form:"type"`
}

func (r ListTransactionsRequest) ApplyCustomFilters(args pgx.NamedArgs) string {
	if r.Type == nil {
		return ""
	}
	args["cat_type"] = *r.Type
	return " AND category_id IN (SELECT id FROM categories WHERE user_id = @user_id AND type = @cat_type)"
}
