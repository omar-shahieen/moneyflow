package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
)

type SubscriptionRepo struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{pool: pool}
}

func (r *SubscriptionRepo) GetByUserID(ctx context.Context, userID string) (*model.Subscription, error) {
	var s model.Subscription
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, plan, status, payment_provider, provider_customer_id,
		        provider_subscription_id, current_period_end, created_at, updated_at
		 FROM subscriptions WHERE user_id = $1`,
		userID,
	).Scan(&s.ID, &s.UserID, &s.Plan, &s.Status, &s.PaymentProvider,
		&s.ProviderCustomerID, &s.ProviderSubscriptionID, &s.CurrentPeriodEnd,
		&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubscriptionRepo) Create(ctx context.Context, sub *model.Subscription) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO subscriptions (id, user_id, plan, status)
		 VALUES ($1, $2, $3, $4)
		 RETURNING created_at, updated_at`,
		sub.ID, sub.UserID, sub.Plan, sub.Status,
	).Scan(&sub.CreatedAt, &sub.UpdatedAt)
	return err
}

func (r *SubscriptionRepo) Update(ctx context.Context, sub *model.Subscription) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE subscriptions
		 SET plan = $1, status = $2, payment_provider = $3, provider_customer_id = $4,
		     provider_subscription_id = $5, current_period_end = $6, updated_at = NOW()
		 WHERE id = $7`,
		sub.Plan, sub.Status, sub.PaymentProvider, sub.ProviderCustomerID,
		sub.ProviderSubscriptionID, sub.CurrentPeriodEnd, sub.ID,
	)
	return err
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	var s model.Subscription
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, plan, status, payment_provider, provider_customer_id,
		        provider_subscription_id, current_period_end, created_at, updated_at
		 FROM subscriptions WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.UserID, &s.Plan, &s.Status, &s.PaymentProvider,
		&s.ProviderCustomerID, &s.ProviderSubscriptionID, &s.CurrentPeriodEnd,
		&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func EnsureFreeSubscription(ctx context.Context, pool *pgxpool.Pool, userID string) (*model.Subscription, error) {
	var existing model.Subscription
	err := pool.QueryRow(ctx,
		`SELECT id FROM subscriptions WHERE user_id = $1`,
		userID,
	).Scan(&existing.ID)
	if err == nil {
		return &existing, nil
	}
	if err != pgx.ErrNoRows {
		return nil, err
	}

	sub := &model.Subscription{
		ID:     uuid.New(),
		UserID: userID,
		Plan:   model.PlanFree,
		Status: model.SubscriptionStatusActive,
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO subscriptions (id, user_id, plan, status)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id) DO NOTHING`,
		sub.ID, sub.UserID, sub.Plan, sub.Status,
	)
	if err != nil {
		return nil, err
	}
	return sub, nil
}
