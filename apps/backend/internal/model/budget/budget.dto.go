package budget

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/model"
)

var validate = validator.New()

type GetBudgetRequest struct {
	ID uuid.UUID `uri:"id" binding:"required"`
}

func (r GetBudgetRequest) Validate() error {
	return validate.Struct(r)
}

type ListBudgetsRequest struct {
	model.PaginationRequest
	Search string `form:"search" binding:"omitempty"`
}

func (r *ListBudgetsRequest) Normalize() {
	r.PaginationRequest.Normalize()
	r.Search = strings.TrimSpace(r.Search)
}

func (r ListBudgetsRequest) ApplyCustomFilters(args pgx.NamedArgs) string {
	if r.Search == "" {
		return ""
	}
	args["search"] = r.Search
	return " AND b.category_id IN (SELECT id FROM categories WHERE user_id = @user_id AND name ILIKE '%' || @search || '%')"
}

type CreateBudgetRequest struct {
	CategoryID        uuid.UUID `json:"category_id" binding:"required,uuid"`
	MonthlyLimitMinor int64     `json:"monthly_limit_minor" binding:"required,gt=0"`
}

func (r CreateBudgetRequest) Validate() error {
	return validate.Struct(r)
}

type UpdateBudgetRequest struct {
	ID                uuid.UUID `uri:"id" binding:"required,uuid"`
	CategoryID        uuid.UUID `json:"category_id" binding:"required,uuid"`
	MonthlyLimitMinor int64     `json:"monthly_limit_minor" binding:"required,gt=0"`
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

func (r *AddMemberRequest) Normalize() { r.UserID = strings.TrimSpace(r.UserID) }

type RemoveMemberRequest struct {
	ID     uuid.UUID `uri:"id" binding:"required,uuid"`
	UserID string    `uri:"userId" binding:"required"`
}

func (r RemoveMemberRequest) Validate() error {
	return validate.Struct(r)
}

func (r *RemoveMemberRequest) Normalize() { r.UserID = strings.TrimSpace(r.UserID) }

type BudgetResponse struct {
	BudgetWithMembers
	Usage        int64   `json:"usage"`
	UsagePercent float64 `json:"usage_percent"`
	Exceeded     bool    `json:"exceeded"`
}
