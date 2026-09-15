package report

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/model"
)

var validate = validator.New()

type GetReportRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r GetReportRequest) Validate() error {
	return validate.Struct(r)
}

type ListReportsRequest struct {
	model.PaginationRequest
	Status string `form:"status" filter:"status,eq" binding:"omitempty,oneof=pending processing ready failed"`
	Format string `form:"format" filter:"format,eq" binding:"omitempty,oneof=pdf csv"`
}

func (r *ListReportsRequest) Normalize() {
	r.PaginationRequest.Normalize()
	r.Status = strings.ToLower(strings.TrimSpace(r.Status))
	r.Format = strings.ToLower(strings.TrimSpace(r.Format))
}

type CreateReportRequest struct {
	Format      string `json:"format" binding:"required,oneof=pdf csv"`
	PeriodStart string `json:"period_start" binding:"required"`
	PeriodEnd   string `json:"period_end" binding:"required"`
}

func (r CreateReportRequest) Validate() error {
	return validate.Struct(r)
}

func (r *CreateReportRequest) Normalize() {
	r.Format = strings.ToLower(strings.TrimSpace(r.Format))
	r.PeriodStart = strings.TrimSpace(r.PeriodStart)
	r.PeriodEnd = strings.TrimSpace(r.PeriodEnd)
}
