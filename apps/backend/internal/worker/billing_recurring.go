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

type RecurringBillingWorker struct {
	subscriptionRepo *repository.SubscriptionRepo
	userRepo         *repository.UserRepo
	billingClient    *billing.Client
	billingEventRepo *repository.BillingEventRepo
	cfg              *config.BillingConfig
	logger           *zerolog.Logger
}

func NewRecurringBillingWorker(
	subRepo *repository.SubscriptionRepo,
	userRepo *repository.UserRepo,
	billingClient *billing.Client,
	billingEventRepo *repository.BillingEventRepo,
	cfg *config.BillingConfig,
	logger *zerolog.Logger,
) *RecurringBillingWorker {
	return &RecurringBillingWorker{
		subscriptionRepo: subRepo,
		userRepo:         userRepo,
		billingClient:    billingClient,
		billingEventRepo: billingEventRepo,
		cfg:              cfg,
		logger:           logger,
	}
}

// Run executes a cycle of recurring subscription renewal charges
func (w *RecurringBillingWorker) Run(ctx context.Context) error {
	now := time.Now()
	subs, err := w.subscriptionRepo.GetDueForRecurringCharge(ctx, now)
	if err != nil {
		w.logger.Error().Err(err).Msg("failed to query subscriptions due for recurring charge")
		return err
	}

	for _, sub := range subs {
		// 1. Check if subscription was cancelled at period end
		if sub.CancelAtPeriodEnd {
			sub.Status = subscription.SubscriptionStatusCancelled
			sub.CancelAtPeriodEnd = false
			if err := w.subscriptionRepo.Update(ctx, &sub); err != nil {
				w.logger.Error().Err(err).Str("user_id", sub.UserID).Msg("failed to cancel subscription at period end")
			} else {
				w.logger.Info().Str("user_id", sub.UserID).Msg("subscription ended and cancelled per user request")
			}
			continue
		}

		// 2. Must have a stored card-on-file customerProfileId
		if sub.ProviderCustomerID == "" {
			w.logger.Warn().Str("user_id", sub.UserID).Msg("no card on file for recurring charge; marking past_due")
			sub.Status = subscription.SubscriptionStatusPastDue
			sub.FailedAttempts = 1
			_ = w.subscriptionRepo.Update(ctx, &sub)
			continue
		}

		// 3. Get plan price
		planName := string(sub.Plan)
		amount, err := billing.GetPlanAmount(planName, w.cfg.PlanPrices)
		if err != nil {
			w.logger.Error().Err(err).Str("plan", planName).Msg("pricing not configured for plan")
			continue
		}

		userEmail := ""
		if w.userRepo != nil {
			u, err := w.userRepo.GetByID(ctx, sub.UserID)
			if err == nil && u != nil {
				userEmail = u.Email
			}
		}

		// 4. Fresh merchantRefNum per attempt
		prefix := sub.UserID
		if len(prefix) > 8 {
			prefix = prefix[:8]
		}
		freshRefNum := fmt.Sprintf("REN-%s-%d", prefix, time.Now().UnixNano())

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
			w.logger.Error().Err(err).Str("user_id", sub.UserID).Msg("failed to create renewal charge on fawry")
			sub.Status = subscription.SubscriptionStatusPastDue
			sub.FailedAttempts++
			_ = w.subscriptionRepo.Update(ctx, &sub)
			continue
		}

		// 5. Update subscription to pending for this cycle (wait for webhook)
		sub.Status = subscription.SubscriptionStatusPending
		sub.ProviderSubscriptionID = freshRefNum
		sub.ExpiresAt = &expiresAt
		if chargeResp.CustomerProfileID != "" {
			sub.ProviderCustomerID = chargeResp.CustomerProfileID
		}

		if err := w.subscriptionRepo.Update(ctx, &sub); err != nil {
			w.logger.Error().Err(err).Str("user_id", sub.UserID).Msg("failed to update subscription to pending")
		} else {
			w.logger.Info().
				Str("user_id", sub.UserID).
				Str("ref_num", freshRefNum).
				Msg("renewal charge created and subscription set to pending awaiting webhook")
		}
	}

	return nil
}
