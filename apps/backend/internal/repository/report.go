package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
)

type ReportRepo struct {
	pool *pgxpool.Pool
}

func NewReportRepository(pool *pgxpool.Pool) *ReportRepo {
	return &ReportRepo{pool: pool}
}

func (r *ReportRepo) GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.Report, error) {
	var rep model.Report
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, format, period_start, period_end, status, storage_key, created_at, completed_at
		 FROM reports
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&rep.ID, &rep.UserID, &rep.Format, &rep.PeriodStart, &rep.PeriodEnd,
		&rep.Status, &rep.StorageKey, &rep.CreatedAt, &rep.CompletedAt)
	if err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *ReportRepo) List(ctx context.Context, userID string, limit int) ([]model.Report, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, format, period_start, period_end, status, storage_key, created_at, completed_at
		 FROM reports
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []model.Report
	for rows.Next() {
		var rep model.Report
		if err := rows.Scan(&rep.ID, &rep.UserID, &rep.Format, &rep.PeriodStart, &rep.PeriodEnd,
			&rep.Status, &rep.StorageKey, &rep.CreatedAt, &rep.CompletedAt); err != nil {
			return nil, err
		}
		reports = append(reports, rep)
	}
	return reports, rows.Err()
}

func (r *ReportRepo) Create(ctx context.Context, report *model.Report) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO reports (id, user_id, format, period_start, period_end, status)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING created_at`,
		report.ID, report.UserID, report.Format, report.PeriodStart, report.PeriodEnd, report.Status,
	).Scan(&report.CreatedAt)
	return err
}

func (r *ReportRepo) Update(ctx context.Context, report *model.Report) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE reports
		 SET status = $1, storage_key = $2, completed_at = $3
		 WHERE id = $4`,
		report.Status, report.StorageKey, report.CompletedAt, report.ID,
	)
	return err
}

func (r *ReportRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM reports WHERE user_id = $1`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
