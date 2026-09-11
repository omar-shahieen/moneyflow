package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
)

type RecurringRuleRepo struct {
	pool *pgxpool.Pool
}

func NewRecurringRuleRepository(pool *pgxpool.Pool) *RecurringRuleRepo {
	return &RecurringRuleRepo{pool: pool}
}

func (r *RecurringRuleRepo) GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.RecurringRule, error) {
	var rr model.RecurringRule
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, category_id, amount_minor, currency, frequency,
		        next_run_date, last_generated_date, end_date
		 FROM recurring_rules
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&rr.ID, &rr.UserID, &rr.CategoryID, &rr.AmountMinor, &rr.Currency,
		&rr.Frequency, &rr.NextRunDate, &rr.LastGeneratedDate, &rr.EndDate)
	if err != nil {
		return nil, err
	}
	return &rr, nil
}

func (r *RecurringRuleRepo) List(ctx context.Context, userID string) ([]model.RecurringRule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, category_id, amount_minor, currency, frequency,
		        next_run_date, last_generated_date, end_date
		 FROM recurring_rules
		 WHERE user_id = $1
		 ORDER BY next_run_date ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []model.RecurringRule
	for rows.Next() {
		var rr model.RecurringRule
		if err := rows.Scan(&rr.ID, &rr.UserID, &rr.CategoryID, &rr.AmountMinor, &rr.Currency,
			&rr.Frequency, &rr.NextRunDate, &rr.LastGeneratedDate, &rr.EndDate); err != nil {
			return nil, err
		}
		rules = append(rules, rr)
	}
	return rules, rows.Err()
}

func (r *RecurringRuleRepo) Create(ctx context.Context, rule *model.RecurringRule) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO recurring_rules (id, user_id, category_id, amount_minor, currency, frequency, next_run_date, end_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING last_generated_date`,
		rule.ID, rule.UserID, rule.CategoryID, rule.AmountMinor, rule.Currency,
		rule.Frequency, rule.NextRunDate, rule.EndDate,
	).Scan(&rule.LastGeneratedDate)
	return err
}

func (r *RecurringRuleRepo) Update(ctx context.Context, rule *model.RecurringRule) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE recurring_rules
		 SET category_id = $1, amount_minor = $2, currency = $3, frequency = $4,
		     next_run_date = $5, last_generated_date = $6, end_date = $7
		 WHERE id = $8 AND user_id = $9`,
		rule.CategoryID, rule.AmountMinor, rule.Currency, rule.Frequency,
		rule.NextRunDate, rule.LastGeneratedDate, rule.EndDate,
		rule.ID, rule.UserID,
	)
	return err
}

func (r *RecurringRuleRepo) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM recurring_rules WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *RecurringRuleRepo) GetDueRules(ctx context.Context, before time.Time) ([]model.RecurringRule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, category_id, amount_minor, currency, frequency,
		        next_run_date, last_generated_date, end_date
		 FROM recurring_rules
		 WHERE next_run_date <= $1
		   AND (end_date IS NULL OR end_date >= next_run_date)
		 ORDER BY next_run_date ASC`,
		before,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []model.RecurringRule
	for rows.Next() {
		var rr model.RecurringRule
		if err := rows.Scan(&rr.ID, &rr.UserID, &rr.CategoryID, &rr.AmountMinor, &rr.Currency,
			&rr.Frequency, &rr.NextRunDate, &rr.LastGeneratedDate, &rr.EndDate); err != nil {
			return nil, err
		}
		rules = append(rules, rr)
	}
	return rules, rows.Err()
}

func (r *RecurringRuleRepo) UpdateNextRunDate(ctx context.Context, ruleID uuid.UUID, nextRun time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE recurring_rules SET next_run_date = $1, last_generated_date = $2 WHERE id = $3`,
		nextRun, time.Now(), ruleID,
	)
	return err
}

func (r *RecurringRuleRepo) IsOccurrenceGenerated(ctx context.Context, ruleID uuid.UUID, occurrenceDate time.Time) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM recurring_rule_dedup WHERE rule_id = $1 AND occurrence_date = $2)`,
		ruleID, occurrenceDate,
	).Scan(&exists)
	return exists, err
}

func (r *RecurringRuleRepo) MarkOccurrenceGenerated(ctx context.Context, ruleID uuid.UUID, occurrenceDate time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO recurring_rule_dedup (rule_id, occurrence_date) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		ruleID, occurrenceDate,
	)
	return err
}
