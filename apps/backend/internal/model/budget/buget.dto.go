package budget

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/model"
)

var validate = validator.New()

type EmptyRequest struct{}

func (r EmptyRequest) Validate() error {
	return validate.Struct(r)
}

type GetBudgetRequest struct {
	ID uuid.UUID `uri:"id" binding:"required"`
}

func (r GetBudgetRequest) Validate() error {
	return validate.Struct(r)
}

type ListBudgetsRequest struct {
	model.PaginationRequest
}

type CreateBudgetRequest struct {
	CategoryID        uuid.UUID `json:"category_id" binding:"required,uuid"`
	MonthlyLimitMinor int64     `json:"monthly_limit_minor" binding:"required,gt=0"`
	Currency          string    `json:"currency" binding:"omitempty,len=3"`
}

func (r CreateBudgetRequest) Validate() error {
	return validate.Struct(r)
}

type UpdateBudgetRequest struct {
	ID                uuid.UUID `uri:"id" binding:"required,uuid"`
	CategoryID        uuid.UUID `json:"category_id" binding:"required,uuid"`
	MonthlyLimitMinor int64     `json:"monthly_limit_minor" binding:"required,gt=0"`
	Currency          string    `json:"currency" binding:"omitempty,len=3"`
}

func (r UpdateBudgetRequest) Validate() error {
	return validate.Struct(r)
}

type AddMemberRequest struct {
	ID     uuid.UUID `uri:"id" binding:"required,uuid"`
	UserID string    `json:"user_id" binding:"required"`
}

func (r AddMemberRequest) Validate() error {
	return validate.Struct(r)
}

type RemoveMemberRequest struct {
	ID     uuid.UUID `uri:"id" binding:"required,uuid"`
	UserID string    `uri:"userId" binding:"required"`
}

func (r RemoveMemberRequest) Validate() error {
	return validate.Struct(r)
}

type BudgetResponse struct {
	BudgetWithMembers
	Usage        int64   `json:"usage"`
	UsagePercent float64 `json:"usage_percent"`
	Exceeded     bool    `json:"exceeded"`
}
