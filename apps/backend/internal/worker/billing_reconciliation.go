package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/omar-shahieen/moneyflow/internal/config"
	"github.com/omar-shahieen/moneyflow/internal/lib/billing"
	"github.com/omar-shahieen/moneyflow/internal/model/subscription"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/rs/zerolog"
)

type ReconciliationWorker struct {
	subscriptionRepo *repository.SubscriptionRepo
	userRepo         *repository.UserRepo
	billingClient    *billing.Client
	cfg              *config.BillingConfig
	logger           *zerolog.Logger
}

func NewReconciliationWorker(
	subRepo *repository.SubscriptionRepo,
	userRepo *repository.UserRepo,
	billingClient *billing.Client,
	cfg *config.BillingConfig,
	logger *zerolog.Logger,
) *ReconciliationWorker {
	return &ReconciliationWorker{
		subscriptionRepo: subRepo,
		userRepo:         userRepo,
		billingClient:    billingClient,
		cfg:              cfg,
		logger:           logger,
	}
}

// Run executes reconciliation tasks: expiring stale charges, resolving stuck renewals via Fawry status API, and retrying bounded dunning
func (w *ReconciliationWorker) Run(ctx context.Context) error {
	now := time.Now()

	// 1. Expire stale pending charges
	if err := w.expireStalePending(ctx, now); err != nil {
		w.logger.Error().Err(err).Msg("error expiring stale pending charges")
	}

	// 2. Resolve stuck past-due / renewals past grace period by querying Fawry
	if err := w.resolveStuckRenewals(ctx, now); err != nil {
		w.logger.Error().Err(err).Msg("error resolving stuck renewals")
	}

	// 3. Drive bounded dunning retries
	if err := w.processDunningRetries(ctx, now); err != nil {
		w.logger.Error().Err(err).Msg("error processing dunning retries")
	}

	return nil
}

// expireStalePending queries Fawry status before expiring pending charges whose TTL passed
func (w *ReconciliationWorker) expireStalePending(ctx context.Context, now time.Time) error {
	pendingSubs, err := w.subscriptionRepo.GetExpiredPending(ctx, now)
	if err != nil {
		return err
	}

	for _, sub := range pendingSubs {
		if sub.ProviderSubscriptionID == "" {
			sub.Status = subscription.SubscriptionStatusExpired
			_ = w.subscriptionRepo.Update(ctx, &sub)
			continue
		}

		// Always query Fawry directly — never assume failure
		statusResp, err := w.billingClient.GetPaymentStatus(ctx, sub.ProviderSubscriptionID)
		if err == nil && statusResp != nil && statusResp.OrderStatus == billing.OrderStatusPaid {
			sub.Status = subscription.SubscriptionStatusActive
			sub.FailedAttempts = 0
			sub.ExpiresAt = nil
			periodEnd := now.AddDate(0, 1, 0)
			sub.CurrentPeriodEnd = &periodEnd
			sub.NextBillingAt = &periodEnd
			if statusResp.CustomerProfileID != "" {
				sub.ProviderCustomerID = statusResp.CustomerProfileID
			}
			_ = w.subscriptionRepo.Update(ctx, &sub)
			w.logger.Info().Str("user_id", sub.UserID).Msg("stale pending charge was actually paid; activated")
			continue
		}

		// Mark expired
		sub.Status = subscription.SubscriptionStatusExpired
		if err := w.subscriptionRepo.Update(ctx, &sub); err != nil {
			w.logger.Error().Err(err).Str("user_id", sub.UserID).Msg("failed to mark subscription expired")
		} else {
			w.logger.Info().Str("user_id", sub.UserID).Msg("stale pending charge expired")
		}
	}

	return nil
}

