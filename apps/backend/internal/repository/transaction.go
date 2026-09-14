package repository

import (
	"context"
	"errors"
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
			id, user_id, category_id, amount_minor, currency, note, receipt_key, occurred_at, created_at
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

func (r *TransactionRepo) List(ctx context.Context, userID string, query *model.ListQuery, filters transaction.TransactionFilters) (*model.PaginatedResponse[transaction.Transaction], error) {
	stmt := `
		SELECT
			id, user_id, category_id, amount_minor, currency, note, receipt_key, occurred_at, created_at
		FROM
			transactions
		WHERE
			user_id = @user_id
	`

	countStmt := `
		SELECT
			COUNT(*)
		FROM
			transactions
		WHERE
			user_id = @user_id
	`

	args := pgx.NamedArgs{
		"user_id": userID,
	}
	countArgs := pgx.NamedArgs{
		"user_id": userID,
	}

	if filters.CategoryID != nil {
		stmt += ` AND category_id = @category_id`
		countStmt += ` AND category_id = @category_id`
		args["category_id"] = *filters.CategoryID
		countArgs["category_id"] = *filters.CategoryID
	}

	if filters.Type != nil {
		stmt += ` AND category_id IN (SELECT id FROM categories WHERE user_id = @user_id AND type = @type)`
		countStmt += ` AND category_id IN (SELECT id FROM categories WHERE user_id = @user_id AND type = @type)`
		args["type"] = *filters.Type
		countArgs["type"] = *filters.Type
	}

	if filters.From != nil {
		stmt += ` AND occurred_at >= @from_date`
		countStmt += ` AND occurred_at >= @from_date`
		args["from_date"] = *filters.From
		countArgs["from_date"] = *filters.From
	}

	if filters.To != nil {
		stmt += ` AND occurred_at <= @to_date`
		countStmt += ` AND occurred_at <= @to_date`
		args["to_date"] = *filters.To
		countArgs["to_date"] = *filters.To
	}

	if filters.MinAmount != nil {
		stmt += ` AND amount_minor >= @min_amount`
		countStmt += ` AND amount_minor >= @min_amount`
		args["min_amount"] = *filters.MinAmount
		countArgs["min_amount"] = *filters.MinAmount
	}

	if filters.MaxAmount != nil {
		stmt += ` AND amount_minor <= @max_amount`
		countStmt += ` AND amount_minor <= @max_amount`
		args["max_amount"] = *filters.MaxAmount
		countArgs["max_amount"] = *filters.MaxAmount
	}

	if query.Search != nil {
		stmt += ` AND note ILIKE '%' || @search || '%'`
		countStmt += ` AND note ILIKE '%' || @search || '%'`
		args["search"] = *query.Search
		countArgs["search"] = *query.Search
	}

	var total int
	err := r.server.DB.Pool.QueryRow(ctx, countStmt, countArgs).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count of transactions for user_id=%s: %w", userID, err)
	}

	sortColumn := EnsureSortColumn(query.Sort, map[string]bool{"amount_minor": true, "currency": true, "occurred_at": true, "created_at": true}, "occurred_at")
	sortOrder := EnsureSortOrder(query.Order)
	stmt += fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder)

	stmt += ` LIMIT @limit OFFSET @offset`
	args["limit"] = query.Limit
	args["offset"] = Offset(query)

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get transactions query for user_id=%s: %w", userID, err)
	}

	transactions, err := pgx.CollectRows(rows, pgx.RowToStructByName[transaction.Transaction])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EmptyPaginatedResponse[transaction.Transaction](query), nil
		}
		return nil, fmt.Errorf("failed to collect rows from table:transactions for user_id=%s: %w", userID, err)
	}

	return NewPaginatedResponse(transactions, query, total), nil
}

func (r *TransactionRepo) Create(ctx context.Context, t *transaction.Transaction) error {
	stmt := `
		INSERT INTO
			transactions (id, user_id, category_id, amount_minor, currency, note, receipt_key, occurred_at)
		VALUES
			(@id, @user_id, @category_id, @amount_minor, @currency, @note, @receipt_key, @occurred_at)
		RETURNING
			created_at
	`

	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"id":           t.ID,
		"user_id":      t.UserID,
		"category_id":  t.CategoryID,
		"amount_minor": t.AmountMinor,
		"currency":     t.Currency,
		"note":         t.Note,
		"receipt_key":  t.ReceiptKey,
		"occurred_at":  t.OccurredAt,
	}).Scan(&t.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to execute create transaction query for user_id=%s: %w", t.UserID, err)
	}

	return nil
}

func (r *TransactionRepo) Update(ctx context.Context, t *transaction.Transaction) error {
	stmt := `
		UPDATE transactions
		SET category_id = @category_id, amount_minor = @amount_minor, currency = @currency,
		    note = @note, occurred_at = @occurred_at
		WHERE id = @id AND user_id = @user_id
	`

	tag, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":           t.ID,
		"user_id":      t.UserID,
		"category_id":  t.CategoryID,
		"amount_minor": t.AmountMinor,
		"currency":     t.Currency,
		"note":         t.Note,
		"occurred_at":  t.OccurredAt,
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
		ByCurrency: []transaction.CurrencyTotal{},
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

	curRows, err := r.server.DB.Pool.Query(ctx,
		`SELECT t.currency,
			COALESCE(SUM(CASE WHEN c.type = 'income' THEN t.amount_minor ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN c.type = 'expense' THEN t.amount_minor ELSE 0 END), 0)
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 WHERE t.user_id = @user_id AND t.occurred_at >= @start_date AND t.occurred_at < @end_date
		 GROUP BY t.currency`,
		pgx.NamedArgs{
			"user_id":    userID,
			"start_date": startOfMonth,
			"end_date":   endOfMonth,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get currency totals for user_id=%s: %w", userID, err)
	}

	for curRows.Next() {
		var ct transaction.CurrencyTotal
		if err := curRows.Scan(&ct.Currency, &ct.TotalIncome, &ct.TotalExpense); err != nil {
			return nil, fmt.Errorf("failed to scan currency total: %w", err)
		}
		summary.ByCurrency = append(summary.ByCurrency, ct)
	}

	return summary, nil
}
