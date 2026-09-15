package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/errs"
	"github.com/omar-shahieen/moneyflow/internal/lib/billing"
	"github.com/omar-shahieen/moneyflow/internal/model/subscription"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type SubscriptionService struct {
	server           *server.Server
	subscriptionRepo *repository.SubscriptionRepo
	userRepo         *repository.UserRepo
	transactionRepo  *repository.TransactionRepo
	billingClient    *billing.Client
}

func NewSubscriptionService(
	s *server.Server,
	subRepo *repository.SubscriptionRepo,
	userRepo *repository.UserRepo,
	txRepo *repository.TransactionRepo,
	billingClient *billing.Client,
) *SubscriptionService {
	return &SubscriptionService{
		server:           s,
		subscriptionRepo: subRepo,
		userRepo:         userRepo,
		transactionRepo:  txRepo,
		billingClient:    billingClient,
	}
}

// GetSubscription ensures a subscription exists for the user and returns it
func (s *SubscriptionService) GetSubscription(ctx context.Context, userID string) (*subscription.Subscription, error) {
	sub, err := s.subscriptionRepo.GetByUserID(ctx, userID)
	if err == nil && sub != nil {
		return sub, nil
	}

	return repository.EnsureFreeSubscription(ctx, s.server, userID)
}

// CreateCheckout initiates a checkout session with Fawry for plan subscription
func (s *SubscriptionService) CreateCheckout(ctx context.Context, userID string, req *subscription.CheckoutRequest) (*subscription.CheckoutResponse, error) {
	plan := strings.ToLower(req.Plan)
	if plan != "pro" && plan != "vip" {
		return nil, errs.NewBadRequestError("invalid plan, must be pro or vip", false, nil, nil, nil)
	}

	amount, err := billing.GetPlanAmount(plan, s.server.Config.Billing.PlanPrices)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan pricing: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	refNum := fmt.Sprintf("SUB-%s-%d", uuid.New().String()[:8], time.Now().Unix())

	expiryMinutes := s.server.Config.Billing.ChargeExpiryMinutes
	if expiryMinutes <= 0 {
		expiryMinutes = 30
	}
	expiresAt := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)

	existingSub, _ := s.subscriptionRepo.GetByUserID(ctx, userID)
	var customerProfileID string
	if existingSub != nil {
		customerProfileID = existingSub.ProviderCustomerID
	}

	chargeResp, err := s.billingClient.CreateCharge(
		ctx,
		userID,
		user.Email,
		"", // mobile optional
		plan,
		amount,
		refNum,
		customerProfileID,
	)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to create charge on fawry")
		return nil, fmt.Errorf("payment gateway error: %w", err)
	}

	// Create or update pending subscription
	sub := &subscription.Subscription{
		ID:                     uuid.New(),
		UserID:                 userID,
		Plan:                   subscription.Plan(plan),
		Status:                 subscription.SubscriptionStatusPending,
		PaymentProvider:        "fawry",
		ProviderCustomerID:     customerProfileID,
		ProviderSubscriptionID: refNum,
		ExpiresAt:              &expiresAt,
		CancelAtPeriodEnd:      false,
		FailedAttempts:         0,
	}
	if chargeResp.CustomerProfileID != "" {
		sub.ProviderCustomerID = chargeResp.CustomerProfileID
	}

	if err := s.subscriptionRepo.Upsert(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to record pending subscription: %w", err)
	}

	checkoutURL := fmt.Sprintf("%s/ECommercePlugin/FawryPay.jsp?chargeRequest=%s", s.server.Config.Billing.BaseURL, refNum)

	return &subscription.CheckoutResponse{
		CheckoutURL: checkoutURL,
		RefNumber:   refNum,
	}, nil
}

// UpgradePlan initiates an upgrade checkout session
func (s *SubscriptionService) UpgradePlan(ctx context.Context, userID string, req *subscription.UpgradePlanRequest) (*subscription.CheckoutResponse, error) {
	checkoutReq := &subscription.CheckoutRequest{
		Plan:       req.Plan,
		SuccessURL: s.server.Config.Billing.SuccessReturnURL,
		CancelURL:  s.server.Config.Billing.CancelReturnURL,
	}
	return s.CreateCheckout(ctx, userID, checkoutReq)
}

