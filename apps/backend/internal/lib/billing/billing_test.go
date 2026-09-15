package billing_test

import (
	"testing"

	"github.com/omar-shahieen/moneyflow/internal/config"
	"github.com/omar-shahieen/moneyflow/internal/lib/billing"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanPricing(t *testing.T) {
	prices := map[string]float64{
		"pro": 9.99,
		"vip": 19.99,
	}

	amount, err := billing.GetPlanAmount("pro", prices)
	require.NoError(t, err)
	assert.Equal(t, 9.99, amount)

	amount, err = billing.GetPlanAmount("VIP", prices)
	require.NoError(t, err)
	assert.Equal(t, 19.99, amount)

	_, err = billing.GetPlanAmount("unknown", prices)
	assert.ErrorIs(t, err, billing.ErrPlanNotFound)
}

func TestSignatureVerification(t *testing.T) {
	cfg := &config.Config{
		Billing: config.BillingConfig{
			Provider:      "fawry",
			MerchantCode:  "1001",
			SecureKey:     "testsecret123",
			BaseURL:       "https://atfawry.fawrystaging.com",
			WebhookSecret: "testsecret123",
		},
	}
	logger := zerolog.Nop()
	client := billing.NewClient(cfg, &logger)

	event := &billing.WebhookEvent{
		FawryRefNumber:    "FAW12345",
		MerchantRefNumber: "REF98765",
		OrderAmount:       9.99,
		PaymentAmount:     9.99,
		OrderStatus:       billing.OrderStatusPaid,
		Signature:         "testsecret123",
	}

	assert.True(t, client.VerifyEventSignature(event))

	// Invalid signature
	event.Signature = "invalid_signature"
	assert.False(t, client.VerifyEventSignature(event))
}
