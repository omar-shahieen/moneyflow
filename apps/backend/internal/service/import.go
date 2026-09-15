package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/lib/job"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/imports"
	"github.com/omar-shahieen/moneyflow/internal/model/transaction"
	"github.com/omar-shahieen/moneyflow/internal/ports"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type ImportService struct {
	server          *server.Server
	importRepo      *repository.ImportRepo
	transactionRepo *repository.TransactionRepo
	categoryRepo    *repository.CategoryRepo
	storage         ports.Storage
}

func NewImportService(server *server.Server, importRepo *repository.ImportRepo, transactionRepo *repository.TransactionRepo, categoryRepo *repository.CategoryRepo, storage ports.Storage) *ImportService {
	return &ImportService{
		server:          server,
		importRepo:      importRepo,
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
		storage:         storage,
	}
}

func (s *ImportService) GetImportByID(ctx context.Context, userID string, importID uuid.UUID) (*imports.Import, error) {
	imp, err := s.importRepo.GetByID(ctx, importID, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch import by ID")
		return nil, err
	}
	return imp, nil
}

func (s *ImportService) GetImports(ctx context.Context, userID string, query *imports.ListImportsRequest) (*model.PaginatedResponse[imports.Import], error) {
	query.Normalize()
	result, err := s.importRepo.List(ctx, userID, query)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch imports")
		return nil, err
	}
	return result, nil
}

func (s *ImportService) GenerateStorageKey(userID string, importID uuid.UUID) string {
	now := time.Now()
	env := s.server.Config.Primary.Env
	return fmt.Sprintf("payments/%s/%s/%04d/%02d/%02d/%s.csv",
		env, userID, now.Year(), now.Month(), now.Day(), importID.String())
}

func (s *ImportService) CreateImport(ctx context.Context, userID string, totalRows int) (*imports.Import, error) {
	imp := &imports.Import{
		ID:          uuid.New(),
		UserID:      userID,
		Status:      imports.ImportStatusPending,
		TotalRows:   totalRows,
		SuccessRows: 0,
		FailedRows:  []byte("[]"),
	}

	imp.StorageKey = s.GenerateStorageKey(userID, imp.ID)

	if err := s.importRepo.Create(ctx, imp); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to create import")
		return nil, err
	}

	s.server.Logger.Info().
		Str("event", "import_created").
		Str("import_id", imp.ID.String()).
		Str("storage_key", imp.StorageKey).
		Msg("Import created successfully")

	return imp, nil
}

func (s *ImportService) ConfirmUpload(ctx context.Context, importID uuid.UUID, userID string) (*imports.Import, error) {
	imp, err := s.importRepo.GetByID(ctx, importID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get import: %w", err)
	}

	if imp.Status != imports.ImportStatusPending {
		return nil, fmt.Errorf("import is not in pending status: %s", imp.Status)
	}

	info, err := s.storage.Head(ctx, imp.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("upload verification failed: %w", err)
	}

	if info.Size == 0 {
		return nil, fmt.Errorf("uploaded file is empty")
	}

	if info.ChecksumSHA256 != "" {
		imp.ChecksumSHA256 = info.ChecksumSHA256
	}

	if imp.ChecksumSHA256 != "" {
		existing, err := s.importRepo.GetByChecksum(ctx, userID, imp.ChecksumSHA256)
		if err != nil {
			return nil, fmt.Errorf("failed to check idempotency: %w", err)
		}
		if existing != nil && existing.ID != imp.ID {
			imp.Status = imports.ImportStatusFailed
			failedJSON, _ := json.Marshal([]imports.ImportRowError{
				{Row: 0, Reason: fmt.Sprintf("duplicate file: already processed as import %s", existing.ID.String())},
			})
			imp.FailedRows = failedJSON
			_ = s.importRepo.Update(ctx, imp)
			return nil, fmt.Errorf("duplicate file: already processed as import %s", existing.ID.String())
		}
	}

	imp.Status = imports.ImportStatusQueued
	if err := s.importRepo.Update(ctx, imp); err != nil {
		return nil, fmt.Errorf("failed to update import status: %w", err)
	}

	task, err := job.NewProcessImportTask(imp.ID.String(), userID, imp.StorageKey, imp.ChecksumSHA256)
	if err != nil {
		return nil, fmt.Errorf("failed to create import task: %w", err)
	}

	if _, err := s.server.Job.Client.Enqueue(task); err != nil {
		return nil, fmt.Errorf("failed to enqueue import task: %w", err)
	}

	s.server.Logger.Info().
		Str("event", "import_confirmed").
		Str("import_id", imp.ID.String()).
		Str("checksum", imp.ChecksumSHA256).
		Msg("Import confirmed and queued for processing")

	return imp, nil
}

