package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/domain"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
	"github.com/omar-shahieen/moneyflow/internal/domain/ports"
)

type ReportService struct {
	repo            ports.ReportRepository
	transactionRepo ports.TransactionRepository
	storage         ports.Storage
	planGuard       ports.PlanGuard
}

func NewReportService(
	repo ports.ReportRepository,
	transactionRepo ports.TransactionRepository,
	storage ports.Storage,
	planGuard ports.PlanGuard,
) *ReportService {
	return &ReportService{
		repo:            repo,
		transactionRepo: transactionRepo,
		storage:         storage,
		planGuard:       planGuard,
	}
}

func (s *ReportService) GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.Report, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *ReportService) List(ctx context.Context, userID string) ([]model.Report, error) {
	return s.repo.List(ctx, userID, 30)
}

type CreateReportInput struct {
	Format      model.ReportFormat
	PeriodStart time.Time
	PeriodEnd   time.Time
}

func (s *ReportService) Create(ctx context.Context, userID string, input CreateReportInput) (*model.Report, error) {
	if input.Format != model.ReportFormatPDF && input.Format != model.ReportFormatCSV {
		return nil, domain.NewDomainError(domain.ErrValidation, "VALIDATION_FAILED", 422, "format must be 'pdf' or 'csv'")
	}

	if s.planGuard != nil {
		if err := s.planGuard.CheckReportLimit(ctx, userID); err != nil {
			return nil, err
		}
	}

	report := &model.Report{
		ID:          uuid.New(),
		UserID:      userID,
		Format:      input.Format,
		PeriodStart: input.PeriodStart,
		PeriodEnd:   input.PeriodEnd,
		Status:      model.ReportStatusPending,
	}

	if err := s.repo.Create(ctx, report); err != nil {
		return nil, err
	}

	go s.generateReport(report.ID, userID)

	return report, nil
}

func (s *ReportService) generateReport(reportID uuid.UUID, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	report, err := s.repo.GetByID(ctx, reportID, userID)
	if err != nil {
		return
	}

	report.Status = model.ReportStatusProcessing
	_ = s.repo.Update(ctx, report)

	filter := ports.TransactionFilter{
		From:  &report.PeriodStart,
		To:    &report.PeriodEnd,
		Page:  1,
		Limit: 10000,
	}

	transactions, _, err := s.transactionRepo.List(ctx, userID, filter)
	if err != nil {
		report.Status = model.ReportStatusFailed
		_ = s.repo.Update(ctx, report)
		return
	}

	storageKey := fmt.Sprintf("reports/%s/%s.csv", userID, reportID.String())

	if report.Format == model.ReportFormatCSV {
		csvData, err := generateCSV(transactions)
		if err != nil {
			report.Status = model.ReportStatusFailed
			_ = s.repo.Update(ctx, report)
			return
		}

		uploadURL, err := s.storage.GenerateUploadURL(storageKey, "text/csv", 24*time.Hour)
		if err != nil {
			report.Status = model.ReportStatusFailed
			_ = s.repo.Update(ctx, report)
			return
		}

		_ = csvData
		_ = uploadURL

		now := time.Now()
		report.StorageKey = storageKey
		report.Status = model.ReportStatusReady
		report.CompletedAt = &now
	}

	_ = s.repo.Update(ctx, report)
}

func generateCSV(transactions []model.Transaction) ([]byte, error) {
	var buf []byte
	writer := csv.NewWriter(nil)

	headers := []string{"ID", "Category ID", "Amount", "Currency", "Note", "Date"}
	_ = writer.Write(headers)

	for _, t := range transactions {
		row := []string{
			t.ID.String(),
			t.CategoryID.String(),
			fmt.Sprintf("%d", t.AmountMinor),
			t.Currency,
			t.Note,
			t.OccurredAt.Format("2006-01-02"),
		}
		_ = writer.Write(row)
	}

	writer.Flush()
	return buf, nil
}

func (s *ReportService) GetDownloadURL(ctx context.Context, reportID uuid.UUID, userID string) (string, error) {
	report, err := s.repo.GetByID(ctx, reportID, userID)
	if err != nil {
		return "", err
	}

	if report.Status != model.ReportStatusReady {
		return "", domain.NewDomainError(domain.ErrValidation, "VALIDATION_FAILED", 422, "report is not ready")
	}

	if s.storage == nil {
		return "", fmt.Errorf("storage not configured")
	}

	return s.storage.GenerateDownloadURL(report.StorageKey, 1*time.Hour)
}
