package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/report"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type ReportRepo struct {
	server *server.Server
}

func NewReportRepository(server *server.Server) *ReportRepo {
	return &ReportRepo{server: server}
}

func (r *ReportRepo) GetByID(ctx context.Context, id uuid.UUID, userID string) (*report.Report, error) {
	stmt := `
		SELECT
			id, user_id, format, period_start, period_end, status, storage_key, created_at, completed_at
		FROM
			reports
		WHERE
			id = @id
			AND user_id = @user_id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get report by id query for report_id=%s user_id=%s: %w", id.String(), userID, err)
	}

	rep, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[report.Report])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:reports for report_id=%s user_id=%s: %w", id.String(), userID, err)
	}

	return &rep, nil
}

func (r *ReportRepo) List(ctx context.Context, userID string, query *model.ListQuery, filters report.ReportFilters) (*model.PaginatedResponse[report.Report], error) {
	stmt := `
		SELECT
			id, user_id, format, period_start, period_end, status, storage_key, created_at, completed_at
		FROM
			reports
		WHERE
			user_id = @user_id
	`

	args := pgx.NamedArgs{
		"user_id": userID,
	}

	if filters.Status != nil {
		stmt += ` AND status = @status`
		args["status"] = *filters.Status
	}

	if filters.Format != nil {
		stmt += ` AND format = @format`
		args["format"] = *filters.Format
	}

	sortColumn := r.ensureSortColumn(query.Sort)
	sortOrder := r.ensureSortOrder(query.Order)
	stmt += fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder)

	stmt += ` LIMIT @limit OFFSET @offset`
	args["limit"] = query.Limit
	args["offset"] = (query.Page - 1) * query.Limit

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get reports query for user_id=%s: %w", userID, err)
	}

	reports, err := pgx.CollectRows(rows, pgx.RowToStructByName[report.Report])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.PaginatedResponse[report.Report]{
				Data:       []report.Report{},
				Page:       query.Page,
				Limit:      query.Limit,
				Total:      0,
				TotalPages: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to collect rows from table:reports for user_id=%s: %w", userID, err)
	}

	countStmt := `
		SELECT
			COUNT(*)
		FROM
			reports
		WHERE
			user_id = @user_id
	`

	countArgs := pgx.NamedArgs{
		"user_id": userID,
	}

	if filters.Status != nil {
		countStmt += ` AND status = @status`
		countArgs["status"] = *filters.Status
	}

	if filters.Format != nil {
		countStmt += ` AND format = @format`
		countArgs["format"] = *filters.Format
	}

	var total int
	err = r.server.DB.Pool.QueryRow(ctx, countStmt, countArgs).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count of reports for user_id=%s: %w", userID, err)
	}

	return &model.PaginatedResponse[report.Report]{
		Data:       reports,
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      total,
		TotalPages: (total + query.Limit - 1) / query.Limit,
	}, nil
}

func (r *ReportRepo) Create(ctx context.Context, rep *report.Report) error {
	stmt := `
		INSERT INTO
			reports (id, user_id, format, period_start, period_end, status)
		VALUES
			(@id, @user_id, @format, @period_start, @period_end, @status)
		RETURNING
			created_at
	`

	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"id":           rep.ID,
		"user_id":      rep.UserID,
		"format":       rep.Format,
		"period_start": rep.PeriodStart,
		"period_end":   rep.PeriodEnd,
		"status":       rep.Status,
	}).Scan(&rep.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to execute create report query for user_id=%s: %w", rep.UserID, err)
	}

	return nil
}

func (r *ReportRepo) Update(ctx context.Context, rep *report.Report) error {
	stmt := `
		UPDATE reports
		SET status = @status, storage_key = @storage_key, completed_at = @completed_at
		WHERE id = @id
	`

	tag, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":           rep.ID,
		"status":       rep.Status,
		"storage_key":  rep.StorageKey,
		"completed_at": rep.CompletedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to execute update report query for report_id=%s: %w", rep.ID.String(), err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("report not found")
	}

	return nil
}

func (r *ReportRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.server.DB.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM reports WHERE user_id = @user_id`,
		pgx.NamedArgs{"user_id": userID},
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count reports for user_id=%s: %w", userID, err)
	}
	return count, nil
}

func (r *ReportRepo) ensureSortColumn(sort *string) string {
	allowed := map[string]bool{
		"created_at":   true,
		"period_start": true,
		"period_end":   true,
		"status":       true,
	}
	if sort != nil && allowed[*sort] {
		return *sort
	}
	return "created_at"
}

func (r *ReportRepo) ensureSortOrder(order *string) string {
	if order != nil {
		o := strings.ToLower(*order)
		if o == "asc" || o == "desc" {
			return o
		}
	}
	return "desc"
}
