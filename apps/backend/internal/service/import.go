package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/model/imports"
	"github.com/omar-shahieen/moneyflow/internal/model/transaction"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type ImportService struct {
	server          *server.Server
	importRepo      *repository.ImportRepo
	transactionRepo *repository.TransactionRepo
	categoryRepo    *repository.CategoryRepo
}

func NewImportService(server *server.Server, importRepo *repository.ImportRepo, transactionRepo *repository.TransactionRepo, categoryRepo *repository.CategoryRepo) *ImportService {
	return &ImportService{
		server:          server,
		importRepo:      importRepo,
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
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

func (s *ImportService) CreateImport(ctx context.Context, userID string, totalRows int) (*imports.Import, error) {
	imp := &imports.Import{
		ID:          uuid.New(),
		UserID:      userID,
		Status:      imports.ImportStatusPending,
		TotalRows:   totalRows,
		SuccessRows: 0,
		FailedRows:  []byte("[]"),
	}

	if err := s.importRepo.Create(ctx, imp); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to create import")
		return nil, err
	}

	s.server.Logger.Info().
		Str("event", "import_created").
		Str("import_id", imp.ID.String()).
		Msg("Import created successfully")

	return imp, nil
}

type CSVRow struct {
	CategoryName string
	Amount       int64
	Currency     string
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
		Currency:    row.Currency,
		Note:        row.Note,
		OccurredAt:  occurredAt,
	}

	return s.transactionRepo.Create(ctx, t)
}
