package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/errs"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/transaction"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type TransactionService struct {
	server          *server.Server
	transactionRepo *repository.TransactionRepo
	categoryRepo    *repository.CategoryRepo
}

func NewTransactionService(server *server.Server, transactionRepo *repository.TransactionRepo, categoryRepo *repository.CategoryRepo) *TransactionService {
	return &TransactionService{
		server:          server,
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
	}
}

func (s *TransactionService) GetTransactions(ctx context.Context, userID string, query *transaction.ListTransactionsRequest) (*model.PaginatedResponse[transaction.Transaction], error) {
	query.Normalize()
	transactions, err := s.transactionRepo.List(ctx, userID, query)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch transactions")
		return nil, err
	}

	return transactions, nil
}

func (s *TransactionService) GetTransactionByID(ctx context.Context, userID string, transactionID uuid.UUID) (*transaction.Transaction, error) {
	t, err := s.transactionRepo.GetByID(ctx, transactionID, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch transaction by ID")
		return nil, err
	}

	return t, nil
}

func (s *TransactionService) CreateTransaction(ctx context.Context, userID string, payload *transaction.CreateTransactionRequest) (*transaction.Transaction, error) {
	categoryID, err := uuid.Parse(payload.CategoryID)
	if err != nil {
		return nil, errs.NewBadRequestError("invalid category_id", false, nil, nil, nil)
	}

	occurredAt, err := time.Parse(time.RFC3339, payload.OccurredAt)
	if err != nil {
		return nil, errs.NewBadRequestError("invalid occurred_at format", false, nil, nil, nil)
	}

	currency := payload.Currency
	if currency == "" {
		currency = "USD"
	}

	t := &transaction.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		CategoryID:  categoryID,
		AmountMinor: payload.AmountMinor,
		Currency:    currency,
		Note:        payload.Note,
		ReceiptKey:  payload.ReceiptKey,
		OccurredAt:  occurredAt,
	}

	if err := s.transactionRepo.Create(ctx, t); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to create transaction")
		return nil, err
	}

	s.server.Logger.Info().
		Str("event", "transaction_created").
		Str("transaction_id", t.ID.String()).
		Int64("amount", t.AmountMinor).
		Msg("Transaction created successfully")

	return t, nil
}

func (s *TransactionService) UpdateTransaction(ctx context.Context, userID string, transactionID uuid.UUID, payload *transaction.UpdateTransactionRequest) (*transaction.Transaction, error) {
	t, err := s.transactionRepo.GetByID(ctx, transactionID, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch transaction for update")
		return nil, errs.NewNotFoundError("resource not found", false, nil)
	}

	categoryID, err := uuid.Parse(payload.CategoryID)
	if err != nil {
		return nil, errs.NewBadRequestError("invalid category_id", false, nil, nil, nil)
	}

	occurredAt, err := time.Parse(time.RFC3339, payload.OccurredAt)
	if err != nil {
		return nil, errs.NewBadRequestError("invalid occurred_at format", false, nil, nil, nil)
	}

	t.CategoryID = categoryID
	t.AmountMinor = payload.AmountMinor
	t.Currency = payload.Currency
	t.Note = payload.Note
	t.ReceiptKey = payload.ReceiptKey
	t.OccurredAt = occurredAt

	if err := s.transactionRepo.Update(ctx, t); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to update transaction")
		return nil, err
	}

	s.server.Logger.Info().
		Str("event", "transaction_updated").
		Str("transaction_id", t.ID.String()).
		Msg("Transaction updated successfully")

	return t, nil
}

func (s *TransactionService) DeleteTransaction(ctx context.Context, userID string, transactionID uuid.UUID) error {
	if err := s.transactionRepo.Delete(ctx, transactionID, userID); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to delete transaction")
		return err
	}

	s.server.Logger.Info().
		Str("event", "transaction_deleted").
		Str("transaction_id", transactionID.String()).
		Msg("Transaction deleted successfully")

	return nil
}

func (s *TransactionService) Summary(ctx context.Context, userID string, month time.Time) (*transaction.TransactionSummary, error) {
	summary, err := s.transactionRepo.Summary(ctx, userID, month)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to get transaction summary")
		return nil, err
	}

	return summary, nil
}
