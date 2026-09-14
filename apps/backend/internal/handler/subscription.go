package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type SubscriptionHandler struct {
	Handler
	server *server.Server
}

func NewSubscriptionHandler(s *server.Server) *SubscriptionHandler {
	return &SubscriptionHandler{
		Handler: NewHandler(s),
		server:  s,
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
}

func (h *SubscriptionHandler) Get(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *EmptyRequest) (*SubscriptionResponse, error) {
			userID := GetUserID(c)

			sub, err := repository.EnsureFreeSubscription(c.Request.Context(), h.server, userID)
			if err != nil {
				return nil, err
			}

			resp := &SubscriptionResponse{
				ID:     sub.ID.String(),
				UserID: sub.UserID,
				Plan:   string(sub.Plan),
				Status: string(sub.Status),
			}

			if sub.CurrentPeriodEnd != nil {
				s := sub.CurrentPeriodEnd.Format("2006-01-02T15:04:05Z")
				resp.CurrentPeriodEnd = &s
			}

			return resp, nil
		},
		http.StatusOK,
		&EmptyRequest{},
	)(c)
}
