package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
)

type CategoryRepo struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{pool: pool}
}

func (r *CategoryRepo) GetByID(ctx context.Context, id uuid.UUID, userID string) (*model.Category, error) {
	var cat model.Category
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, type, created_at
		 FROM categories
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Type, &cat.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepo) List(ctx context.Context, userID string) ([]model.Category, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, type, created_at
		 FROM categories
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var cat model.Category
		if err := rows.Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Type, &cat.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, rows.Err()
}

func (r *CategoryRepo) Create(ctx context.Context, category *model.Category) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO categories (id, user_id, name, type)
		 VALUES ($1, $2, $3, $4)
		 RETURNING created_at`,
		category.ID, category.UserID, category.Name, category.Type,
	).Scan(&category.CreatedAt)
	return err
}

func (r *CategoryRepo) Update(ctx context.Context, category *model.Category) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE categories
		 SET name = $1, type = $2
		 WHERE id = $3 AND user_id = $4`,
		category.Name, category.Type, category.ID, category.UserID,
	)
	return err
}

func (r *CategoryRepo) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM categories
		 WHERE id = $1 AND user_id = $2`,
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

func (r *CategoryRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM categories WHERE user_id = $1`,
		userID,
	).Scan(&count)
	return count, err
}
