package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/errs"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/report"
	"github.com/omar-shahieen/moneyflow/internal/model/transaction"
	"github.com/omar-shahieen/moneyflow/internal/ports"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type ReportService struct {
	server          *server.Server
	reportRepo      *repository.ReportRepo
	transactionRepo *repository.TransactionRepo
	storage         ports.Storage
}

func NewReportService(server *server.Server, reportRepo *repository.ReportRepo, transactionRepo *repository.TransactionRepo, storage ports.Storage) *ReportService {
	return &ReportService{
		server:          server,
		reportRepo:      reportRepo,
		transactionRepo: transactionRepo,
		storage:         storage,
	}
}

func (s *ReportService) GetReports(ctx context.Context, userID string, query *report.ListReportsRequest) (*model.PaginatedResponse[report.Report], error) {
	query.Normalize()
	reports, err := s.reportRepo.List(ctx, userID, query)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch reports")
		return nil, err
	}

	return reports, nil
}

func (s *ReportService) GetReportByID(ctx context.Context, userID string, reportID uuid.UUID) (*report.Report, error) {
	rep, err := s.reportRepo.GetByID(ctx, reportID, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch report by ID")
		return nil, err
	}

	return rep, nil
}

func (s *ReportService) CreateReport(ctx context.Context, userID string, payload *report.CreateReportRequest) (*report.Report, error) {
	periodStart, err := time.Parse("2006-01-02", payload.PeriodStart)
	if err != nil {
		return nil, errs.NewBadRequestError("invalid period_start format", false, nil, nil, nil)
	}

	periodEnd, err := time.Parse("2006-01-02", payload.PeriodEnd)
	if err != nil {
		return nil, errs.NewBadRequestError("invalid period_end format", false, nil, nil, nil)
	}

	rep := &report.Report{
		ID:          uuid.New(),
		UserID:      userID,
		Format:      report.ReportFormat(payload.Format),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Status:      report.ReportStatusPending,
	}

	if err := s.reportRepo.Create(ctx, rep); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to create report")
		return nil, err
	}

	go s.generateReport(rep.ID, userID)

	s.server.Logger.Info().
		Str("event", "report_created").
		Str("report_id", rep.ID.String()).
		Msg("Report created successfully")

	return rep, nil
}

func (s *ReportService) generateReport(reportID uuid.UUID, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	rep, err := s.reportRepo.GetByID(ctx, reportID, userID)
	if err != nil {
		return
	}

	rep.Status = report.ReportStatusProcessing
	_ = s.reportRepo.Update(ctx, rep)

	pageSize := 10000
	req := &transaction.ListTransactionsRequest{
		From: &rep.PeriodStart,
		To:   &rep.PeriodEnd,
	}
	req.Page = &pageSize
	req.PageSize = &pageSize

	result, err := s.transactionRepo.List(ctx, userID, req)
	if err != nil {
		rep.Status = report.ReportStatusFailed
		_ = s.reportRepo.Update(ctx, rep)
		return
	}

	storageKey := fmt.Sprintf("reports/%s/%s.csv", userID, reportID.String())

	if rep.Format == report.ReportFormatCSV {
		csvData, err := generateCSV(result.Data)
		if err != nil {
			rep.Status = report.ReportStatusFailed
			_ = s.reportRepo.Update(ctx, rep)
			return
		}

		uploadURL, err := s.storage.GenerateUploadURL(storageKey, "text/csv", 24*time.Hour)
		if err != nil {
			rep.Status = report.ReportStatusFailed
			_ = s.reportRepo.Update(ctx, rep)
			return
		}

		_ = csvData
		_ = uploadURL

		now := time.Now()
		rep.StorageKey = storageKey
		rep.Status = report.ReportStatusReady
		rep.CompletedAt = &now
	}

	_ = s.reportRepo.Update(ctx, rep)
	s.server.Logger.Info().Str("report_id", reportID.String()).Msg("Report generation completed")
}

func generateCSV(transactions []transaction.Transaction) ([]byte, error) {
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
	rep, err := s.reportRepo.GetByID(ctx, reportID, userID)
	if err != nil {
		return "", err
	}

	if rep.Status != report.ReportStatusReady {
		return "", errs.NewBadRequestError("report is not ready", false, nil, nil, nil)
	}

	if s.storage == nil {
		return "", fmt.Errorf("storage not configured")
	}

	url, err := s.storage.GenerateDownloadURL(rep.StorageKey, 1*time.Hour)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to generate download URL")
		return "", err
	}

	return url, nil
}
