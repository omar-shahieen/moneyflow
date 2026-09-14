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
