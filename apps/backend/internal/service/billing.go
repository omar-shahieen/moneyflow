package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/errs"
	"github.com/omar-shahieen/moneyflow/internal/lib/billing"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/category"
	"github.com/omar-shahieen/moneyflow/internal/model/subscription"
	"github.com/omar-shahieen/moneyflow/internal/model/transaction"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type BillingService struct {
	server              *server.Server
	billingClient       *billing.Client
	transactionRepo     *repository.TransactionRepo
	subscriptionRepo    *repository.SubscriptionRepo
	billingEventRepo    *repository.BillingEventRepo
	categoryRepo        *repository.CategoryRepo
	subscriptionService *SubscriptionService
}

func NewBillingService(
	s *server.Server,
	billingClient *billing.Client,
	txRepo *repository.TransactionRepo,
	subRepo *repository.SubscriptionRepo,
	billingEventRepo *repository.BillingEventRepo,
	categoryRepo *repository.CategoryRepo,
	subService *SubscriptionService,
) *BillingService {
	return &BillingService{
		server:              s,
		billingClient:       billingClient,
		transactionRepo:     txRepo,
		subscriptionRepo:    subRepo,
		billingEventRepo:    billingEventRepo,
		categoryRepo:        categoryRepo,
		subscriptionService: subService,
	}
}

// CreateCheckout creates a checkout session for a subscription plan
func (s *BillingService) CreateCheckout(ctx context.Context, userID, plan, successURL, cancelURL string) (*subscription.CheckoutResponse, error) {
	req := &subscription.CheckoutRequest{
		Plan:       plan,
		SuccessURL: successURL,
		CancelURL:  cancelURL,
	}
	return s.subscriptionService.CreateCheckout(ctx, userID, req)
}

// HandleWebhook processes Fawry notifications with strict signature verification and idempotency gating
func (s *BillingService) HandleWebhook(ctx context.Context, rawBody []byte, signatureHeader string) error {
	event, err := s.billingClient.ParseWebhookEvent(rawBody)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to parse fawry webhook event")
		return errs.NewBadRequestError("invalid webhook payload", false, nil, nil, nil)
	}

	// 1. Verify signature
	sigValid := s.billingClient.VerifyWebhookSignature(rawBody, signatureHeader) ||
		s.billingClient.VerifyEventSignature(event)
	if !sigValid {
		s.server.Logger.Warn().
			Str("fawry_ref", event.FawryRefNumber).
			Str("merchant_ref", event.MerchantRefNumber).
			Msg("webhook signature verification failed")
		return errs.NewUnauthorizedError("invalid signature", false)
	}

	// 2. Webhook idempotency ledger
	created, err := s.billingEventRepo.RecordEvent(ctx, event)
	if err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to record billing event in ledger")
		return fmt.Errorf("failed to record billing event: %w", err)
	}
	if !created {
		// Duplicate delivery: stop here and drop
		s.server.Logger.Info().
			Str("fawry_ref", event.FawryRefNumber).
			Str("merchant_ref", event.MerchantRefNumber).
			Msg("duplicate fawry webhook dropped (already processed)")
		return nil
	}

	// 3. Update subscription status
	if err := s.subscriptionService.HandleWebhook(ctx, event); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to process subscription webhook transition")
		return err
	}

	// 4. On PAID, auto-create the transaction
	if event.OrderStatus == billing.OrderStatusPaid {
		sub, err := s.subscriptionRepo.GetByProviderSubscriptionID(ctx, event.MerchantRefNumber)
		if err == nil && sub != nil {
			categoryID := s.getOrCreateSubscriptionCategory(ctx, sub.UserID)
			amountMinor := int64(event.PaymentAmount * 100)
			if amountMinor == 0 && event.OrderAmount > 0 {
				amountMinor = int64(event.OrderAmount * 100)
			}

			tx := &transaction.Transaction{
				ID:              uuid.New(),
				UserID:          sub.UserID,
				CategoryID:      categoryID,
				AmountMinor:     amountMinor,
				Note:            fmt.Sprintf("Fawry Subscription: %s Plan", strings.ToUpper(string(sub.Plan))),
				BillingProvider: "fawry",
				BillingRefID:    event.FawryRefNumber,
				BillingStatus:   "paid",
				OccurredAt:      time.Now(),
			}

			if err := s.transactionRepo.CreateFromBilling(ctx, tx); err != nil {
				s.server.Logger.Error().Err(err).Msg("failed to auto-create transaction from paid billing event")
			}
		}
	}

	return nil
}

// GetPaymentStatus queries Fawry directly for charge status
func (s *BillingService) GetPaymentStatus(ctx context.Context, refNumber string) (*billing.ChargeResponse, error) {
	return s.billingClient.GetPaymentStatus(ctx, refNumber)
}

// GetBillingHistory retrieves transaction history filtered for billing records
func (s *BillingService) GetBillingHistory(ctx context.Context, userID string, req *transaction.ListTransactionsRequest) (*model.PaginatedResponse[transaction.Transaction], error) {
	if req.BillingStatus == nil {
		paidStatus := "paid"
		req.BillingStatus = &paidStatus
	}
	req.Normalize()
	return s.transactionRepo.List(ctx, userID, req)
}

func (s *BillingService) getOrCreateSubscriptionCategory(ctx context.Context, userID string) uuid.UUID {
	var categoryID uuid.UUID
	err := s.server.DB.Pool.QueryRow(ctx,
		`SELECT id FROM categories WHERE user_id = $1 AND name = $2 LIMIT 1`,
		userID, "Subscription",
	).Scan(&categoryID)
	if err == nil {
		return categoryID
	}

	// Fallback to first expense category if available
	err = s.server.DB.Pool.QueryRow(ctx,
		`SELECT id FROM categories WHERE user_id = $1 AND type = 'expense' LIMIT 1`,
		userID,
	).Scan(&categoryID)
	if err == nil {
		return categoryID
	}

	// Create new category
	newCat := &category.Category{
		ID:     uuid.New(),
		UserID: userID,
		Name:   "Subscription",
		Type:   category.CategoryTypeExpense,
	}
	if err := s.categoryRepo.Create(ctx, newCat); err == nil {
		return newCat.ID
	}

	return uuid.New()
}
