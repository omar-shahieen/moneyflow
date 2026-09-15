package job

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const TaskProcessImport = "import:process"

type ProcessImportPayload struct {
	ImportID      string `json:"import_id"`
	UserID        string `json:"user_id"`
	StorageKey    string `json:"storage_key"`
	ChecksumSHA256 string `json:"checksum_sha256"`
}

func NewProcessImportTask(importID, userID, storageKey, checksumSHA256 string) (*asynq.Task, error) {
	payload, err := json.Marshal(ProcessImportPayload{
		ImportID:      importID,
		UserID:        userID,
		StorageKey:    storageKey,
		ChecksumSHA256: checksumSHA256,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TaskProcessImport, payload,
		asynq.MaxRetry(3),
		asynq.Queue("default"),
		asynq.Timeout(5*time.Minute)), nil
}
