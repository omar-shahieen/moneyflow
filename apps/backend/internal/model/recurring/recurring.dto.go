package recurring

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/model"
)

var validate = validator.New()

type GetRecurringRuleRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r GetRecurringRuleRequest) Validate() error {
	return validate.Struct(r)
}

type ListRecurringRulesRequest struct {
	model.PaginationRequest
	Frequency string `form:"frequency" filter:"frequency,eq" binding:"omitempty,oneof=weekly monthly"`
	Search    string `form:"search" binding:"omitempty"`
}

func (r *ListRecurringRulesRequest) Normalize() {
	r.PaginationRequest.Normalize()
	r.Frequency = strings.ToLower(strings.TrimSpace(r.Frequency))
	r.Search = strings.TrimSpace(r.Search)
}

func (r ListRecurringRulesRequest) ApplyCustomFilters(args pgx.NamedArgs) string {
	if r.Search == "" {
		return ""
	}
	args["search"] = r.Search
	return " AND category_id IN (SELECT id FROM categories WHERE user_id = @user_id AND name ILIKE '%' || @search || '%')"
}

type CreateRecurringRuleRequest struct {
	CategoryID  string `json:"category_id" binding:"required,uuid"`
	AmountMinor int64  `json:"amount_minor" binding:"required"`
	Frequency   string `json:"frequency" binding:"required,oneof=weekly monthly"`
	NextRunDate string `json:"next_run_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"omitempty"`
}

func (r CreateRecurringRuleRequest) Validate() error {
	return validate.Struct(r)
}

func (r *CreateRecurringRuleRequest) Normalize() {
	r.CategoryID = strings.TrimSpace(r.CategoryID)
	r.Frequency = strings.ToLower(strings.TrimSpace(r.Frequency))
	r.NextRunDate = strings.TrimSpace(r.NextRunDate)
	r.EndDate = strings.TrimSpace(r.EndDate)
}

type UpdateRecurringRuleRequest struct {
	ID          uuid.UUID `uri:"id" binding:"required,uuid"`
	CategoryID  string    `json:"category_id" binding:"required,uuid"`
	AmountMinor int64     `json:"amount_minor" binding:"required"`
	Frequency   string    `json:"frequency" binding:"required,oneof=weekly monthly"`
	NextRunDate string    `json:"next_run_date" binding:"required"`
	EndDate     string    `json:"end_date" binding:"omitempty"`
}

func (r UpdateRecurringRuleRequest) Validate() error {
	return validate.Struct(r)
}

func (r *UpdateRecurringRuleRequest) Normalize() {
	r.CategoryID = strings.TrimSpace(r.CategoryID)
	r.Frequency = strings.ToLower(strings.TrimSpace(r.Frequency))
	r.NextRunDate = strings.TrimSpace(r.NextRunDate)
	r.EndDate = strings.TrimSpace(r.EndDate)
}

type DeleteRecurringRuleRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r DeleteRecurringRuleRequest) Validate() error {
	return validate.Struct(r)
}
