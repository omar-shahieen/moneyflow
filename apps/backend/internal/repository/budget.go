package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/budget"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type BudgetRepo struct {
	server *server.Server
}

func NewBudgetRepository(server *server.Server) *BudgetRepo {
	return &BudgetRepo{server: server}
}

func (r *BudgetRepo) GetByID(ctx context.Context, id uuid.UUID) (*budget.Budget, error) {
	stmt := `
		SELECT
			id, category_id, monthly_limit_minor, currency, created_at
		FROM
			budgets
		WHERE
			id = @id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id": id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get budget by id query for budget_id=%s: %w", id.String(), err)
	}

	b, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[budget.Budget])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:budgets for budget_id=%s: %w", id.String(), err)
	}

	return &b, nil
}

func (r *BudgetRepo) ListByUser(ctx context.Context, userID string, query *model.ListQuery) (*model.PaginatedResponse[budget.BudgetWithMembers], error) {
	stmt := `
		SELECT
			b.id, b.category_id, b.monthly_limit_minor, b.currency, b.created_at
		FROM
			budgets b
		JOIN
			budget_members bm ON b.id = bm.budget_id
		WHERE
			bm.user_id = @user_id
	`

	args := pgx.NamedArgs{
		"user_id": userID,
	}

	if query.Search != nil {
		stmt += ` AND b.category_id IN (SELECT id FROM categories WHERE user_id = @user_id AND name ILIKE '%' || @search || '%')`
		args["search"] = *query.Search
	}

	sortColumn := "b." + EnsureSortColumn(query.Sort, map[string]bool{"created_at": true, "monthly_limit_minor": true, "currency": true}, "created_at")
	sortOrder := EnsureSortOrder(query.Order)
	stmt += fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder)

	stmt += ` LIMIT @limit OFFSET @offset`
	args["limit"] = query.Limit
	args["offset"] = Offset(query)

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get budgets query for user_id=%s: %w", userID, err)
	}

	budgets, err := pgx.CollectRows(rows, pgx.RowToStructByName[budget.Budget])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EmptyPaginatedResponse[budget.BudgetWithMembers](query), nil
		}
		return nil, fmt.Errorf("failed to collect rows from table:budgets for user_id=%s: %w", userID, err)
	}

	var budgetsWithMembers []budget.BudgetWithMembers
	for _, b := range budgets {
		memberRows, err := r.server.DB.Pool.Query(ctx,
			`SELECT budget_id, user_id, role FROM budget_members WHERE budget_id = @budget_id`,
			pgx.NamedArgs{"budget_id": b.ID},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get members for budget_id=%s: %w", b.ID.String(), err)
		}

		members, err := pgx.CollectRows(memberRows, pgx.RowToStructByName[budget.BudgetMember])
		if err != nil {
			return nil, fmt.Errorf("failed to collect budget members for budget_id=%s: %w", b.ID.String(), err)
		}

		budgetsWithMembers = append(budgetsWithMembers, budget.BudgetWithMembers{
			Budget:  b,
			Members: members,
		})
	}

	countStmt := `
		SELECT
			COUNT(DISTINCT b.id)
		FROM
			budgets b
		JOIN
			budget_members bm ON b.id = bm.budget_id
		WHERE
			bm.user_id = @user_id
	`

	countArgs := pgx.NamedArgs{
		"user_id": userID,
	}

	if query.Search != nil {
		countStmt += ` AND b.category_id IN (SELECT id FROM categories WHERE user_id = @user_id AND name ILIKE '%' || @search || '%')`
		countArgs["search"] = *query.Search
	}

	var total int
	err = r.server.DB.Pool.QueryRow(ctx, countStmt, countArgs).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count of budgets for user_id=%s: %w", userID, err)
	}

	return NewPaginatedResponse(budgetsWithMembers, query, total), nil
}

func (r *BudgetRepo) Create(ctx context.Context, b *budget.Budget) error {
	stmt := `
		INSERT INTO
			budgets (id, category_id, monthly_limit_minor, currency)
		VALUES
			(@id, @category_id, @monthly_limit_minor, @currency)
		RETURNING
			created_at
	`

	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"id":                  b.ID,
		"category_id":         b.CategoryID,
		"monthly_limit_minor": b.MonthlyLimitMinor,
		"currency":            b.Currency,
	}).Scan(&b.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to execute create budget query for category_id=%s: %w", b.CategoryID.String(), err)
	}

	return nil
}

func (r *BudgetRepo) Update(ctx context.Context, b *budget.Budget) error {
	stmt := `
		UPDATE budgets
		SET category_id = @category_id, monthly_limit_minor = @monthly_limit_minor, currency = @currency
		WHERE id = @id
	`

	tag, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":                  b.ID,
		"category_id":         b.CategoryID,
		"monthly_limit_minor": b.MonthlyLimitMinor,
		"currency":            b.Currency,
	})
	if err != nil {
		return fmt.Errorf("failed to execute update budget query for budget_id=%s: %w", b.ID.String(), err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("budget not found")
	}

	return nil
}

func (r *BudgetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.server.DB.Pool.Exec(ctx, `
		DELETE FROM budgets
		WHERE id = @id
	`, pgx.NamedArgs{"id": id})
	if err != nil {
		return fmt.Errorf("failed to delete budget: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("budget not found")
	}

	return nil
}

func (r *BudgetRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.server.DB.Pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT b.id)
		 FROM budgets b
		 JOIN budget_members bm ON b.id = bm.budget_id
		 WHERE bm.user_id = @user_id AND bm.role = 'owner'`,
		pgx.NamedArgs{"user_id": userID},
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count budgets for user_id=%s: %w", userID, err)
	}
	return count, nil
}

