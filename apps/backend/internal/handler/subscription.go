package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/repository"
)

type SubscriptionHandler struct {
	pool *pgxpool.Pool
}

func NewSubscriptionHandler(pool *pgxpool.Pool) *SubscriptionHandler {
	return &SubscriptionHandler{pool: pool}
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
}

func (h *SubscriptionHandler) Get(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	sub, err := repository.EnsureFreeSubscription(c.Request.Context(), h.pool, userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	resp := SubscriptionResponse{
		ID:     sub.ID.String(),
		UserID: sub.UserID,
		Plan:   string(sub.Plan),
		Status: string(sub.Status),
	}

	if sub.CurrentPeriodEnd != nil {
		s := sub.CurrentPeriodEnd.Format("2006-01-02T15:04:05Z")
		resp.CurrentPeriodEnd = &s
	}

	RespondJSON(c, http.StatusOK, resp)
}
