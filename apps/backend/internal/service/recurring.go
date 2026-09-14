package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/errs"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/recurring"
	"github.com/omar-shahieen/moneyflow/internal/model/transaction"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type RecurringRuleService struct {
	server          *server.Server
	recurringRepo   *repository.RecurringRuleRepo
	categoryRepo    *repository.CategoryRepo
	transactionRepo *repository.TransactionRepo
}

func NewRecurringRuleService(server *server.Server, recurringRepo *repository.RecurringRuleRepo, categoryRepo *repository.CategoryRepo, transactionRepo *repository.TransactionRepo) *RecurringRuleService {
	return &RecurringRuleService{
		server:          server,
		recurringRepo:   recurringRepo,
		categoryRepo:    categoryRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *RecurringRuleService) GetRecurringRules(ctx context.Context, userID string, query *recurring.ListRecurringRulesRequest) (*model.PaginatedResponse[recurring.RecurringRule], error) {
	query.Normalize()
	rules, err := s.recurringRepo.List(ctx, userID, query)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch recurring rules")
		return nil, err
	}

	return rules, nil
}

func (s *RecurringRuleService) GetRecurringRuleByID(ctx context.Context, userID string, ruleID uuid.UUID) (*recurring.RecurringRule, error) {
	rule, err := s.recurringRepo.GetByID(ctx, ruleID, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch recurring rule by ID")
		return nil, err
	}

	return rule, nil
}

func (s *RecurringRuleService) CreateRecurringRule(ctx context.Context, userID string, payload *recurring.CreateRecurringRuleRequest) (*recurring.RecurringRule, error) {
	categoryID, err := uuid.Parse(payload.CategoryID)
	if err != nil {
		return nil, errs.NewBadRequestError("invalid category_id", false, nil, nil, nil)
	}

	nextRunDate, err := time.Parse("2006-01-02", payload.NextRunDate)
	if err != nil {
		return nil, errs.NewBadRequestError("invalid next_run_date format", false, nil, nil, nil)
	}

	if _, err := s.categoryRepo.GetByID(ctx, categoryID, userID); err != nil {
		s.server.Logger.Error().Err(err).Msg("category not found")
		return nil, errs.NewNotFoundError("resource not found", false, nil)
	}

	currency := payload.Currency
	if currency == "" {
		currency = "USD"
	}

	rule := &recurring.RecurringRule{
		ID:          uuid.New(),
		UserID:      userID,
		CategoryID:  categoryID,
		AmountMinor: payload.AmountMinor,
		Currency:    currency,
		Frequency:   recurring.RecurringFrequency(payload.Frequency),
		NextRunDate: nextRunDate,
	}

	if payload.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", payload.EndDate)
		if err == nil {
			rule.EndDate = &endDate
		}
	}

	if err := s.recurringRepo.Create(ctx, rule); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to create recurring rule")
		return nil, err
	}

	s.server.Logger.Info().
		Str("event", "recurring_rule_created").
		Str("rule_id", rule.ID.String()).
		Msg("Recurring rule created successfully")

	return rule, nil
}

func (s *RecurringRuleService) UpdateRecurringRule(ctx context.Context, userID string, ruleID uuid.UUID, payload *recurring.UpdateRecurringRuleRequest) (*recurring.RecurringRule, error) {
	rule, err := s.recurringRepo.GetByID(ctx, ruleID, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("recurring rule not found")
		return nil, errs.NewNotFoundError("resource not found", false, nil)
	}

	categoryID, err := uuid.Parse(payload.CategoryID)
	if err != nil {
		return nil, errs.NewBadRequestError("invalid category_id", false, nil, nil, nil)
	}

	nextRunDate, err := time.Parse("2006-01-02", payload.NextRunDate)
	if err != nil {
		return nil, errs.NewBadRequestError("invalid next_run_date format", false, nil, nil, nil)
	}

	rule.CategoryID = categoryID
	rule.AmountMinor = payload.AmountMinor
	rule.Currency = payload.Currency
	rule.Frequency = recurring.RecurringFrequency(payload.Frequency)
	rule.NextRunDate = nextRunDate

	if payload.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", payload.EndDate)
		if err == nil {
			rule.EndDate = &endDate
		}
	} else {
		rule.EndDate = nil
	}

	if err := s.recurringRepo.Update(ctx, rule); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to update recurring rule")
		return nil, err
	}

	s.server.Logger.Info().
		Str("event", "recurring_rule_updated").
		Str("rule_id", rule.ID.String()).
		Msg("Recurring rule updated successfully")

	return rule, nil
}

func (s *RecurringRuleService) DeleteRecurringRule(ctx context.Context, userID string, ruleID uuid.UUID) error {
	if err := s.recurringRepo.Delete(ctx, ruleID, userID); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to delete recurring rule")
		return err
	}

	s.server.Logger.Info().
		Str("event", "recurring_rule_deleted").
		Str("rule_id", ruleID.String()).
		Msg("Recurring rule deleted successfully")

	return nil
}

func (s *RecurringRuleService) ProcessDueRules(ctx context.Context) error {
	rules, err := s.recurringRepo.GetDueRules(ctx, time.Now())
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to fetch due rules")
		return err
	}

	for _, rule := range rules {
		if err := s.processRule(ctx, &rule); err != nil {
			s.server.Logger.Error().Err(err).Str("rule_id", rule.ID.String()).Msg("failed to process rule")
			continue
		}
	}
	return nil
}

func (s *RecurringRuleService) processRule(ctx context.Context, rule *recurring.RecurringRule) error {
	occurrenceDate := rule.NextRunDate

	generated, err := s.recurringRepo.IsOccurrenceGenerated(ctx, rule.ID, occurrenceDate)
	if err != nil {
		return err
	}
	if generated {
		return s.advanceNextRunDate(ctx, rule)
	}

	t := &transaction.Transaction{
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

	if err := s.recurringRepo.MarkOccurrenceGenerated(ctx, rule.ID, occurrenceDate); err != nil {
		return err
	}

	return s.advanceNextRunDate(ctx, rule)
}

func (s *RecurringRuleService) advanceNextRunDate(ctx context.Context, rule *recurring.RecurringRule) error {
	var nextRun time.Time
	switch rule.Frequency {
	case recurring.RecurringFrequencyWeekly:
		nextRun = rule.NextRunDate.AddDate(0, 0, 7)
	case recurring.RecurringFrequencyMonthly:
		nextRun = rule.NextRunDate.AddDate(0, 1, 0)
	default:
		return nil
	}

	if rule.EndDate != nil && nextRun.After(*rule.EndDate) {
		return nil
	}

	return s.recurringRepo.UpdateNextRunDate(ctx, rule.ID, nextRun)
}