func (r *BudgetRepo) GetUsage(ctx context.Context, budgetID uuid.UUID, month time.Time) (int64, error) {
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	var usage int64
	err := r.server.DB.Pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(t.amount_minor), 0)
		 FROM transactions t
		 JOIN budgets b ON t.category_id = b.category_id
		 WHERE b.id = @budget_id
		   AND t.user_id IN (SELECT user_id FROM budget_members WHERE budget_id = @budget_id)
		   AND t.occurred_at >= @start_date AND t.occurred_at < @end_date`,
		pgx.NamedArgs{
			"budget_id":  budgetID,
			"start_date": startOfMonth,
			"end_date":   endOfMonth,
		},
	).Scan(&usage)
	if err != nil {
		return 0, fmt.Errorf("failed to get budget usage for budget_id=%s: %w", budgetID.String(), err)
	}
	return usage, nil
}

type BudgetMemberRepo struct {
	server *server.Server
}

func NewBudgetMemberRepository(server *server.Server) *BudgetMemberRepo {
	return &BudgetMemberRepo{server: server}
}

func (r *BudgetMemberRepo) Get(ctx context.Context, budgetID uuid.UUID, userID string) (*budget.BudgetMember, error) {
	stmt := `
		SELECT
			budget_id, user_id, role
		FROM
			budget_members
		WHERE
			budget_id = @budget_id
			AND user_id = @user_id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"budget_id": budgetID,
		"user_id":   userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get budget member for budget_id=%s user_id=%s: %w", budgetID.String(), userID, err)
	}

	m, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[budget.BudgetMember])
	if err != nil {
		return nil, fmt.Errorf("failed to collect budget member for budget_id=%s user_id=%s: %w", budgetID.String(), userID, err)
	}

	return &m, nil
}

func (r *BudgetMemberRepo) ListByBudget(ctx context.Context, budgetID uuid.UUID) ([]budget.BudgetMember, error) {
	stmt := `
		SELECT
			budget_id, user_id, role
		FROM
			budget_members
		WHERE
			budget_id = @budget_id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"budget_id": budgetID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list budget members for budget_id=%s: %w", budgetID.String(), err)
	}

	members, err := pgx.CollectRows(rows, pgx.RowToStructByName[budget.BudgetMember])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []budget.BudgetMember{}, nil
		}
		return nil, fmt.Errorf("failed to collect budget members for budget_id=%s: %w", budgetID.String(), err)
	}

	return members, nil
}

func (r *BudgetMemberRepo) Add(ctx context.Context, m *budget.BudgetMember) error {
	_, err := r.server.DB.Pool.Exec(ctx,
		`INSERT INTO budget_members (budget_id, user_id, role)
		 VALUES (@budget_id, @user_id, @role)
		 ON CONFLICT (budget_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		pgx.NamedArgs{
			"budget_id": m.BudgetID,
			"user_id":   m.UserID,
			"role":      m.Role,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to add budget member for budget_id=%s user_id=%s: %w", m.BudgetID.String(), m.UserID, err)
	}
	return nil
}

func (r *BudgetMemberRepo) Remove(ctx context.Context, budgetID uuid.UUID, userID string) error {
	result, err := r.server.DB.Pool.Exec(ctx,
		`DELETE FROM budget_members WHERE budget_id = @budget_id AND user_id = @user_id`,
		pgx.NamedArgs{
			"budget_id": budgetID,
			"user_id":   userID,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to remove budget member for budget_id=%s user_id=%s: %w", budgetID.String(), userID, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("budget member not found")
	}

	return nil
}

func (r *BudgetMemberRepo) CountByBudget(ctx context.Context, budgetID uuid.UUID) (int, error) {
	var count int
	err := r.server.DB.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM budget_members WHERE budget_id = @budget_id`,
		pgx.NamedArgs{"budget_id": budgetID},
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count budget members for budget_id=%s: %w", budgetID.String(), err)
	}
	return count, nil
}
