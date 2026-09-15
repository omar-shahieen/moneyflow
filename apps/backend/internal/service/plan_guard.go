package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/errs"
	"github.com/omar-shahieen/moneyflow/internal/model/subscription"
)

type PlanLimits struct {
	MaxCategories    int
	MaxTransactions  int
	MaxBudgets       int
	MaxCSVRows       int
	MaxReports       int
	HasRecurring     bool
	HasSharedBudgets bool
	HasCSVImport     bool
	HasReceipts      bool
	HasPDFReports    bool
}

var planLimitsMap = map[subscription.Plan]PlanLimits{
	subscription.PlanFree: {
		MaxCategories:    5,
		MaxTransactions:  50,
		MaxBudgets:       1,
		MaxCSVRows:       0,
		MaxReports:       3,
		HasRecurring:     false,
		HasSharedBudgets: false,
		HasCSVImport:     false,
		HasReceipts:      false,
		HasPDFReports:    false,
	},
	subscription.PlanPro: {
		MaxCategories:    50,
		MaxTransactions:  500,
		MaxBudgets:       10,
		MaxCSVRows:       100,
		MaxReports:       30,
		HasRecurring:     true,
		HasSharedBudgets: true,
		HasCSVImport:     true,
		HasReceipts:      true,
		HasPDFReports:    true,
	},
	subscription.PlanVIP: {
		MaxCategories:    -1,
		MaxTransactions:  -1,
		MaxBudgets:       -1,
		MaxCSVRows:       10000,
		MaxReports:       -1,
		HasRecurring:     true,
		HasSharedBudgets: true,
		HasCSVImport:     true,
		HasReceipts:      true,
		HasPDFReports:    true,
	},
}

type PlanGuardService struct {
	pool *pgxpool.Pool
}

func NewPlanGuardService(pool *pgxpool.Pool) *PlanGuardService {
	return &PlanGuardService{pool: pool}
}

func (s *PlanGuardService) getUserPlan(ctx context.Context, userID string) (subscription.Plan, error) {
	var plan subscription.Plan
	err := s.pool.QueryRow(ctx,
		`SELECT plan FROM subscriptions WHERE user_id = $1`,
		userID,
	).Scan(&plan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.PlanFree, nil
		}
		return "", err
	}
	return plan, nil
}

func (s *PlanGuardService) getLimits(ctx context.Context, userID string) (PlanLimits, error) {
	plan, err := s.getUserPlan(ctx, userID)
	if err != nil {
		return PlanLimits{}, err
	}
	limits, ok := planLimitsMap[plan]
	if !ok {
		return planLimitsMap[subscription.PlanFree], nil
	}
	return limits, nil
}

func (s *PlanGuardService) CheckCategoryLimit(ctx context.Context, userID string) error {
	limits, err := s.getLimits(ctx, userID)
	if err != nil {
		return err
	}
	if limits.MaxCategories < 0 {
		return nil
	}

	var count int
	err = s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM categories WHERE user_id = $1`,
		userID,
	).Scan(&count)
	if err != nil {
		return err
	}

	if count >= limits.MaxCategories {
		return &errs.HTTPError{Code: "PLAN_LIMIT_EXCEEDED", Message: "Category limit reached for your plan", Status: 402}
	}
	return nil
}

func (s *PlanGuardService) CheckTransactionLimit(ctx context.Context, userID string) error {
	limits, err := s.getLimits(ctx, userID)
	if err != nil {
		return err
	}
	if limits.MaxTransactions < 0 {
		return nil
	}

	var count int
	err = s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM transactions WHERE user_id = $1`,
		userID,
	).Scan(&count)
	if err != nil {
		return err
	}

	if count >= limits.MaxTransactions {
		return &errs.HTTPError{Code: "PLAN_LIMIT_EXCEEDED", Message: "Transaction limit reached for your plan", Status: 402}
	}
	return nil
}

func (s *PlanGuardService) CheckBudgetLimit(ctx context.Context, userID string) error {
	limits, err := s.getLimits(ctx, userID)
	if err != nil {
		return err
	}
	if limits.MaxBudgets < 0 {
		return nil
	}

	var count int
	err = s.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT b.id)
		 FROM budgets b
		 JOIN budget_members bm ON b.id = bm.budget_id
		 WHERE bm.user_id = $1 AND bm.role = 'owner'`,
		userID,
	).Scan(&count)
	if err != nil {
		return err
	}

	if count >= limits.MaxBudgets {
		return &errs.HTTPError{Code: "PLAN_LIMIT_EXCEEDED", Message: "Budget limit reached for your plan", Status: 402}
	}
	return nil
}

func (s *PlanGuardService) CheckCSVImportLimit(ctx context.Context, userID string, rowCount int) error {
	limits, err := s.getLimits(ctx, userID)
	if err != nil {
		return err
	}
	if !limits.HasCSVImport {
		return &errs.HTTPError{Code: "PLAN_LIMIT_EXCEEDED", Message: "CSV import is not available on your plan", Status: 402}
	}
	if limits.MaxCSVRows > 0 && rowCount > limits.MaxCSVRows {
		return &errs.HTTPError{Code: "PLAN_LIMIT_EXCEEDED", Message: "Row limit exceeded for your plan", Status: 402}
	}
	return nil
}

func (s *PlanGuardService) CheckReportLimit(ctx context.Context, userID string) error {
	limits, err := s.getLimits(ctx, userID)
	if err != nil {
		return err
	}
	if limits.MaxReports < 0 {
		return nil
	}

	var count int
	err = s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM reports WHERE user_id = $1`,
		userID,
	).Scan(&count)
	if err != nil {
		return err
	}

	if count >= limits.MaxReports {
		return &errs.HTTPError{Code: "PLAN_LIMIT_EXCEEDED", Message: "Report limit reached for your plan", Status: 402}
	}
	return nil
}
