package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/imports"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type ImportRepo struct {
	server *server.Server
}

func NewImportRepository(server *server.Server) *ImportRepo {
	return &ImportRepo{server: server}
}

func (r *ImportRepo) GetByID(ctx context.Context, id uuid.UUID, userID string) (*imports.Import, error) {
	stmt := `
		SELECT
			id, user_id, status, total_rows, success_rows, failed_rows, created_at,
			COALESCE(storage_key, ''), COALESCE(checksum_sha256, '')
		FROM
			imports
		WHERE
			id = @id
			AND user_id = @user_id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get import by id query for import_id=%s user_id=%s: %w", id.String(), userID, err)
	}

	imp, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[imports.Import])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:imports for import_id=%s user_id=%s: %w", id.String(), userID, err)
	}

	return &imp, nil
}

func (r *ImportRepo) Create(ctx context.Context, imp *imports.Import) error {
	stmt := `
		INSERT INTO
			imports (id, user_id, status, total_rows, success_rows, failed_rows, storage_key, checksum_sha256)
		VALUES
			(@id, @user_id, @status, @total_rows, @success_rows, @failed_rows, @storage_key, @checksum_sha256)
		RETURNING
			created_at
	`

	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"id":              imp.ID,
		"user_id":         imp.UserID,
		"status":          imp.Status,
		"total_rows":      imp.TotalRows,
		"success_rows":    imp.SuccessRows,
		"failed_rows":     imp.FailedRows,
		"storage_key":     imp.StorageKey,
		"checksum_sha256": imp.ChecksumSHA256,
	}).Scan(&imp.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to execute create import query for user_id=%s: %w", imp.UserID, err)
	}

	return nil
}

func (r *ImportRepo) Update(ctx context.Context, imp *imports.Import) error {
	stmt := `
		UPDATE imports
		SET status = @status, total_rows = @total_rows, success_rows = @success_rows,
		    failed_rows = @failed_rows, storage_key = @storage_key, checksum_sha256 = @checksum_sha256
		WHERE id = @id
	`

	tag, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":              imp.ID,
		"status":          imp.Status,
		"total_rows":      imp.TotalRows,
		"success_rows":    imp.SuccessRows,
		"failed_rows":     imp.FailedRows,
		"storage_key":     imp.StorageKey,
		"checksum_sha256": imp.ChecksumSHA256,
	})
	if err != nil {
		return fmt.Errorf("failed to execute update import query for import_id=%s: %w", imp.ID.String(), err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("import not found")
	}

	return nil
}

func (r *ImportRepo) GetByChecksum(ctx context.Context, userID string, checksum string) (*imports.Import, error) {
	stmt := `
		SELECT
			id, user_id, status, total_rows, success_rows, failed_rows, created_at,
			COALESCE(storage_key, ''), COALESCE(checksum_sha256, '')
		FROM
			imports
		WHERE
			user_id = @user_id
			AND checksum_sha256 = @checksum
		ORDER BY created_at DESC
		LIMIT 1
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id":  userID,
		"checksum": checksum,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get import by checksum query: %w", err)
	}

	imp, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[imports.Import])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to collect row: %w", err)
	}

	return &imp, nil
}

func (r *ImportRepo) List(ctx context.Context, userID string, req *imports.ListImportsRequest) (*model.PaginatedResponse[imports.Import], error) {
	args := pgx.NamedArgs{"user_id": userID}
	where := BuildFilterClause(req, args)

	baseStmt := `SELECT id, user_id, status, total_rows, success_rows, failed_rows, created_at,
		COALESCE(storage_key, ''), COALESCE(checksum_sha256, '')
		FROM imports WHERE user_id = @user_id` + where
	countStmt := `SELECT COUNT(*) FROM imports WHERE user_id = @user_id` + where

	return RunPaginatedQuery[imports.Import](
		ctx, r.server.DB.Pool, baseStmt, countStmt, args, req.ToListQuery(),
		map[string]bool{"status": true, "created_at": true, "total_rows": true, "success_rows": true},
		"created_at",
	)
}

func (r *ImportRepo) EnsureNoActiveImport(ctx context.Context, userID string) error {
	var count int
	err := r.server.DB.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM imports WHERE user_id = @user_id AND status IN ('pending', 'processing')`,
		pgx.NamedArgs{"user_id": userID},
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check active imports for user_id=%s: %w", userID, err)
	}
	if count > 0 {
		return pgx.ErrNoRows
	}
	return nil
}
