package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/recurring"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type RecurringRuleRepo struct {
	server *server.Server
}

func NewRecurringRuleRepository(server *server.Server) *RecurringRuleRepo {
	return &RecurringRuleRepo{server: server}
}

func (r *RecurringRuleRepo) GetByID(ctx context.Context, id uuid.UUID, userID string) (*recurring.RecurringRule, error) {
	stmt := `
		SELECT
			id, user_id, category_id, amount_minor, frequency,
			next_run_date, last_generated_date, end_date
		FROM
			recurring_rules
		WHERE
			id = @id
			AND user_id = @user_id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get recurring rule by id query for rule_id=%s user_id=%s: %w", id.String(), userID, err)
	}

	rr, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[recurring.RecurringRule])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:recurring_rules for rule_id=%s user_id=%s: %w", id.String(), userID, err)
	}

	return &rr, nil
}

func (r *RecurringRuleRepo) List(ctx context.Context, userID string, req *recurring.ListRecurringRulesRequest) (*model.PaginatedResponse[recurring.RecurringRule], error) {
	args := pgx.NamedArgs{"user_id": userID}
	where := BuildFilterClause(req, args)

	baseStmt := `SELECT id, user_id, category_id, amount_minor, frequency, next_run_date, last_generated_date, end_date
		FROM recurring_rules WHERE user_id = @user_id` + where
	countStmt := `SELECT COUNT(*) FROM recurring_rules WHERE user_id = @user_id` + where

	return RunPaginatedQuery[recurring.RecurringRule](
		ctx, r.server.DB.Pool, baseStmt, countStmt, args, req.ToListQuery(),
		map[string]bool{"next_run_date": true, "amount_minor": true, "created_at": true}, "next_run_date",
	)
}

func (r *RecurringRuleRepo) Create(ctx context.Context, rule *recurring.RecurringRule) error {
	stmt := `
		INSERT INTO
			recurring_rules (id, user_id, category_id, amount_minor, frequency, next_run_date, end_date)
		VALUES
			(@id, @user_id, @category_id, @amount_minor, @frequency, @next_run_date, @end_date)
		RETURNING
			last_generated_date
	`

	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"id":            rule.ID,
		"user_id":       rule.UserID,
		"category_id":   rule.CategoryID,
		"amount_minor":  rule.AmountMinor,
		"frequency":     rule.Frequency,
		"next_run_date": rule.NextRunDate,
		"end_date":      rule.EndDate,
	}).Scan(&rule.LastGeneratedDate)
	if err != nil {
		return fmt.Errorf("failed to execute create recurring rule query for user_id=%s: %w", rule.UserID, err)
	}

	return nil
}

func (r *RecurringRuleRepo) Update(ctx context.Context, rule *recurring.RecurringRule) error {
	stmt := `
		UPDATE recurring_rules
		SET category_id = @category_id, amount_minor = @amount_minor,
		    frequency = @frequency, next_run_date = @next_run_date,
		    last_generated_date = @last_generated_date, end_date = @end_date
		WHERE id = @id AND user_id = @user_id
	`

	tag, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":                  rule.ID,
		"user_id":             rule.UserID,
		"category_id":         rule.CategoryID,
		"amount_minor":        rule.AmountMinor,
		"frequency":           rule.Frequency,
		"next_run_date":       rule.NextRunDate,
		"last_generated_date": rule.LastGeneratedDate,
		"end_date":            rule.EndDate,
	})
	if err != nil {
		return fmt.Errorf("failed to execute update recurring rule query for rule_id=%s user_id=%s: %w", rule.ID.String(), rule.UserID, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("recurring rule not found")
	}

	return nil
}

func (r *RecurringRuleRepo) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	result, err := r.server.DB.Pool.Exec(ctx, `
		DELETE FROM recurring_rules
		WHERE id = @id AND user_id = @user_id
	`, pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete recurring rule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("recurring rule not found")
	}

	return nil
}

func (r *RecurringRuleRepo) GetDueRules(ctx context.Context, before time.Time) ([]recurring.RecurringRule, error) {
	stmt := `
		SELECT
			id, user_id, category_id, amount_minor, frequency,
			next_run_date, last_generated_date, end_date
		FROM
			recurring_rules
		WHERE
			next_run_date <= @before
			AND (end_date IS NULL OR end_date >= next_run_date)
		ORDER BY
			next_run_date ASC
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"before": before,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get due rules query: %w", err)
	}

	rules, err := pgx.CollectRows(rows, pgx.RowToStructByName[recurring.RecurringRule])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []recurring.RecurringRule{}, nil
		}
		return nil, fmt.Errorf("failed to collect rows from table:recurring_rules for due rules: %w", err)
	}

	return rules, nil
}

func (r *RecurringRuleRepo) UpdateNextRunDate(ctx context.Context, ruleID uuid.UUID, nextRun time.Time) error {
	_, err := r.server.DB.Pool.Exec(ctx,
		`UPDATE recurring_rules SET next_run_date = @next_run_date, last_generated_date = @last_generated_date WHERE id = @id`,
		pgx.NamedArgs{
			"id":                  ruleID,
			"next_run_date":       nextRun,
			"last_generated_date": time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update next run date for rule_id=%s: %w", ruleID.String(), err)
	}
	return nil
}

func (r *RecurringRuleRepo) IsOccurrenceGenerated(ctx context.Context, ruleID uuid.UUID, occurrenceDate time.Time) (bool, error) {
	var exists bool
	err := r.server.DB.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM recurring_rule_dedup WHERE rule_id = @rule_id AND occurrence_date = @occurrence_date)`,
		pgx.NamedArgs{
			"rule_id":         ruleID,
			"occurrence_date": occurrenceDate,
		},
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check occurrence dedup for rule_id=%s: %w", ruleID.String(), err)
	}
	return exists, nil
}

func (r *RecurringRuleRepo) MarkOccurrenceGenerated(ctx context.Context, ruleID uuid.UUID, occurrenceDate time.Time) error {
	_, err := r.server.DB.Pool.Exec(ctx,
		`INSERT INTO recurring_rule_dedup (rule_id, occurrence_date) VALUES (@rule_id, @occurrence_date) ON CONFLICT DO NOTHING`,
		pgx.NamedArgs{
			"rule_id":         ruleID,
			"occurrence_date": occurrenceDate,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to mark occurrence generated for rule_id=%s: %w", ruleID.String(), err)
	}
	return nil
}
