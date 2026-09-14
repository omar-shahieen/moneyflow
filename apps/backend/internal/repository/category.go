package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/category"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type CategoryRepo struct {
	server *server.Server
}

func NewCategoryRepository(server *server.Server) *CategoryRepo {
	return &CategoryRepo{server: server}
}

func (r *CategoryRepo) GetByID(ctx context.Context, id uuid.UUID, userID string) (*category.Category, error) {
	stmt := `
		SELECT
			id, user_id, name, type, created_at
		FROM
			categories
		WHERE
			id = @id
			AND user_id = @user_id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get category by id query for category_id=%s user_id=%s: %w", id.String(), userID, err)
	}

	cat, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:categories for category_id=%s user_id=%s: %w", id.String(), userID, err)
	}

	return &cat, nil
}

// repository/category.go
func (r *CategoryRepo) List(ctx context.Context, userID string, req *category.ListCategoriesRequest) (*model.PaginatedResponse[category.Category], error) {
	args := pgx.NamedArgs{"user_id": userID}
	where := BuildFilterClause(req, args)

	baseStmt := `SELECT id, user_id, name, type, created_at FROM categories WHERE user_id = @user_id` + where
	countStmt := `SELECT COUNT(*) FROM categories WHERE user_id = @user_id` + where

	return RunPaginatedQuery[category.Category](
		ctx, r.server.DB.Pool, baseStmt, countStmt, args, req.ToListQuery(),
		map[string]bool{"name": true, "type": true, "created_at": true}, "created_at",
	)
}

func (r *CategoryRepo) Create(ctx context.Context, cat *category.Category) error {
	stmt := `
		INSERT INTO
			categories (id, user_id, name, type)
		VALUES
			(@id, @user_id, @name, @type)
		RETURNING
			created_at
	`

	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"id":      cat.ID,
		"user_id": cat.UserID,
		"name":    cat.Name,
		"type":    cat.Type,
	}).Scan(&cat.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to execute create category query for user_id=%s name=%s: %w", cat.UserID, cat.Name, err)
	}

	return nil
}

func (r *CategoryRepo) Update(ctx context.Context, cat *category.Category) error {
	stmt := `
		UPDATE categories
		SET name = @name, type = @type
		WHERE id = @id AND user_id = @user_id
	`

	tag, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":      cat.ID,
		"user_id": cat.UserID,
		"name":    cat.Name,
		"type":    cat.Type,
	})
	if err != nil {
		return fmt.Errorf("failed to execute update category query for category_id=%s user_id=%s: %w", cat.ID.String(), cat.UserID, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	result, err := r.server.DB.Pool.Exec(ctx, `
		DELETE FROM categories
		WHERE id = @id AND user_id = @user_id
	`, pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (r *CategoryRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.server.DB.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM categories WHERE user_id = @user_id`,
		pgx.NamedArgs{"user_id": userID},
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count categories for user_id=%s: %w", userID, err)
	}
	return count, nil
}