func (s *ImportService) IsAlreadyProcessed(ctx context.Context, userID string, checksum string) (bool, error) {
	existing, err := s.importRepo.GetByChecksum(ctx, userID, checksum)
	if err != nil {
		return false, err
	}
	if existing == nil {
		return false, nil
	}
	return existing.Status == imports.ImportStatusCompleted || existing.Status == imports.ImportStatusFailed, nil
}

func (s *ImportService) ProcessImportFromStorage(ctx context.Context, importID string, userID string, storageKey string, checksumSHA256 string) error {
	id, err := uuid.Parse(importID)
	if err != nil {
		return fmt.Errorf("invalid import ID: %w", err)
	}

	imp, err := s.importRepo.GetByID(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("failed to get import: %w", err)
	}

	imp.Status = imports.ImportStatusProcessing
	if err := s.importRepo.Update(ctx, imp); err != nil {
		return fmt.Errorf("failed to update import status: %w", err)
	}

	info, err := s.storage.Head(ctx, storageKey)
	if err != nil {
		s.markFailed(ctx, imp, fmt.Sprintf("storage head failed: %v", err))
		return fmt.Errorf("storage head failed: %w", err)
	}

	if checksumSHA256 != "" && info.ChecksumSHA256 != "" && info.ChecksumSHA256 != checksumSHA256 {
		s.markFailed(ctx, imp, fmt.Sprintf("checksum mismatch: expected %s, got %s", checksumSHA256, info.ChecksumSHA256))
		return fmt.Errorf("checksum mismatch: expected %s, got %s", checksumSHA256, info.ChecksumSHA256)
	}

	data, err := s.storage.Download(ctx, storageKey)
	if err != nil {
		s.markFailed(ctx, imp, fmt.Sprintf("download failed: %v", err))
		return fmt.Errorf("download failed: %w", err)
	}

	if err := ValidateSchema(data, imp.TotalRows); err != nil {
		s.markFailed(ctx, imp, fmt.Sprintf("schema validation failed: %v", err))
		return fmt.Errorf("schema validation failed: %w", err)
	}

	records, err := parseCSV(data)
	if err != nil {
		s.markFailed(ctx, imp, fmt.Sprintf("CSV parse error: %v", err))
		return fmt.Errorf("CSV parse error: %w", err)
	}

	var failedRows []imports.ImportRowError
	successCount := 0

	for i, record := range records[1:] {
		if len(record) < 2 {
			continue
		}

		amount, _ := strconv.ParseInt(record[1], 10, 64)
		note := ""
		if len(record) > 2 {
			note = record[2]
		}
		date := ""
		if len(record) > 3 {
			date = record[3]
		}

		row := CSVRow{
			CategoryName: record[0],
			Amount:       amount,
			Note:         note,
			Date:         date,
		}

		if err := s.processRow(ctx, userID, row); err != nil {
			failedRows = append(failedRows, imports.ImportRowError{
				Row:    i + 2,
				Reason: err.Error(),
			})
		} else {
			successCount++
		}
	}

	failedJSON, _ := json.Marshal(failedRows)
	imp.SuccessRows = successCount
	imp.FailedRows = failedJSON
	imp.TotalRows = len(records) - 1
	imp.Status = imports.ImportStatusCompleted

	if err := s.importRepo.Update(ctx, imp); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to update import status")
		return err
	}

	s.server.Logger.Info().
		Str("event", "import_completed").
		Str("import_id", id.String()).
		Int("success", successCount).
		Int("failed", len(failedRows)).
		Msg("Import processing completed")

	return nil
}

