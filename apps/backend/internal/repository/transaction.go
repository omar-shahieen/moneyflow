package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/transaction"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type TransactionRepo struct {
	server *server.Server
}

func NewTransactionRepository(server *server.Server) *TransactionRepo {
	return &TransactionRepo{server: server}
}

func (r *TransactionRepo) GetByID(ctx context.Context, id uuid.UUID, userID string) (*transaction.Transaction, error) {
	stmt := `
		SELECT
			id, user_id, category_id, amount_minor, note, receipt_key,
			billing_provider, billing_ref_id, billing_status, occurred_at, created_at
		FROM
			transactions
		WHERE
			id = @id
			AND user_id = @user_id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get transaction by id query for transaction_id=%s user_id=%s: %w", id.String(), userID, err)
	}

	t, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[transaction.Transaction])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:transactions for transaction_id=%s user_id=%s: %w", id.String(), userID, err)
	}

	return &t, nil
}

func (r *TransactionRepo) GetByBillingRef(ctx context.Context, refID string) (*transaction.Transaction, error) {
	stmt := `
		SELECT
			id, user_id, category_id, amount_minor, note, receipt_key,
			billing_provider, billing_ref_id, billing_status, occurred_at, created_at
		FROM
			transactions
		WHERE
			billing_ref_id = @billing_ref_id
		LIMIT 1
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"billing_ref_id": refID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by billing_ref_id=%s: %w", refID, err)
	}

	t, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[transaction.Transaction])
	if err != nil {
		return nil, fmt.Errorf("failed to collect transaction by billing_ref_id=%s: %w", refID, err)
	}

	return &t, nil
}

func (r *TransactionRepo) List(ctx context.Context, userID string, req *transaction.ListTransactionsRequest) (*model.PaginatedResponse[transaction.Transaction], error) {
	args := pgx.NamedArgs{"user_id": userID}
	where := BuildFilterClause(req, args)

	baseStmt := `SELECT id, user_id, category_id, amount_minor, note, receipt_key, billing_provider, billing_ref_id, billing_status, occurred_at, created_at
		FROM transactions WHERE user_id = @user_id` + where
	countStmt := `SELECT COUNT(*) FROM transactions WHERE user_id = @user_id` + where

	return RunPaginatedQuery[transaction.Transaction](
		ctx, r.server.DB.Pool, baseStmt, countStmt, args, req.ToListQuery(),
		map[string]bool{"amount_minor": true, "occurred_at": true, "created_at": true}, "occurred_at",
	)
}

func (r *TransactionRepo) Create(ctx context.Context, t *transaction.Transaction) error {
	if t.BillingProvider == "" {
		t.BillingProvider = "manual"
	}
	if t.BillingStatus == "" {
		t.BillingStatus = "manual"
	}

	stmt := `
		INSERT INTO
			transactions (id, user_id, category_id, amount_minor, note, receipt_key, billing_provider, billing_ref_id, billing_status, occurred_at)
		VALUES
			(@id, @user_id, @category_id, @amount_minor, @note, @receipt_key, @billing_provider, @billing_ref_id, @billing_status, @occurred_at)
		RETURNING
			created_at
	`

	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"id":               t.ID,
		"user_id":          t.UserID,
		"category_id":      t.CategoryID,
		"amount_minor":     t.AmountMinor,
		"note":             t.Note,
		"receipt_key":      t.ReceiptKey,
		"billing_provider": t.BillingProvider,
		"billing_ref_id":   t.BillingRefID,
		"billing_status":   t.BillingStatus,
		"occurred_at":      t.OccurredAt,
	}).Scan(&t.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to execute create transaction query for user_id=%s: %w", t.UserID, err)
	}

	return nil
}

