package job

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockImportProcessor struct {
	processedCalls  []string
	alreadyProcessed bool
}

func (m *mockImportProcessor) ProcessImportFromStorage(ctx context.Context, importID, userID, storageKey, checksumSHA256 string) error {
	m.processedCalls = append(m.processedCalls, importID)
	return nil
}

func (m *mockImportProcessor) IsAlreadyProcessed(ctx context.Context, userID, checksum string) (bool, error) {
	return m.alreadyProcessed, nil
}

func TestHandleProcessImport_IdempotencySkip(t *testing.T) {
	mock := &mockImportProcessor{alreadyProcessed: true}
	logger := zerolog.Nop()

	j := &JobService{
		importService: mock,
		logger:       &logger,
	}

	payload := ProcessImportPayload{
		ImportID:       "import-123",
		UserID:         "user-456",
		StorageKey:     "payments/local/user-456/2024/01/15/import-123.csv",
		ChecksumSHA256: "abc123hash",
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TaskProcessImport, payloadBytes)

	err = j.handleProcessImport(context.Background(), task)
	assert.NoError(t, err)
	assert.Empty(t, mock.processedCalls, "should not have called ProcessImportFromStorage")
}

func TestHandleProcessImport_ProcessesWhenNotAlreadyDone(t *testing.T) {
	mock := &mockImportProcessor{alreadyProcessed: false}
	logger := zerolog.Nop()

	j := &JobService{
		importService: mock,
		logger:       &logger,
	}

	payload := ProcessImportPayload{
		ImportID:       "import-789",
		UserID:         "user-456",
		StorageKey:     "payments/local/user-456/2024/01/15/import-789.csv",
		ChecksumSHA256: "def456hash",
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TaskProcessImport, payloadBytes)

	err = j.handleProcessImport(context.Background(), task)
	assert.NoError(t, err)
	assert.Equal(t, []string{"import-789"}, mock.processedCalls)
}

func TestHandleProcessImport_NoServiceReturnsError(t *testing.T) {
	logger := zerolog.Nop()
	j := &JobService{
		logger: &logger,
	}

	payload := ProcessImportPayload{
		ImportID:   "import-123",
		UserID:     "user-456",
		StorageKey: "some/key.csv",
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TaskProcessImport, payloadBytes)

	err = j.handleProcessImport(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "import service not initialized")
}

func TestHandleProcessImport_EmptyChecksumStillProcesses(t *testing.T) {
	mock := &mockImportProcessor{alreadyProcessed: false}
	logger := zerolog.Nop()

	j := &JobService{
		importService: mock,
		logger:       &logger,
	}

	payload := ProcessImportPayload{
		ImportID:   "import-no-checksum",
		UserID:     "user-456",
		StorageKey: "some/key.csv",
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TaskProcessImport, payloadBytes)

	err = j.handleProcessImport(context.Background(), task)
	assert.NoError(t, err)
	assert.Equal(t, []string{"import-no-checksum"}, mock.processedCalls)
}
