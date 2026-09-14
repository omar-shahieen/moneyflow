package report

import (
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
}

type CreateReportRequest struct {
	Format      string `json:"format" binding:"required,oneof=pdf csv"`
	PeriodStart string `json:"period_start" binding:"required"`
	PeriodEnd   string `json:"period_end" binding:"required"`
}

func (r CreateReportRequest) Validate() error {
	return validate.Struct(r)
}

type DeleteReportRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r DeleteReportRequest) Validate() error {
	return validate.Struct(r)
}

type ReportResponse struct {
	Report
}
