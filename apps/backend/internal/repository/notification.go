package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/notification"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type NotificationRepo struct {
	server *server.Server
}

func NewNotificationRepository(server *server.Server) *NotificationRepo {
	return &NotificationRepo{server: server}
}

func (r *NotificationRepo) GetByID(ctx context.Context, id uuid.UUID, userID string) (*notification.Notification, error) {
	stmt := `
		SELECT id, user_id, title, message, type, read, created_at, updated_at
		FROM notifications
		WHERE id = @id AND user_id = @user_id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get notification by id query: %w", err)
	}

	n, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[notification.Notification])
	if err != nil {
		return nil, fmt.Errorf("failed to collect notification row: %w", err)
	}

	return &n, nil
}

func (r *NotificationRepo) List(ctx context.Context, userID string, req *notification.ListNotificationsRequest) (*model.PaginatedResponse[notification.Notification], error) {
	args := pgx.NamedArgs{"user_id": userID}
	where := BuildFilterClause(req, args)

	baseStmt := `SELECT id, user_id, title, message, type, read, created_at, updated_at FROM notifications WHERE user_id = @user_id` + where
	countStmt := `SELECT COUNT(*) FROM notifications WHERE user_id = @user_id` + where

	return RunPaginatedQuery[notification.Notification](
		ctx, r.server.DB.Pool, baseStmt, countStmt, args, req.ToListQuery(),
		map[string]bool{"created_at": true}, "created_at",
	)
}

func (r *NotificationRepo) Create(ctx context.Context, n *notification.Notification) error {
	stmt := `
		INSERT INTO notifications (user_id, title, message, type)
		VALUES (@user_id, @title, @message, @type)
		RETURNING id, created_at, updated_at
	`

	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"user_id": n.UserID,
		"title":   n.Title,
		"message": n.Message,
		"type":    n.Type,
	}).Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to execute create notification query: %w", err)
	}

	return nil
}

func (r *NotificationRepo) MarkAsRead(ctx context.Context, id uuid.UUID, userID string) error {
	stmt := `
		UPDATE notifications
		SET read = TRUE, updated_at = NOW()
		WHERE id = @id AND user_id = @user_id
	`

	tag, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	})
	if err != nil {
		return fmt.Errorf("failed to execute mark as read query: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

func (r *NotificationRepo) MarkAllAsRead(ctx context.Context, userID string) error {
	stmt := `
		UPDATE notifications
		SET read = TRUE, updated_at = NOW()
		WHERE user_id = @user_id AND read = FALSE
	`

	_, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
	})
	if err != nil {
		return fmt.Errorf("failed to execute mark all as read query: %w", err)
	}

	return nil
}

func (r *NotificationRepo) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	stmt := `
		DELETE FROM notifications
		WHERE id = @id AND user_id = @user_id
	`

	tag, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	})
	if err != nil {
		return fmt.Errorf("failed to execute delete notification query: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

func (r *NotificationRepo) UnreadCount(ctx context.Context, userID string) (int, error) {
	stmt := `SELECT COUNT(*) FROM notifications WHERE user_id = @user_id AND read = FALSE`

	var count int
	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
	}).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}
