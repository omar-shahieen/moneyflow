package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/model/subscription"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type SubscriptionRepo struct {
	server *server.Server
}

func NewSubscriptionRepository(server *server.Server) *SubscriptionRepo {
	return &SubscriptionRepo{server: server}
}

func (r *SubscriptionRepo) GetByUserID(ctx context.Context, userID string) (*subscription.Subscription, error) {
	stmt := `
		SELECT
			id, user_id, plan, status, payment_provider, provider_customer_id,
			provider_subscription_id, current_period_end, created_at, updated_at
		FROM
			subscriptions
		WHERE
			user_id = @user_id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get subscription by user_id query for user_id=%s: %w", userID, err)
	}

	s, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[subscription.Subscription])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:subscriptions for user_id=%s: %w", userID, err)
	}

	return &s, nil
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*subscription.Subscription, error) {
	stmt := `
		SELECT
			id, user_id, plan, status, payment_provider, provider_customer_id,
			provider_subscription_id, current_period_end, created_at, updated_at
		FROM
			subscriptions
		WHERE
			id = @id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id": id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get subscription by id query for subscription_id=%s: %w", id.String(), err)
	}

	s, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[subscription.Subscription])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:subscriptions for subscription_id=%s: %w", id.String(), err)
	}

	return &s, nil
}

func (r *SubscriptionRepo) Create(ctx context.Context, sub *subscription.Subscription) error {
	stmt := `
		INSERT INTO
			subscriptions (id, user_id, plan, status)
		VALUES
			(@id, @user_id, @plan, @status)
		RETURNING
			created_at, updated_at
	`

	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"id":      sub.ID,
		"user_id": sub.UserID,
		"plan":    sub.Plan,
		"status":  sub.Status,
	}).Scan(&sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to execute create subscription query for user_id=%s: %w", sub.UserID, err)
	}

	return nil
}

func (r *SubscriptionRepo) Update(ctx context.Context, sub *subscription.Subscription) error {
	stmt := `
		UPDATE subscriptions
		SET plan = @plan, status = @status, payment_provider = @payment_provider,
		    provider_customer_id = @provider_customer_id,
		    provider_subscription_id = @provider_subscription_id,
		    current_period_end = @current_period_end, updated_at = NOW()
		WHERE id = @id
	`

	tag, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":                       sub.ID,
		"plan":                     sub.Plan,
		"status":                   sub.Status,
		"payment_provider":         sub.PaymentProvider,
		"provider_customer_id":     sub.ProviderCustomerID,
		"provider_subscription_id": sub.ProviderSubscriptionID,
		"current_period_end":       sub.CurrentPeriodEnd,
	})
	if err != nil {
		return fmt.Errorf("failed to execute update subscription query for subscription_id=%s: %w", sub.ID.String(), err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

func EnsureFreeSubscription(ctx context.Context, server *server.Server, userID string) (*subscription.Subscription, error) {
	stmt := `
		SELECT
			id
		FROM
			subscriptions
		WHERE
			user_id = @user_id
	`

	var existingID uuid.UUID
	err := server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
	}).Scan(&existingID)
	if err == nil {
		return &subscription.Subscription{ID: existingID}, nil
	}
	if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing subscription for user_id=%s: %w", userID, err)
	}

	sub := &subscription.Subscription{
		ID:     uuid.New(),
		UserID: userID,
		Plan:   subscription.PlanFree,
		Status: subscription.SubscriptionStatusActive,
	}
	_, err = server.DB.Pool.Exec(ctx,
		`INSERT INTO subscriptions (id, user_id, plan, status)
		 VALUES (@id, @user_id, @plan, @status)
		 ON CONFLICT (user_id) DO NOTHING`,
		pgx.NamedArgs{
			"id":      sub.ID,
			"user_id": sub.UserID,
			"plan":    sub.Plan,
			"status":  sub.Status,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create free subscription for user_id=%s: %w", userID, err)
	}
	return sub, nil
}