// resolveStuckRenewals queries Fawry for status of past-due charges past grace days
func (w *ReconciliationWorker) resolveStuckRenewals(ctx context.Context, now time.Time) error {
	graceDays := w.cfg.DunningGraceDays
	if graceDays <= 0 {
		graceDays = 3
	}
	graceCutoff := now.AddDate(0, 0, -graceDays)

	stuckSubs, err := w.subscriptionRepo.GetStuckRenewals(ctx, graceCutoff)
	if err != nil {
		return err
	}

	for _, sub := range stuckSubs {
		if sub.ProviderSubscriptionID != "" {
			statusResp, err := w.billingClient.GetPaymentStatus(ctx, sub.ProviderSubscriptionID)
			if err == nil && statusResp != nil && statusResp.OrderStatus == billing.OrderStatusPaid {
				sub.Status = subscription.SubscriptionStatusActive
				sub.FailedAttempts = 0
				sub.ExpiresAt = nil
				periodEnd := now.AddDate(0, 1, 0)
				sub.CurrentPeriodEnd = &periodEnd
				sub.NextBillingAt = &periodEnd
				if statusResp.CustomerProfileID != "" {
					sub.ProviderCustomerID = statusResp.CustomerProfileID
				}
				_ = w.subscriptionRepo.Update(ctx, &sub)
				w.logger.Info().Str("user_id", sub.UserID).Msg("stuck past-due renewal was paid; reactivated")
				continue
			}
		}

		// Past grace period and not paid: downgrade to cancelled
		sub.Status = subscription.SubscriptionStatusCancelled
		if err := w.subscriptionRepo.Update(ctx, &sub); err != nil {
			w.logger.Error().Err(err).Str("user_id", sub.UserID).Msg("failed to downgrade subscription past grace period")
		} else {
			w.logger.Info().Str("user_id", sub.UserID).Msg("subscription downgraded to cancelled after grace period expired")
		}
	}

	return nil
}

// processDunningRetries retries failed renewal charges within the max retries limit
func (w *ReconciliationWorker) processDunningRetries(ctx context.Context, now time.Time) error {
	maxRetries := w.cfg.DunningMaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	intervalHours := w.cfg.DunningRetryIntervalHr
	if intervalHours <= 0 {
		intervalHours = 24
	}
	retryCutoff := now.Add(-time.Duration(intervalHours) * time.Hour)

	retrySubs, err := w.subscriptionRepo.GetPastDueForRetry(ctx, maxRetries, retryCutoff)
	if err != nil {
		return err
	}

	for _, sub := range retrySubs {
		if sub.CancelAtPeriodEnd {
			sub.Status = subscription.SubscriptionStatusCancelled
			_ = w.subscriptionRepo.Update(ctx, &sub)
			continue
		}

		if sub.ProviderCustomerID == "" {
			continue
		}

		planName := string(sub.Plan)
		amount, err := billing.GetPlanAmount(planName, w.cfg.PlanPrices)
		if err != nil {
			continue
		}

		userEmail := ""
		if w.userRepo != nil {
			u, err := w.userRepo.GetByID(ctx, sub.UserID)
			if err == nil && u != nil {
				userEmail = u.Email
			}
		}

		prefix := sub.UserID
		if len(prefix) > 8 {
			prefix = prefix[:8]
		}
		freshRefNum := fmt.Sprintf("DUN-%s-%d", prefix, time.Now().UnixNano())

		expiryMin := w.cfg.ChargeExpiryMinutes
		if expiryMin <= 0 {
			expiryMin = 30
		}
		expiresAt := now.Add(time.Duration(expiryMin) * time.Minute)

		chargeResp, err := w.billingClient.CreateCharge(
			ctx,
			sub.UserID,
			userEmail,
			"",
			planName,
			amount,
			freshRefNum,
			sub.ProviderCustomerID,
		)
		if err != nil {
			w.logger.Warn().
				Err(err).
				Str("user_id", sub.UserID).
				Int("attempt", sub.FailedAttempts+1).
				Msg("dunning retry charge attempt failed")
			sub.FailedAttempts++
			_ = w.subscriptionRepo.Update(ctx, &sub)
			continue
		}

		sub.Status = subscription.SubscriptionStatusPending
		sub.ProviderSubscriptionID = freshRefNum
		sub.ExpiresAt = &expiresAt
		if chargeResp.CustomerProfileID != "" {
			sub.ProviderCustomerID = chargeResp.CustomerProfileID
		}

		if err := w.subscriptionRepo.Update(ctx, &sub); err != nil {
			w.logger.Error().Err(err).Str("user_id", sub.UserID).Msg("failed to update subscription for dunning attempt")
		} else {
			w.logger.Info().
				Str("user_id", sub.UserID).
				Str("ref_num", freshRefNum).
				Int("attempt", sub.FailedAttempts+1).
				Msg("dunning retry charge initiated")
		}
	}

	return nil
}
