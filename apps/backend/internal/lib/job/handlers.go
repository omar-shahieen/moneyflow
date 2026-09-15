package job

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/omar-shahieen/moneyflow/internal/config"
	"github.com/omar-shahieen/moneyflow/internal/lib/email"
	"github.com/rs/zerolog"
)

var emailClient *email.Client

// ImportProcessor is the interface the import service must satisfy for the job handler.
type ImportProcessor interface {
	ProcessImportFromStorage(ctx context.Context, importID, userID, storageKey, checksumSHA256 string) error
	IsAlreadyProcessed(ctx context.Context, userID, checksum string) (bool, error)
}

func (j *JobService) InitHandlers(config *config.Config, logger *zerolog.Logger) {
	emailClient = email.NewClient(config, logger)
}

func (j *JobService) SetImportService(svc ImportProcessor) {
	j.importService = svc
}

func (j *JobService) handleWelcomeEmailTask(ctx context.Context, t *asynq.Task) error {
	var p WelcomeEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal welcome email payload: %w", err)
	}

	j.logger.Info().
		Str("type", "welcome").
		Str("to", p.To).
		Msg("Processing welcome email task")

	err := emailClient.SendWelcomeEmail(
		p.To,
		p.FirstName,
	)
	if err != nil {
		j.logger.Error().
			Str("type", "welcome").
			Str("to", p.To).
			Err(err).
			Msg("Failed to send welcome email")
		return err
	}

	j.logger.Info().
		Str("type", "welcome").
		Str("to", p.To).
		Msg("Successfully sent welcome email")
	return nil
}

func (j *JobService) handleProcessImport(ctx context.Context, t *asynq.Task) error {
	var p ProcessImportPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal process import payload: %w", err)
	}

	j.logger.Info().
		Str("type", "import:process").
		Str("import_id", p.ImportID).
		Str("user_id", p.UserID).
		Str("storage_key", p.StorageKey).
		Msg("Processing import task")

	if j.importService == nil {
		return fmt.Errorf("import service not initialized")
	}

	// Idempotency check: skip if this checksum has already been processed
	if p.ChecksumSHA256 != "" {
		processed, err := j.importService.IsAlreadyProcessed(ctx, p.UserID, p.ChecksumSHA256)
		if err != nil {
			j.logger.Error().
				Err(err).
				Str("import_id", p.ImportID).
				Msg("Failed to check import idempotency")
			return fmt.Errorf("failed to check idempotency: %w", err)
		}
		if processed {
			j.logger.Info().
				Str("import_id", p.ImportID).
				Str("checksum", p.ChecksumSHA256).
				Msg("Import already processed, skipping (idempotent)")
			return nil
		}
	}

	if err := j.importService.ProcessImportFromStorage(ctx, p.ImportID, p.UserID, p.StorageKey, p.ChecksumSHA256); err != nil {
		j.logger.Error().
			Err(err).
			Str("import_id", p.ImportID).
			Msg("Failed to process import")
		return err
	}

	j.logger.Info().
		Str("type", "import:process").
		Str("import_id", p.ImportID).
		Msg("Import processed successfully")
	return nil
}
