package subscription

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate = validator.New()

type GetSubscriptionRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r GetSubscriptionRequest) Validate() error {
	return validate.Struct(r)
}

type CreateSubscriptionRequest struct {
	Plan                   string `json:"plan" binding:"required,oneof=free pro vip"`
	PaymentProvider        string `json:"payment_provider" binding:"required"`
	ProviderCustomerID     string `json:"provider_customer_id" binding:"required"`
	ProviderSubscriptionID string `json:"provider_subscription_id" binding:"required"`
}

func (r CreateSubscriptionRequest) Validate() error {
	return validate.Struct(r)
}

type UpdateSubscriptionRequest struct {
	ID                     uuid.UUID `uri:"id" binding:"required,uuid"`
	Plan                   string    `json:"plan" binding:"required,oneof=free pro vip"`
	PaymentProvider        string    `json:"payment_provider" binding:"required"`
	ProviderCustomerID     string    `json:"provider_customer_id" binding:"required"`
	ProviderSubscriptionID string    `json:"provider_subscription_id" binding:"required"`
}

func (r UpdateSubscriptionRequest) Validate() error {
	return validate.Struct(r)
}

type CancelSubscriptionRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r CancelSubscriptionRequest) Validate() error {
	return validate.Struct(r)
}

type SubscriptionResponse struct {
	Subscription
}

type SubscriptionListResponse struct {
	Data       []SubscriptionResponse `json:"data"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	Total      int                    `json:"total"`
	TotalPages int                    `json:"totalPages"`
}
