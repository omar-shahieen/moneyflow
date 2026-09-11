package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.UserAccount, error) {
	var user model.UserAccount
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, display_name, avatar_url, default_currency, timezone, created_at, updated_at
		 FROM user_accounts
		 WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL,
		&user.DefaultCurrency, &user.Timezone, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Create(ctx context.Context, user *model.UserAccount) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_accounts (id, email, display_name, avatar_url, default_currency, timezone)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (id) DO UPDATE SET
		 email = EXCLUDED.email,
		 display_name = EXCLUDED.display_name,
		 avatar_url = EXCLUDED.avatar_url,
		 updated_at = NOW()`,
		user.ID, user.Email, user.DisplayName, user.AvatarURL,
		user.DefaultCurrency, user.Timezone,
	)
	return err
}

func (r *UserRepo) Update(ctx context.Context, user *model.UserAccount) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE user_accounts
		 SET email = $1, display_name = $2, avatar_url = $3, default_currency = $4, timezone = $5, updated_at = NOW()
		 WHERE id = $6`,
		user.Email, user.DisplayName, user.AvatarURL,
		user.DefaultCurrency, user.Timezone, user.ID,
	)
	return err
}
