package model

import (
	"time"

	"github.com/google/uuid"
)

type ReportFormat string

const (
	ReportFormatPDF ReportFormat = "pdf"
	ReportFormatCSV ReportFormat = "csv"
)

type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusProcessing ReportStatus = "processing"
	ReportStatusReady      ReportStatus = "ready"
	ReportStatusFailed     ReportStatus = "failed"
)

type Report struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	UserID      string       `json:"user_id" db:"user_id"`
	Format      ReportFormat `json:"format" db:"format"`
	PeriodStart time.Time    `json:"period_start" db:"period_start"`
	PeriodEnd   time.Time    `json:"period_end" db:"period_end"`
	Status      ReportStatus `json:"status" db:"status"`
	StorageKey  string       `json:"storage_key,omitempty" db:"storage_key"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty" db:"completed_at"`
}
