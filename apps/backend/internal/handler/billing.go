package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/subscription"
	"github.com/omar-shahieen/moneyflow/internal/model/transaction"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type BillingHandler struct {
	Handler
	billingService *service.BillingService
}

func NewBillingHandler(s *server.Server, billingService *service.BillingService) *BillingHandler {
	return &BillingHandler{
		Handler:        NewHandler(s),
		billingService: billingService,
	}
}

// CreateCheckout handles POST /api/v1/billing/checkout
func (h *BillingHandler) CreateCheckout(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *subscription.CheckoutRequest) (*subscription.CheckoutResponse, error) {
			userID := GetUserID(c)
			return h.billingService.CreateCheckout(c.Request.Context(), userID, req.Plan, req.SuccessURL, req.CancelURL)
		},
		http.StatusOK,
		&subscription.CheckoutRequest{},
	)(c)
}

// Webhook handles POST /api/v1/billing/webhook (no auth middleware)
func (h *BillingHandler) Webhook(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	sig := c.GetHeader("Fawry-Signature")
	if sig == "" {
		sig = c.GetHeader("X-Fawry-Signature")
	}

	if err := h.billingService.HandleWebhook(c.Request.Context(), bodyBytes, sig); err != nil {
		h.server.Logger.Error().Err(err).Msg("fawry webhook processing failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.String(http.StatusOK, "OK")
}

// GetStatus handles GET /api/v1/billing/status/:ref_number
func (h *BillingHandler) GetStatus(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *subscription.PaymentStatusRequest) (interface{}, error) {
			return h.billingService.GetPaymentStatus(c.Request.Context(), req.RefNumber)
		},
		http.StatusOK,
		&subscription.PaymentStatusRequest{},
	)(c)
}

// GetHistory handles GET /api/v1/billing/history
func (h *BillingHandler) GetHistory(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *transaction.ListTransactionsRequest) (*model.PaginatedResponse[transaction.Transaction], error) {
			userID := GetUserID(c)
			return h.billingService.GetBillingHistory(c.Request.Context(), userID, req)
		},
		http.StatusOK,
		&transaction.ListTransactionsRequest{},
	)(c)
}
