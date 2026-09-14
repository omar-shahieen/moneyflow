package imports

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate = validator.New()

type GetImportRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r GetImportRequest) Validate() error {
	return validate.Struct(r)
}

type ImportResponse struct {
	ID          string           `json:"id"`
	Status      string           `json:"status"`
	TotalRows   int              `json:"total_rows"`
	SuccessRows int              `json:"success_rows"`
	FailedRows  []ImportRowError `json:"failed_rows,omitempty"`
	CreatedAt   string           `json:"created_at"`
}

type ImportCreateRequest struct{}

func (r ImportCreateRequest) Validate() error {
	return validate.Struct(r)
}
