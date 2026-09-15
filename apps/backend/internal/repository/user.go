package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/model/user"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type UserRepo struct {
	server *server.Server
}

func NewUserRepository(server *server.Server) *UserRepo {
	return &UserRepo{server: server}
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*user.UserAccount, error) {
	stmt := `
		SELECT
			id, email, display_name, avatar_url, timezone, created_at, updated_at
		FROM
			user_accounts
		WHERE
			id = @id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id": id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get user by id query for user_id=%s: %w", id, err)
	}

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user.UserAccount])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:user_accounts for user_id=%s: %w", id, err)
	}

	return &u, nil
}

func (r *UserRepo) Create(ctx context.Context, u *user.UserAccount) error {
	stmt := `
		INSERT INTO
			user_accounts (id, email, display_name, avatar_url, timezone)
		VALUES
			(@id, @email, @display_name, @avatar_url, @timezone)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			avatar_url = EXCLUDED.avatar_url,
			updated_at = NOW()
	`

	_, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":           u.ID,
		"email":        u.Email,
		"display_name": u.DisplayName,
		"avatar_url":   u.AvatarURL,
		"timezone":     u.Timezone,
	})
	if err != nil {
		return fmt.Errorf("failed to execute create user query for user_id=%s: %w", u.ID, err)
	}

	return nil
}

func (r *UserRepo) Update(ctx context.Context, u *user.UserAccount) error {
	stmt := `
		UPDATE user_accounts
		SET email = @email, display_name = @display_name, avatar_url = @avatar_url,
		    timezone = @timezone, updated_at = NOW()
		WHERE id = @id
	`

	tag, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":           u.ID,
		"email":        u.Email,
		"display_name": u.DisplayName,
		"avatar_url":   u.AvatarURL,
		"timezone":     u.Timezone,
	})
	if err != nil {
		return fmt.Errorf("failed to execute update user query for user_id=%s: %w", u.ID, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
