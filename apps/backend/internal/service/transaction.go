package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/domain"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
	"github.com/omar-shahieen/moneyflow/internal/domain/ports"
)

type TransactionService struct {
	repo         ports.TransactionRepository
	categoryRepo ports.CategoryRepository
	planGuard    ports.PlanGuard
}

func NewTransactionService(repo ports.TransactionRepository, categoryRepo ports.CategoryRepository, planGuard ports.PlanGuard) *TransactionService {
	return &TransactionService{
		repo:         repo,
		categoryRepo: categoryRepo,
		planGuard:    planGuard,
	}
}

func (s *TransactionService) List(ctx context.Context, userID string, filter ports.TransactionFilter) ([]model.Transaction, int, error) {
	return s.repo.List(ctx, userID, filter)
}

func (s *TransactionService) GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.Transaction, error) {
	t, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

type CreateTransactionInput struct {
	CategoryID  uuid.UUID
	AmountMinor int64
	Currency    string
	Note        string
	OccurredAt  time.Time
}

func (s *TransactionService) Create(ctx context.Context, userID string, input CreateTransactionInput) (*model.Transaction, error) {
	if input.AmountMinor == 0 {
		return nil, domain.NewDomainError(domain.ErrValidation, "VALIDATION_FAILED", 422, "amount must be non-zero")
	}

	if input.Currency == "" {
		input.Currency = "USD"
	}

	if len(input.Currency) != 3 {
		return nil, domain.NewDomainError(domain.ErrValidation, "VALIDATION_FAILED", 422, "currency must be a 3-letter code")
	}

	if s.planGuard != nil {
		if err := s.planGuard.CheckTransactionLimit(ctx, userID); err != nil {
			return nil, err
		}
	}

	_, err := s.categoryRepo.GetByID(ctx, input.CategoryID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	t := &model.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		CategoryID:  input.CategoryID,
		AmountMinor: input.AmountMinor,
		Currency:    input.Currency,
		Note:        input.Note,
		OccurredAt:  input.OccurredAt,
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

type UpdateTransactionInput struct {
	CategoryID  uuid.UUID
	AmountMinor int64
	Currency    string
	Note        string
	OccurredAt  time.Time
}

func (s *TransactionService) Update(ctx context.Context, id uuid.UUID, userID string, input UpdateTransactionInput) (*model.Transaction, error) {
	t, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	t.CategoryID = input.CategoryID
	t.AmountMinor = input.AmountMinor
	t.Currency = input.Currency
	t.Note = input.Note
	t.OccurredAt = input.OccurredAt

	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TransactionService) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}

func (s *TransactionService) Summary(ctx context.Context, userID string, month time.Time) (*ports.TransactionSummary, error) {
	return s.repo.Summary(ctx, userID, month)
}
