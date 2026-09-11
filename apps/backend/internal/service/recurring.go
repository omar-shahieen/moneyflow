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

type RecurringRuleService struct {
	repo            ports.RecurringRuleRepository
	categoryRepo    ports.CategoryRepository
	transactionRepo ports.TransactionRepository
	planGuard       ports.PlanGuard
}

func NewRecurringRuleService(
	repo ports.RecurringRuleRepository,
	categoryRepo ports.CategoryRepository,
	transactionRepo ports.TransactionRepository,
	planGuard ports.PlanGuard,
) *RecurringRuleService {
	return &RecurringRuleService{
		repo:            repo,
		categoryRepo:    categoryRepo,
		transactionRepo: transactionRepo,
		planGuard:       planGuard,
	}
}

func (s *RecurringRuleService) List(ctx context.Context, userID string) ([]model.RecurringRule, error) {
	return s.repo.List(ctx, userID)
}

func (s *RecurringRuleService) GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.RecurringRule, error) {
	rule, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return rule, nil
}

type CreateRecurringRuleInput struct {
	CategoryID  uuid.UUID
	AmountMinor int64
	Currency    string
	Frequency   model.RecurringFrequency
	NextRunDate time.Time
	EndDate     *time.Time
}

func (s *RecurringRuleService) Create(ctx context.Context, userID string, input CreateRecurringRuleInput) (*model.RecurringRule, error) {
	if input.AmountMinor <= 0 {
		return nil, domain.NewDomainError(domain.ErrValidation, "VALIDATION_FAILED", 422, "amount_minor must be positive")
	}

	if input.Frequency != model.RecurringFrequencyWeekly && input.Frequency != model.RecurringFrequencyMonthly {
		return nil, domain.NewDomainError(domain.ErrValidation, "VALIDATION_FAILED", 422, "frequency must be 'weekly' or 'monthly'")
	}

	if _, err := s.categoryRepo.GetByID(ctx, input.CategoryID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	rule := &model.RecurringRule{
		ID:          uuid.New(),
		UserID:      userID,
		CategoryID:  input.CategoryID,
		AmountMinor: input.AmountMinor,
		Currency:    input.Currency,
		Frequency:   input.Frequency,
		NextRunDate: input.NextRunDate,
		EndDate:     input.EndDate,
	}

	if err := s.repo.Create(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

type UpdateRecurringRuleInput struct {
	CategoryID  uuid.UUID
	AmountMinor int64
	Currency    string
	Frequency   model.RecurringFrequency
	NextRunDate time.Time
	EndDate     *time.Time
}

func (s *RecurringRuleService) Update(ctx context.Context, id uuid.UUID, userID string, input UpdateRecurringRuleInput) (*model.RecurringRule, error) {
	rule, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	rule.CategoryID = input.CategoryID
	rule.AmountMinor = input.AmountMinor
	rule.Currency = input.Currency
	rule.Frequency = input.Frequency
	rule.NextRunDate = input.NextRunDate
	rule.EndDate = input.EndDate

	if err := s.repo.Update(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *RecurringRuleService) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}

func (s *RecurringRuleService) ProcessDueRules(ctx context.Context) error {
	rules, err := s.repo.GetDueRules(ctx, time.Now())
	if err != nil {
		return err
	}

	for _, rule := range rules {
		if err := s.processRule(ctx, &rule); err != nil {
			continue
		}
	}
	return nil
}

func (s *RecurringRuleService) processRule(ctx context.Context, rule *model.RecurringRule) error {
	occurrenceDate := rule.NextRunDate

	generated, err := s.repo.IsOccurrenceGenerated(ctx, rule.ID, occurrenceDate)
	if err != nil {
		return err
	}
	if generated {
		return s.advanceNextRunDate(ctx, rule)
	}

	t := &model.Transaction{
		ID:          uuid.New(),
		UserID:      rule.UserID,
		CategoryID:  rule.CategoryID,
		AmountMinor: rule.AmountMinor,
		Currency:    rule.Currency,
		Note:        "Generated from recurring rule",
		OccurredAt:  occurrenceDate,
	}

	if err := s.transactionRepo.Create(ctx, t); err != nil {
		return err
	}

	if err := s.repo.MarkOccurrenceGenerated(ctx, rule.ID, occurrenceDate); err != nil {
		return err
	}

	return s.advanceNextRunDate(ctx, rule)
}

func (s *RecurringRuleService) advanceNextRunDate(ctx context.Context, rule *model.RecurringRule) error {
	var nextRun time.Time
	switch rule.Frequency {
	case model.RecurringFrequencyWeekly:
		nextRun = rule.NextRunDate.AddDate(0, 0, 7)
	case model.RecurringFrequencyMonthly:
		nextRun = rule.NextRunDate.AddDate(0, 1, 0)
	default:
		return nil
	}

	if rule.EndDate != nil && nextRun.After(*rule.EndDate) {
		return nil
	}

	return s.repo.UpdateNextRunDate(ctx, rule.ID, nextRun)
}
