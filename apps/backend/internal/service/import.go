package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
	"github.com/omar-shahieen/moneyflow/internal/domain/ports"
)

type ImportService struct {
	repo            ports.ImportRepository
	transactionRepo ports.TransactionRepository
	categoryRepo    ports.CategoryRepository
	planGuard       ports.PlanGuard
	pool            *pgxpool.Pool
}

func NewImportService(
	repo ports.ImportRepository,
	transactionRepo ports.TransactionRepository,
	categoryRepo ports.CategoryRepository,
	planGuard ports.PlanGuard,
	pool *pgxpool.Pool,
) *ImportService {
	return &ImportService{
		repo:            repo,
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
		planGuard:       planGuard,
		pool:            pool,
	}
}

func (s *ImportService) GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.Import, error) {
	return s.repo.GetByID(ctx, id, userID)
}

type CreateImportInput struct {
	TotalRows int
}

func (s *ImportService) Create(ctx context.Context, userID string, input CreateImportInput) (*model.Import, error) {
	if s.planGuard != nil {
		if err := s.planGuard.CheckCSVImportLimit(ctx, userID, input.TotalRows); err != nil {
			return nil, err
		}
	}

	imp := &model.Import{
		ID:          uuid.New(),
		UserID:      userID,
		Status:      model.ImportStatusPending,
		TotalRows:   input.TotalRows,
		SuccessRows: 0,
		FailedRows:  []byte("[]"),
	}

	if err := s.repo.Create(ctx, imp); err != nil {
		return nil, err
	}
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
	imp, err := s.repo.GetByID(ctx, importID, userID)
	if err != nil {
		return err
	}

	imp.Status = model.ImportStatusProcessing
	if err := s.repo.Update(ctx, imp); err != nil {
		return err
	}

	var failedRows []model.ImportRowError
	successCount := 0

	for i, row := range rows {
		if err := s.processRow(ctx, userID, row); err != nil {
			failedRows = append(failedRows, model.ImportRowError{
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

	if len(failedRows) > 0 {
		imp.Status = model.ImportStatusCompleted
	} else {
		imp.Status = model.ImportStatusCompleted
	}

	return s.repo.Update(ctx, imp)
}

func (s *ImportService) processRow(ctx context.Context, userID string, row CSVRow) error {
	if row.CategoryName == "" {
		return fmt.Errorf("category name is required")
	}

	if row.Amount == 0 {
		return fmt.Errorf("amount must be non-zero")
	}

	var categoryID uuid.UUID
	err := s.pool.QueryRow(ctx,
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

	t := &model.Transaction{
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
