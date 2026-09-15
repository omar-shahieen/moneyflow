package imports

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/model"
)

var validate = validator.New()

type GetImportRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r GetImportRequest) Validate() error {
	return validate.Struct(r)
}

type ImportCreateRequest struct {
	TotalRows int `json:"total_rows" binding:"required,min=1"`
}

func (r ImportCreateRequest) Validate() error {
	return validate.Struct(r)
}

type ImportCreateResponse struct {
	ImportID  string    `json:"import_id"`
	UploadURL string    `json:"upload_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ImportConfirmRequest struct{}

func (r ImportConfirmRequest) Validate() error {
	return nil
}

type ImportResponse struct {
	ID             string           `json:"id"`
	Status         string           `json:"status"`
	TotalRows      int              `json:"total_rows"`
	SuccessRows    int              `json:"success_rows"`
	FailedRows     []ImportRowError `json:"failed_rows,omitempty"`
	CreatedAt      string           `json:"created_at"`
	StorageKey     string           `json:"storage_key,omitempty"`
	ChecksumSHA256 string           `json:"checksum_sha256,omitempty"`
}

type ListImportsRequest struct {
	model.PaginationRequest
	Status *string `form:"status" filter:"status,eq" binding:"omitempty,oneof=pending queued processing completed failed"`
}

func (r ListImportsRequest) Validate() error {
	return validate.Struct(r)
}
