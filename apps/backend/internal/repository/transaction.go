package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
	"github.com/omar-shahieen/moneyflow/internal/domain/ports"
)

type TransactionRepo struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{pool: pool}
}

func (r *TransactionRepo) GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.Transaction, error) {
	var t model.Transaction
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, category_id, amount_minor, currency, note, receipt_key, occurred_at, created_at
		 FROM transactions
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&t.ID, &t.UserID, &t.CategoryID, &t.AmountMinor, &t.Currency,
		&t.Note, &t.ReceiptKey, &t.OccurredAt, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TransactionRepo) List(ctx context.Context, userID string, filter ports.TransactionFilter) ([]model.Transaction, int, error) {
	query := `SELECT id, user_id, category_id, amount_minor, currency, note, receipt_key, occurred_at, created_at
	          FROM transactions WHERE user_id = $1`
	countQuery := `SELECT COUNT(*) FROM transactions WHERE user_id = $1`
	args := []interface{}{userID}
	argIdx := 2

	if filter.CategoryID != nil {
		query += fmt.Sprintf(` AND category_id = $%d`, argIdx)
		countQuery += fmt.Sprintf(` AND category_id = $%d`, argIdx)
		args = append(args, *filter.CategoryID)
		argIdx++
	}

	if filter.Type != nil {
		query += fmt.Sprintf(` AND category_id IN (SELECT id FROM categories WHERE user_id = $1 AND type = $%d)`, argIdx)
		countQuery += fmt.Sprintf(` AND category_id IN (SELECT id FROM categories WHERE user_id = $1 AND type = $%d)`, argIdx)
		args = append(args, *filter.Type)
		argIdx++
	}

	if filter.From != nil {
		query += fmt.Sprintf(` AND occurred_at >= $%d`, argIdx)
		countQuery += fmt.Sprintf(` AND occurred_at >= $%d`, argIdx)
		args = append(args, *filter.From)
		argIdx++
	}

	if filter.To != nil {
		query += fmt.Sprintf(` AND occurred_at <= $%d`, argIdx)
		countQuery += fmt.Sprintf(` AND occurred_at <= $%d`, argIdx)
		args = append(args, *filter.To)
		argIdx++
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query += ` ORDER BY occurred_at DESC`

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []model.Transaction
	for rows.Next() {
		var t model.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.CategoryID, &t.AmountMinor, &t.Currency,
			&t.Note, &t.ReceiptKey, &t.OccurredAt, &t.CreatedAt); err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, t)
	}
	return transactions, total, rows.Err()
}

func (r *TransactionRepo) Create(ctx context.Context, transaction *model.Transaction) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO transactions (id, user_id, category_id, amount_minor, currency, note, receipt_key, occurred_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING created_at`,
		transaction.ID, transaction.UserID, transaction.CategoryID, transaction.AmountMinor,
		transaction.Currency, transaction.Note, transaction.ReceiptKey, transaction.OccurredAt,
	).Scan(&transaction.CreatedAt)
	return err
}

func (r *TransactionRepo) Update(ctx context.Context, transaction *model.Transaction) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE transactions
		 SET category_id = $1, amount_minor = $2, currency = $3, note = $4, occurred_at = $5
		 WHERE id = $6 AND user_id = $7`,
		transaction.CategoryID, transaction.AmountMinor, transaction.Currency,
		transaction.Note, transaction.OccurredAt, transaction.ID, transaction.UserID,
	)
	return err
}

func (r *TransactionRepo) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM transactions WHERE id = $1 AND user_id = $2`,
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

func (r *TransactionRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM transactions WHERE user_id = $1`,
		userID,
	).Scan(&count)
	return count, err
}

func (r *TransactionRepo) Summary(ctx context.Context, userID string, month time.Time) (*ports.TransactionSummary, error) {
	summary := &ports.TransactionSummary{
		ByCategory: []ports.CategoryTotal{},
		ByCurrency: []ports.CurrencyTotal{},
	}

	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	err := r.pool.QueryRow(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN c.type = 'income' THEN t.amount_minor ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN c.type = 'expense' THEN t.amount_minor ELSE 0 END), 0)
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 WHERE t.user_id = $1 AND t.occurred_at >= $2 AND t.occurred_at < $3`,
		userID, startOfMonth, endOfMonth,
	).Scan(&summary.TotalIncome, &summary.TotalExpense)
	if err != nil {
		return nil, err
	}

	catRows, err := r.pool.Query(ctx,
		`SELECT t.category_id, c.name, c.type, SUM(t.amount_minor)
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 WHERE t.user_id = $1 AND t.occurred_at >= $2 AND t.occurred_at < $3
		 GROUP BY t.category_id, c.name, c.type
		 ORDER BY SUM(t.amount_minor) DESC`,
		userID, startOfMonth, endOfMonth,
	)
	if err != nil {
		return nil, err
	}
	defer catRows.Close()

	for catRows.Next() {
		var ct ports.CategoryTotal
		if err := catRows.Scan(&ct.CategoryID, &ct.CategoryName, &ct.Type, &ct.Total); err != nil {
			return nil, err
		}
		summary.ByCategory = append(summary.ByCategory, ct)
	}

	curRows, err := r.pool.Query(ctx,
		`SELECT t.currency,
			COALESCE(SUM(CASE WHEN c.type = 'income' THEN t.amount_minor ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN c.type = 'expense' THEN t.amount_minor ELSE 0 END), 0)
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 WHERE t.user_id = $1 AND t.occurred_at >= $2 AND t.occurred_at < $3
		 GROUP BY t.currency`,
		userID, startOfMonth, endOfMonth,
	)
	if err != nil {
		return nil, err
	}
	defer curRows.Close()

	for curRows.Next() {
		var ct ports.CurrencyTotal
		if err := curRows.Scan(&ct.Currency, &ct.TotalIncome, &ct.TotalExpense); err != nil {
			return nil, err
		}
		summary.ByCurrency = append(summary.ByCurrency, ct)
	}

	return summary, nil
}
