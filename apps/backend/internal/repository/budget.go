package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
)

type BudgetRepo struct {
	pool *pgxpool.Pool
}

func NewBudgetRepository(pool *pgxpool.Pool) *BudgetRepo {
	return &BudgetRepo{pool: pool}
}

func (r *BudgetRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Budget, error) {
	var b model.Budget
	err := r.pool.QueryRow(ctx,
		`SELECT id, category_id, monthly_limit_minor, currency, created_at
		 FROM budgets WHERE id = $1`,
		id,
	).Scan(&b.ID, &b.CategoryID, &b.MonthlyLimitMinor, &b.Currency, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BudgetRepo) ListByUser(ctx context.Context, userID string) ([]model.BudgetWithMembers, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT b.id, b.category_id, b.monthly_limit_minor, b.currency, b.created_at
		 FROM budgets b
		 JOIN budget_members bm ON b.id = bm.budget_id
		 WHERE bm.user_id = $1
		 ORDER BY b.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []model.BudgetWithMembers
	for rows.Next() {
		var bw model.BudgetWithMembers
		if err := rows.Scan(&bw.ID, &bw.CategoryID, &bw.MonthlyLimitMinor, &bw.Currency, &bw.CreatedAt); err != nil {
			return nil, err
		}

		memberRows, err := r.pool.Query(ctx,
			`SELECT budget_id, user_id, role FROM budget_members WHERE budget_id = $1`,
			bw.ID,
		)
		if err != nil {
			return nil, err
		}

		for memberRows.Next() {
			var m model.BudgetMember
			if err := memberRows.Scan(&m.BudgetID, &m.UserID, &m.Role); err != nil {
				memberRows.Close()
				return nil, err
			}
			bw.Members = append(bw.Members, m)
		}
		memberRows.Close()

		budgets = append(budgets, bw)
	}
	return budgets, rows.Err()
}

func (r *BudgetRepo) Create(ctx context.Context, budget *model.Budget) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO budgets (id, category_id, monthly_limit_minor, currency)
		 VALUES ($1, $2, $3, $4)
		 RETURNING created_at`,
		budget.ID, budget.CategoryID, budget.MonthlyLimitMinor, budget.Currency,
	).Scan(&budget.CreatedAt)
	return err
}

func (r *BudgetRepo) Update(ctx context.Context, budget *model.Budget) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE budgets SET category_id = $1, monthly_limit_minor = $2, currency = $3
		 WHERE id = $4`,
		budget.CategoryID, budget.MonthlyLimitMinor, budget.Currency, budget.ID,
	)
	return err
}

func (r *BudgetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM budgets WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *BudgetRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT b.id)
		 FROM budgets b
		 JOIN budget_members bm ON b.id = bm.budget_id
		 WHERE bm.user_id = $1 AND bm.role = 'owner'`,
		userID,
	).Scan(&count)
	return count, err
}

func (r *BudgetRepo) GetUsage(ctx context.Context, budgetID uuid.UUID, month time.Time) (int64, error) {
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	var usage int64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(t.amount_minor), 0)
		 FROM transactions t
		 JOIN budgets b ON t.category_id = b.category_id
		 WHERE b.id = $1
		   AND t.user_id IN (SELECT user_id FROM budget_members WHERE budget_id = $1)
		   AND t.occurred_at >= $2 AND t.occurred_at < $3`,
		budgetID, startOfMonth, endOfMonth,
	).Scan(&usage)
	return usage, err
}

type BudgetMemberRepo struct {
	pool *pgxpool.Pool
}

func NewBudgetMemberRepository(pool *pgxpool.Pool) *BudgetMemberRepo {
	return &BudgetMemberRepo{pool: pool}
}

func (r *BudgetMemberRepo) Get(ctx context.Context, budgetID uuid.UUID, userID string) (*model.BudgetMember, error) {
	var m model.BudgetMember
	err := r.pool.QueryRow(ctx,
		`SELECT budget_id, user_id, role FROM budget_members
		 WHERE budget_id = $1 AND user_id = $2`,
		budgetID, userID,
	).Scan(&m.BudgetID, &m.UserID, &m.Role)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *BudgetMemberRepo) ListByBudget(ctx context.Context, budgetID uuid.UUID) ([]model.BudgetMember, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT budget_id, user_id, role FROM budget_members WHERE budget_id = $1`,
		budgetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.BudgetMember
	for rows.Next() {
		var m model.BudgetMember
		if err := rows.Scan(&m.BudgetID, &m.UserID, &m.Role); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *BudgetMemberRepo) Add(ctx context.Context, member *model.BudgetMember) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO budget_members (budget_id, user_id, role)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (budget_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		member.BudgetID, member.UserID, member.Role,
	)
	return err
}

func (r *BudgetMemberRepo) Remove(ctx context.Context, budgetID uuid.UUID, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM budget_members WHERE budget_id = $1 AND user_id = $2`,
		budgetID, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *BudgetMemberRepo) CountByBudget(ctx context.Context, budgetID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM budget_members WHERE budget_id = $1`,
		budgetID,
	).Scan(&count)
	return count, err
}
