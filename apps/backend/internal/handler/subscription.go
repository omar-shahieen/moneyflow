package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/model/subscription"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type SubscriptionHandler struct {
	Handler
	service *service.SubscriptionService
}

func NewSubscriptionHandler(s *server.Server, svc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		Handler: NewHandler(s),
		service: svc,
	}
}

type SubscriptionResponse struct {
	ID                     string  `json:"id"`
	UserID                 string  `json:"user_id"`
	Plan                   string  `json:"plan"`
	Status                 string  `json:"status"`
	PaymentProvider        string  `json:"payment_provider,omitempty"`
	ProviderCustomerID     string  `json:"provider_customer_id,omitempty"`
	ProviderSubscriptionID string  `json:"provider_subscription_id,omitempty"`
	CurrentPeriodEnd       *string `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd      bool    `json:"cancel_at_period_end"`
	CancelledAt            *string `json:"cancelled_at,omitempty"`
}

func (h *SubscriptionHandler) Get(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *EmptyRequest) (*SubscriptionResponse, error) {
			userID := GetUserID(c)

			sub, err := h.service.GetSubscription(c.Request.Context(), userID)
			if err != nil {
				return nil, err
			}

			resp := &SubscriptionResponse{
				ID:                     sub.ID.String(),
				UserID:                 sub.UserID,
				Plan:                   string(sub.Plan),
				Status:                 string(sub.Status),
				PaymentProvider:        sub.PaymentProvider,
				ProviderCustomerID:     sub.ProviderCustomerID,
				ProviderSubscriptionID: sub.ProviderSubscriptionID,
				CancelAtPeriodEnd:      sub.CancelAtPeriodEnd,
			}

			if sub.CurrentPeriodEnd != nil {
				s := sub.CurrentPeriodEnd.Format("2006-01-02T15:04:05Z")
				resp.CurrentPeriodEnd = &s
			}
			if sub.CancelledAt != nil {
				s := sub.CancelledAt.Format("2006-01-02T15:04:05Z")
				resp.CancelledAt = &s
			}

			return resp, nil
		},
		http.StatusOK,
		&EmptyRequest{},
	)(c)
}

func (h *SubscriptionHandler) Upgrade(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *subscription.UpgradePlanRequest) (*subscription.CheckoutResponse, error) {
			userID := GetUserID(c)
			return h.service.UpgradePlan(c.Request.Context(), userID, req)
		},
		http.StatusOK,
		&subscription.UpgradePlanRequest{},
	)(c)
}

func (h *SubscriptionHandler) Cancel(c *gin.Context) {
	HandleNoContent(
		h.Handler,
		func(c *gin.Context, req *EmptyRequest) error {
			userID := GetUserID(c)
			return h.service.CancelSubscription(c.Request.Context(), userID)
		},
		http.StatusOK,
		&EmptyRequest{},
	)(c)
}

func (h *SubscriptionHandler) Entitlements(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *EmptyRequest) (*subscription.Entitlements, error) {
			userID := GetUserID(c)
			return h.service.Entitlements(c.Request.Context(), userID)
		},
		http.StatusOK,
		&EmptyRequest{},
	)(c)
}