// CancelSubscription marks subscription for cancellation at end of period
func (s *SubscriptionService) CancelSubscription(ctx context.Context, userID string) error {
	sub, err := s.subscriptionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return errs.NewNotFoundError("subscription not found", false, nil)
	}

	if sub.Plan == subscription.PlanFree {
		return errs.NewBadRequestError("cannot cancel free subscription", false, nil, nil, nil)
	}

	now := time.Now()
	sub.CancelAtPeriodEnd = true
	sub.CancelledAt = &now

	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to update subscription for cancellation")
		return err
	}

	s.server.Logger.Info().
		Str("user_id", userID).
		Str("plan", string(sub.Plan)).
		Msg("Subscription marked for cancellation at period end")

	return nil
}

// HandleWebhook updates subscription state based on webhook event
func (s *SubscriptionService) HandleWebhook(ctx context.Context, event *billing.WebhookEvent) error {
	sub, err := s.subscriptionRepo.GetByProviderSubscriptionID(ctx, event.MerchantRefNumber)
	if err != nil {
		s.server.Logger.Warn().
			Str("merchant_ref", event.MerchantRefNumber).
			Msg("no subscription found for merchant ref in webhook event")
		return nil
	}

	now := time.Now()

	switch event.OrderStatus {
	case billing.OrderStatusPaid:
		sub.Status = subscription.SubscriptionStatusActive
		sub.FailedAttempts = 0
		sub.ExpiresAt = nil
		periodEnd := now.AddDate(0, 1, 0)
		sub.CurrentPeriodEnd = &periodEnd
		sub.NextBillingAt = &periodEnd
		if event.CustomerProfileID != "" {
			sub.ProviderCustomerID = event.CustomerProfileID
		}

		if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
			return fmt.Errorf("failed to activate subscription: %w", err)
		}

	case billing.OrderStatusExpired:
		sub.Status = subscription.SubscriptionStatusExpired
		if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
			return fmt.Errorf("failed to expire subscription: %w", err)
		}

	case billing.OrderStatusFailed, billing.OrderStatusCanceled:
		if sub.Status == subscription.SubscriptionStatusActive || sub.Status == subscription.SubscriptionStatusPastDue {
			sub.Status = subscription.SubscriptionStatusPastDue
			sub.FailedAttempts++
		} else {
			sub.Status = subscription.SubscriptionStatusExpired
		}
		if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
			return fmt.Errorf("failed to update subscription status: %w", err)
		}

	case billing.OrderStatusRefunded:
		sub.Status = subscription.SubscriptionStatusCancelled
		if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
			return fmt.Errorf("failed to cancel refunded subscription: %w", err)
		}
	}

	return nil
}

// Entitlements derives feature access purely from status and plan.
// Active or past_due during grace period is entitled; cancelled/expired/incomplete is not.
func (s *SubscriptionService) Entitlements(ctx context.Context, userID string) (*subscription.Entitlements, error) {
	sub, err := s.subscriptionRepo.GetByUserID(ctx, userID)
	if err != nil || sub == nil {
		sub, err = repository.EnsureFreeSubscription(ctx, s.server, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure subscription: %w", err)
		}
	}

	isEntitled := false
	if sub.Status == subscription.SubscriptionStatusActive || sub.Status == subscription.SubscriptionStatusPastDue {
		isEntitled = true
	}

	effectivePlan := sub.Plan
	if !isEntitled || effectivePlan == "" {
		effectivePlan = subscription.PlanFree
	}

	switch effectivePlan {
	case subscription.PlanVIP:
		return &subscription.Entitlements{
			Plan:             subscription.PlanVIP,
			IsEntitled:       true,
			MaxCategories:    -1,
			MaxTransactions:  -1,
			MaxBudgets:       -1,
			MaxCSVRows:       10000,
			MaxReports:       -1,
			HasRecurring:     true,
			HasSharedBudgets: true,
			HasCSVImport:     true,
			HasReceipts:      true,
			HasPDFReports:    true,
		}, nil

	case subscription.PlanPro:
		return &subscription.Entitlements{
			Plan:             subscription.PlanPro,
			IsEntitled:       true,
			MaxCategories:    50,
			MaxTransactions:  500,
			MaxBudgets:       10,
			MaxCSVRows:       100,
			MaxReports:       30,
			HasRecurring:     true,
			HasSharedBudgets: true,
			HasCSVImport:     true,
			HasReceipts:      true,
			HasPDFReports:    true,
		}, nil

	default:
		return &subscription.Entitlements{
			Plan:             subscription.PlanFree,
			IsEntitled:       isEntitled,
			MaxCategories:    5,
			MaxTransactions:  50,
			MaxBudgets:       1,
			MaxCSVRows:       0,
			MaxReports:       3,
			HasRecurring:     false,
			HasSharedBudgets: false,
			HasCSVImport:     false,
			HasReceipts:      false,
			HasPDFReports:    false,
		}, nil
	}
}