func (r *TransactionRepo) CreateFromBilling(ctx context.Context, t *transaction.Transaction) error {
	if t.BillingProvider == "" {
		t.BillingProvider = "fawry"
	}
	if t.BillingStatus == "" {
		t.BillingStatus = "paid"
	}

	stmt := `
		INSERT INTO
			transactions (id, user_id, category_id, amount_minor, note, receipt_key, billing_provider, billing_ref_id, billing_status, occurred_at)
		VALUES
			(@id, @user_id, @category_id, @amount_minor, @note, @receipt_key, @billing_provider, @billing_ref_id, @billing_status, @occurred_at)
		RETURNING
			created_at
	`

	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"id":               t.ID,
		"user_id":          t.UserID,
		"category_id":      t.CategoryID,
		"amount_minor":     t.AmountMinor,
		"note":             t.Note,
		"receipt_key":      t.ReceiptKey,
		"billing_provider": t.BillingProvider,
		"billing_ref_id":   t.BillingRefID,
		"billing_status":   t.BillingStatus,
		"occurred_at":      t.OccurredAt,
	}).Scan(&t.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to execute create billing transaction query for user_id=%s: %w", t.UserID, err)
	}

	return nil
}

func (r *TransactionRepo) Update(ctx context.Context, t *transaction.Transaction) error {
	stmt := `
		UPDATE transactions
		SET category_id = @category_id, amount_minor = @amount_minor,
		    note = @note, billing_provider = @billing_provider,
		    billing_ref_id = @billing_ref_id, billing_status = @billing_status,
		    occurred_at = @occurred_at
		WHERE id = @id AND user_id = @user_id
	`

	tag, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":               t.ID,
		"user_id":          t.UserID,
		"category_id":      t.CategoryID,
		"amount_minor":     t.AmountMinor,
		"note":             t.Note,
		"billing_provider": t.BillingProvider,
		"billing_ref_id":   t.BillingRefID,
		"billing_status":   t.BillingStatus,
		"occurred_at":      t.OccurredAt,
	})
	if err != nil {
		return fmt.Errorf("failed to execute update transaction query for transaction_id=%s user_id=%s: %w", t.ID.String(), t.UserID, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

func (r *TransactionRepo) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	result, err := r.server.DB.Pool.Exec(ctx, `
		DELETE FROM transactions
		WHERE id = @id AND user_id = @user_id
	`, pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

func (r *TransactionRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.server.DB.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM transactions WHERE user_id = @user_id`,
		pgx.NamedArgs{"user_id": userID},
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count transactions for user_id=%s: %w", userID, err)
	}
	return count, nil
}

func (r *TransactionRepo) Summary(ctx context.Context, userID string, month time.Time) (*transaction.TransactionSummary, error) {
	summary := &transaction.TransactionSummary{
		ByCategory: []transaction.CategoryTotal{},
	}

	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	err := r.server.DB.Pool.QueryRow(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN c.type = 'income' THEN t.amount_minor ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN c.type = 'expense' THEN t.amount_minor ELSE 0 END), 0)
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 WHERE t.user_id = @user_id AND t.occurred_at >= @start_date AND t.occurred_at < @end_date`,
		pgx.NamedArgs{
			"user_id":    userID,
			"start_date": startOfMonth,
			"end_date":   endOfMonth,
		},
	).Scan(&summary.TotalIncome, &summary.TotalExpense)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction summary for user_id=%s: %w", userID, err)
	}

	catRows, err := r.server.DB.Pool.Query(ctx,
		`SELECT t.category_id, c.name, c.type, SUM(t.amount_minor)
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 WHERE t.user_id = @user_id AND t.occurred_at >= @start_date AND t.occurred_at < @end_date
		 GROUP BY t.category_id, c.name, c.type
		 ORDER BY SUM(t.amount_minor) DESC`,
		pgx.NamedArgs{
			"user_id":    userID,
			"start_date": startOfMonth,
			"end_date":   endOfMonth,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get category totals for user_id=%s: %w", userID, err)
	}

	for catRows.Next() {
		var ct transaction.CategoryTotal
		if err := catRows.Scan(&ct.CategoryID, &ct.CategoryName, &ct.Type, &ct.Total); err != nil {
			return nil, fmt.Errorf("failed to scan category total: %w", err)
		}
		summary.ByCategory = append(summary.ByCategory, ct)
	}

	return summary, nil
}
