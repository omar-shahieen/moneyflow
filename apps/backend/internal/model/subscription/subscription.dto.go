package subscription

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type CheckoutRequest struct {
	Plan       string `json:"plan" binding:"required,oneof=pro vip"`
	SuccessURL string `json:"success_url" binding:"required"`
	CancelURL  string `json:"cancel_url" binding:"required"`
}

func (r CheckoutRequest) Validate() error {
	return validate.Struct(r)
}

func (r *CheckoutRequest) Normalize() {
	r.Plan = strings.ToLower(strings.TrimSpace(r.Plan))
	r.SuccessURL = strings.TrimSpace(r.SuccessURL)
	r.CancelURL = strings.TrimSpace(r.CancelURL)
}

type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	RefNumber   string `json:"ref_number"`
}

type UpgradePlanRequest struct {
	Plan string `json:"plan" binding:"required,oneof=pro vip"`
}

func (r UpgradePlanRequest) Validate() error {
	return validate.Struct(r)
}

func (r *UpgradePlanRequest) Normalize() {
	r.Plan = strings.ToLower(strings.TrimSpace(r.Plan))
}

type PaymentStatusRequest struct {
	RefNumber string `uri:"ref_number" binding:"required"`
}

func (r PaymentStatusRequest) Validate() error { return validate.Struct(r) }

func (r *PaymentStatusRequest) Normalize() { r.RefNumber = strings.TrimSpace(r.RefNumber) }
