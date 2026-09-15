package repository

import (
	"context"
	"fmt"
	"time"

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
			provider_subscription_id, current_period_end, expires_at,
			cancel_at_period_end, cancelled_at, failed_attempts, next_billing_at,
			created_at, updated_at
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
			provider_subscription_id, current_period_end, expires_at,
			cancel_at_period_end, cancelled_at, failed_attempts, next_billing_at,
			created_at, updated_at
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

func (r *SubscriptionRepo) GetByProviderSubscriptionID(ctx context.Context, subID string) (*subscription.Subscription, error) {
	stmt := `
		SELECT
			id, user_id, plan, status, payment_provider, provider_customer_id,
			provider_subscription_id, current_period_end, expires_at,
			cancel_at_period_end, cancelled_at, failed_attempts, next_billing_at,
			created_at, updated_at
		FROM
			subscriptions
		WHERE
			provider_subscription_id = @provider_subscription_id
		LIMIT 1
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"provider_subscription_id": subID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription by provider_subscription_id=%s: %w", subID, err)
	}

	s, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[subscription.Subscription])
	if err != nil {
		return nil, fmt.Errorf("failed to collect subscription: %w", err)
	}

	return &s, nil
}

func (r *SubscriptionRepo) Create(ctx context.Context, sub *subscription.Subscription) error {
	stmt := `
		INSERT INTO
			subscriptions (
				id, user_id, plan, status, payment_provider, provider_customer_id,
				provider_subscription_id, current_period_end, expires_at,
				cancel_at_period_end, cancelled_at, failed_attempts, next_billing_at
			)
		VALUES
			(
				@id, @user_id, @plan, @status, @payment_provider, @provider_customer_id,
				@provider_subscription_id, @current_period_end, @expires_at,
				@cancel_at_period_end, @cancelled_at, @failed_attempts, @next_billing_at
			)
		RETURNING
			created_at, updated_at
	`

	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"id":                       sub.ID,
		"user_id":                  sub.UserID,
		"plan":                     sub.Plan,
		"status":                   sub.Status,
		"payment_provider":         sub.PaymentProvider,
		"provider_customer_id":     sub.ProviderCustomerID,
		"provider_subscription_id": sub.ProviderSubscriptionID,
		"current_period_end":       sub.CurrentPeriodEnd,
		"expires_at":               sub.ExpiresAt,
		"cancel_at_period_end":     sub.CancelAtPeriodEnd,
		"cancelled_at":             sub.CancelledAt,
		"failed_attempts":          sub.FailedAttempts,
		"next_billing_at":          sub.NextBillingAt,
	}).Scan(&sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to execute create subscription query for user_id=%s: %w", sub.UserID, err)
	}

	return nil
}

func (r *SubscriptionRepo) Upsert(ctx context.Context, sub *subscription.Subscription) error {
	stmt := `
		INSERT INTO
			subscriptions (
				id, user_id, plan, status, payment_provider, provider_customer_id,
				provider_subscription_id, current_period_end, expires_at,
				cancel_at_period_end, cancelled_at, failed_attempts, next_billing_at
			)
		VALUES
			(
				@id, @user_id, @plan, @status, @payment_provider, @provider_customer_id,
				@provider_subscription_id, @current_period_end, @expires_at,
				@cancel_at_period_end, @cancelled_at, @failed_attempts, @next_billing_at
			)
		ON CONFLICT (user_id) DO UPDATE SET
			plan = EXCLUDED.plan,
			status = EXCLUDED.status,
			payment_provider = EXCLUDED.payment_provider,
			provider_customer_id = EXCLUDED.provider_customer_id,
			provider_subscription_id = EXCLUDED.provider_subscription_id,
			current_period_end = EXCLUDED.current_period_end,
			expires_at = EXCLUDED.expires_at,
			cancel_at_period_end = EXCLUDED.cancel_at_period_end,
			cancelled_at = EXCLUDED.cancelled_at,
			failed_attempts = EXCLUDED.failed_attempts,
			next_billing_at = EXCLUDED.next_billing_at,
			updated_at = NOW()
		RETURNING
			id, created_at, updated_at
	`

	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"id":                       sub.ID,
		"user_id":                  sub.UserID,
		"plan":                     sub.Plan,
		"status":                   sub.Status,
		"payment_provider":         sub.PaymentProvider,
		"provider_customer_id":     sub.ProviderCustomerID,
		"provider_subscription_id": sub.ProviderSubscriptionID,
		"current_period_end":       sub.CurrentPeriodEnd,
		"expires_at":               sub.ExpiresAt,
		"cancel_at_period_end":     sub.CancelAtPeriodEnd,
		"cancelled_at":             sub.CancelledAt,
		"failed_attempts":          sub.FailedAttempts,
		"next_billing_at":          sub.NextBillingAt,
	}).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to upsert subscription: %w", err)
	}

	return nil
}

func (r *SubscriptionRepo) Update(ctx context.Context, sub *subscription.Subscription) error {
	stmt := `
		UPDATE subscriptions
		SET plan = @plan, status = @status, payment_provider = @payment_provider,
		    provider_customer_id = @provider_customer_id,
		    provider_subscription_id = @provider_subscription_id,
		    current_period_end = @current_period_end,
		    expires_at = @expires_at,
		    cancel_at_period_end = @cancel_at_period_end,
		    cancelled_at = @cancelled_at,
		    failed_attempts = @failed_attempts,
		    next_billing_at = @next_billing_at,
		    updated_at = NOW()
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
		"expires_at":               sub.ExpiresAt,
		"cancel_at_period_end":     sub.CancelAtPeriodEnd,
		"cancelled_at":             sub.CancelledAt,
		"failed_attempts":          sub.FailedAttempts,
		"next_billing_at":          sub.NextBillingAt,
	})
	if err != nil {
		return fmt.Errorf("failed to execute update subscription query for subscription_id=%s: %w", sub.ID.String(), err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

func (r *SubscriptionRepo) GetDueForRecurringCharge(ctx context.Context, now time.Time) ([]subscription.Subscription, error) {
	stmt := `
		SELECT
			id, user_id, plan, status, payment_provider, provider_customer_id,
			provider_subscription_id, current_period_end, expires_at,
			cancel_at_period_end, cancelled_at, failed_attempts, next_billing_at,
			created_at, updated_at
		FROM
			subscriptions
		WHERE
			status = 'active'
			AND plan != 'free'
			AND next_billing_at <= @now
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{"now": now})
	if err != nil {
		return nil, fmt.Errorf("failed to get due subscriptions for recurring charge: %w", err)
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[subscription.Subscription])
}

func (r *SubscriptionRepo) GetExpiredPending(ctx context.Context, now time.Time) ([]subscription.Subscription, error) {
	stmt := `
		SELECT
			id, user_id, plan, status, payment_provider, provider_customer_id,
			provider_subscription_id, current_period_end, expires_at,
			cancel_at_period_end, cancelled_at, failed_attempts, next_billing_at,
			created_at, updated_at
		FROM
			subscriptions
		WHERE
			status = 'pending'
			AND expires_at IS NOT NULL
			AND expires_at < @now
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{"now": now})
	if err != nil {
		return nil, fmt.Errorf("failed to get expired pending subscriptions: %w", err)
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[subscription.Subscription])
}

func (r *SubscriptionRepo) GetStuckRenewals(ctx context.Context, graceCutoff time.Time) ([]subscription.Subscription, error) {
	stmt := `
		SELECT
			id, user_id, plan, status, payment_provider, provider_customer_id,
			provider_subscription_id, current_period_end, expires_at,
			cancel_at_period_end, cancelled_at, failed_attempts, next_billing_at,
			created_at, updated_at
		FROM
			subscriptions
		WHERE
			status IN ('pending', 'past_due')
			AND plan != 'free'
			AND current_period_end IS NOT NULL
			AND current_period_end < @grace_cutoff
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{"grace_cutoff": graceCutoff})
	if err != nil {
		return nil, fmt.Errorf("failed to get stuck renewal subscriptions: %w", err)
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[subscription.Subscription])
}

func (r *SubscriptionRepo) GetPastDueForRetry(ctx context.Context, maxRetries int, retryCutoff time.Time) ([]subscription.Subscription, error) {
	stmt := `
		SELECT
			id, user_id, plan, status, payment_provider, provider_customer_id,
			provider_subscription_id, current_period_end, expires_at,
			cancel_at_period_end, cancelled_at, failed_attempts, next_billing_at,
			created_at, updated_at
		FROM
			subscriptions
		WHERE
			status = 'past_due'
			AND plan != 'free'
			AND failed_attempts < @max_retries
			AND updated_at <= @retry_cutoff
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"max_retries":  maxRetries,
		"retry_cutoff": retryCutoff,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get past_due subscriptions for retry: %w", err)
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[subscription.Subscription])
}

func EnsureFreeSubscription(ctx context.Context, server *server.Server, userID string) (*subscription.Subscription, error) {
	stmt := `
		SELECT
			id, user_id, plan, status, payment_provider, provider_customer_id,
			provider_subscription_id, current_period_end, expires_at,
			cancel_at_period_end, cancelled_at, failed_attempts, next_billing_at,
			created_at, updated_at
		FROM
			subscriptions
		WHERE
			user_id = @user_id
	`

	rows, err := server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
	})
	if err == nil {
		s, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[subscription.Subscription])
		if err == nil {
			return &s, nil
		}
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
