package imports

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ImportStatus string

const (
	ImportStatusPending    ImportStatus = "pending"
	ImportStatusProcessing ImportStatus = "processing"
	ImportStatusCompleted  ImportStatus = "completed"
	ImportStatusFailed     ImportStatus = "failed"
)

type ImportRowError struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

type Import struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	UserID      string          `json:"user_id" db:"user_id"`
	Status      ImportStatus    `json:"status" db:"status"`
	TotalRows   int             `json:"total_rows" db:"total_rows"`
	SuccessRows int             `json:"success_rows" db:"success_rows"`
	FailedRows  json.RawMessage `json:"failed_rows" db:"failed_rows"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}
