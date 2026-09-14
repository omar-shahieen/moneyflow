package transaction

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/model"
)

var validate = validator.New()

type GetTransactionRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r GetTransactionRequest) Validate() error {
	return validate.Struct(r)
}

type ListTransactionsRequest struct {
	model.PaginationRequest
	CategoryID string     `form:"category_id" binding:"omitempty,uuid"`
	Type       string     `form:"type" binding:"omitempty,oneof=income expense"`
	StartDate  *time.Time `form:"start_date" binding:"omitempty"`
	EndDate    *time.Time `form:"end_date" binding:"omitempty"`
	MinAmount  *int64     `form:"min_amount" binding:"omitempty"`
	MaxAmount  *int64     `form:"max_amount" binding:"omitempty"`
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

type TransactionResponse struct {
	Transaction
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
type TransactionFilters struct {
	CategoryID *uuid.UUID
	Type       *string
	From       *time.Time
	To         *time.Time
	MinAmount  *int64
	MaxAmount  *int64
}
