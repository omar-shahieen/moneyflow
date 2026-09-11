package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
)

type ImportRepo struct {
	pool *pgxpool.Pool
}

func NewImportRepository(pool *pgxpool.Pool) *ImportRepo {
	return &ImportRepo{pool: pool}
}

func (r *ImportRepo) GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.Import, error) {
	var imp model.Import
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, status, total_rows, success_rows, failed_rows, created_at
		 FROM imports
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&imp.ID, &imp.UserID, &imp.Status, &imp.TotalRows, &imp.SuccessRows, &imp.FailedRows, &imp.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &imp, nil
}

func (r *ImportRepo) Create(ctx context.Context, imp *model.Import) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO imports (id, user_id, status, total_rows, success_rows, failed_rows)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING created_at`,
		imp.ID, imp.UserID, imp.Status, imp.TotalRows, imp.SuccessRows, imp.FailedRows,
	).Scan(&imp.CreatedAt)
	return err
}

func (r *ImportRepo) Update(ctx context.Context, imp *model.Import) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE imports
		 SET status = $1, total_rows = $2, success_rows = $3, failed_rows = $4
		 WHERE id = $5`,
		imp.Status, imp.TotalRows, imp.SuccessRows, imp.FailedRows, imp.ID,
	)
	return err
}

func (r *ImportRepo) EnsureNoActiveImport(ctx context.Context, userID string) error {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM imports WHERE user_id = $1 AND status IN ('pending', 'processing')`,
		userID,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return pgx.ErrNoRows
	}
	return nil
}
