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
	Status string `form:"status" filter:"status,eq" binding:"omitempty,oneof=pending processing ready failed"`
	Format string `form:"format" filter:"format,eq" binding:"omitempty,oneof=pdf csv"`
}

type CreateReportRequest struct {
	Format      string `json:"format" binding:"required,oneof=pdf csv"`
	PeriodStart string `json:"period_start" binding:"required"`
	PeriodEnd   string `json:"period_end" binding:"required"`
}

func (r CreateReportRequest) Validate() error {
	return validate.Struct(r)
}