func (s *ImportService) markFailed(ctx context.Context, imp *imports.Import, reason string) {
	imp.Status = imports.ImportStatusFailed
	failedJSON, _ := json.Marshal([]imports.ImportRowError{
		{Row: 0, Reason: reason},
	})
	imp.FailedRows = failedJSON
	if err := s.importRepo.Update(ctx, imp); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to mark import as failed")
	}
}

func ValidateSchema(data []byte, expectedRows int) error {
	records, err := parseCSV(data)
	if err != nil {
		return fmt.Errorf("cannot parse CSV: %w", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("CSV must have a header row and at least one data row")
	}

	header := records[0]
	requiredCols := map[string]bool{
		"CategoryName": false,
		"Amount":       false,
	}

	for _, col := range header {
		trimmed := strings.TrimSpace(col)
		if _, ok := requiredCols[trimmed]; ok {
			requiredCols[trimmed] = true
		}
	}

	for col, found := range requiredCols {
		if !found {
			return fmt.Errorf("missing required column: %s", col)
		}
	}

	dataRows := len(records) - 1
	if expectedRows > 0 && dataRows != expectedRows {
		return fmt.Errorf("row count mismatch: expected %d rows, got %d", expectedRows, dataRows)
	}

	return nil
}

type CSVRow struct {
	CategoryName string
	Amount       int64
	Note         string
	Date         string
}

func (s *ImportService) ProcessImport(ctx context.Context, importID uuid.UUID, userID string, rows []CSVRow) error {
	imp, err := s.importRepo.GetByID(ctx, importID, userID)
	if err != nil {
		return err
	}

	imp.Status = imports.ImportStatusProcessing
	if err := s.importRepo.Update(ctx, imp); err != nil {
		return err
	}

	var failedRows []imports.ImportRowError
	successCount := 0

	for i, row := range rows {
		if err := s.processRow(ctx, userID, row); err != nil {
			failedRows = append(failedRows, imports.ImportRowError{
				Row:    i + 1,
				Reason: err.Error(),
			})
		} else {
			successCount++
		}
	}

	failedJSON, _ := json.Marshal(failedRows)
	imp.SuccessRows = successCount
	imp.FailedRows = failedJSON
	imp.TotalRows = len(rows)
	imp.Status = imports.ImportStatusCompleted

	if err := s.importRepo.Update(ctx, imp); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to update import status")
		return err
	}

	s.server.Logger.Info().
		Str("event", "import_completed").
		Str("import_id", importID.String()).
		Int("success", successCount).
		Int("failed", len(failedRows)).
		Msg("Import processing completed")

	return nil
}

func (s *ImportService) processRow(ctx context.Context, userID string, row CSVRow) error {
	if row.CategoryName == "" {
		return fmt.Errorf("category name is required")
	}

	if row.Amount == 0 {
		return fmt.Errorf("amount must be non-zero")
	}

	pool := s.server.DB.Pool
	var categoryID uuid.UUID
	err := pool.QueryRow(ctx,
		`SELECT id FROM categories WHERE user_id = $1 AND name = $2 LIMIT 1`,
		userID, row.CategoryName,
	).Scan(&categoryID)
	if err != nil {
		return fmt.Errorf("unknown category: %s", row.CategoryName)
	}

	occurredAt, err := time.Parse("2006-01-02", row.Date)
	if err != nil {
		occurredAt = time.Now()
	}

	t := &transaction.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		CategoryID:  categoryID,
		AmountMinor: row.Amount,
		Note:        row.Note,
		OccurredAt:  occurredAt,
	}

	return s.transactionRepo.Create(ctx, t)
}

func parseCSV(data []byte) ([][]string, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	return reader.ReadAll()
}

func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
