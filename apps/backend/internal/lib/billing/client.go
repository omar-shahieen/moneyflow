package billing

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/omar-shahieen/moneyflow/internal/config"
	"github.com/rs/zerolog"
)

type Client struct {
	merchantCode     string
	secureKey        string
	baseURL          string
	webhookSecret    string
	successReturnURL string
	cancelReturnURL  string
	webhookURL       string
	expiryMinutes    int
	logger           *zerolog.Logger
	httpClient       *http.Client
}

func NewClient(cfg *config.Config, logger *zerolog.Logger) *Client {
	return &Client{
		merchantCode:     cfg.Billing.MerchantCode,
		secureKey:        cfg.Billing.SecureKey,
		baseURL:          strings.TrimRight(cfg.Billing.BaseURL, "/"),
		webhookSecret:    cfg.Billing.WebhookSecret,
		successReturnURL: cfg.Billing.SuccessReturnURL,
		cancelReturnURL:  cfg.Billing.CancelReturnURL,
		webhookURL:       cfg.Billing.WebhookURL,
		expiryMinutes:    cfg.Billing.ChargeExpiryMinutes,
		logger:           logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateCharge initiates a payment charge on Fawry
func (c *Client) CreateCharge(ctx context.Context, userID, email, mobile, plan string, amount float64, refNum string, customerProfileID string) (*ChargeResponse, error) {
	endpoint := fmt.Sprintf("%s/ECommerceWeb/Fawry/payments/charge", c.baseURL)

	// Expiry in unix milliseconds
	expiryDuration := time.Duration(c.expiryMinutes) * time.Minute
	if expiryDuration <= 0 {
		expiryDuration = 30 * time.Minute
	}
	paymentExpiry := time.Now().Add(expiryDuration).UnixMilli()

	amountStr := fmt.Sprintf("%.2f", amount)
	sig := c.GenerateChargeSignature(refNum, customerProfileID, amountStr)

	reqPayload := ChargeRequest{
		MerchantCode:      c.merchantCode,
		MerchantRefNum:    refNum,
		CustomerProfileID: customerProfileID,
		PaymentMethod:     "PAYATFAWRY", // or CARD/CARD_TOKEN depending on profile
		CustomerMobile:    mobile,
		CustomerEmail:     email,
		Amount:            amount,
		CurrencyCode:      "EGP",
		PaymentExpiry:     paymentExpiry,
		Description:       fmt.Sprintf("MoneyFlow %s Subscription", strings.ToUpper(plan)),
		Language:          "en-gb",
		ChargeItems: []ChargeItem{
			{
				ItemID:      plan,
				Description: fmt.Sprintf("%s Plan Subscription", strings.ToUpper(plan)),
				Price:       amount,
				Quantity:    1,
			},
		},
		OrderWebHookUrl: c.webhookURL,
		Signature:       sig,
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal charge request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("fawry charge request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read charge response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fawry charge returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chargeResp ChargeResponse
	if err := json.Unmarshal(bodyBytes, &chargeResp); err != nil {
		return nil, fmt.Errorf("failed to decode charge response: %w", err)
	}

	return &chargeResp, nil
}

// GenerateChargeSignature generates SHA-256 signature for charge creation
func (c *Client) GenerateChargeSignature(merchantRefNum, customerProfileID, amount string) string {
	// Standard Fawry hash format: merchantCode + merchantRefNum + customerProfileId + returnUrl + itemId + quantity + price + secureKey
	// Or merchantCode + merchantRefNum + amount + secureKey
	raw := fmt.Sprintf("%s%s%s%s%s", c.merchantCode, merchantRefNum, customerProfileID, amount, c.secureKey)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// GenerateStatusSignature generates SHA-256 signature for status query
func (c *Client) GenerateStatusSignature(merchantRefNum string) string {
	raw := fmt.Sprintf("%s%s%s", c.merchantCode, merchantRefNum, c.secureKey)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// VerifyWebhookSignature verifies the Fawry webhook event signature
func (c *Client) VerifyWebhookSignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	// First try direct secret check if provided via header/signature
	if signature == c.webhookSecret || signature == c.secureKey {
		return true
	}

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return false
	}

	return c.VerifyEventSignature(&event)
}

// VerifyEventSignature checks the event's embedded signature against computed SHA-256
func (c *Client) VerifyEventSignature(event *WebhookEvent) bool {
	if event == nil || event.Signature == "" {
		return false
	}

	// Secret comparison if Fawry sends auth token as signature
	if event.Signature == c.webhookSecret || event.Signature == c.secureKey {
		return true
	}

	// Fawry notification signature: fawryRefNumber + merchantRefNumber + paymentAmount + orderStatus + secureKey
	candidates := []string{
		fmt.Sprintf("%s%s%.2f%s%s", event.FawryRefNumber, event.MerchantRefNumber, event.PaymentAmount, event.OrderStatus, c.secureKey),
		fmt.Sprintf("%s%s%.2f%.2f%s%s", event.FawryRefNumber, event.MerchantRefNumber, event.PaymentAmount, event.OrderAmount, event.OrderStatus, c.secureKey),
		fmt.Sprintf("%s%s%s%s", event.FawryRefNumber, event.MerchantRefNumber, event.OrderStatus, c.secureKey),
		fmt.Sprintf("%s%s%.2f%s%s", event.FawryRefNumber, event.MerchantRefNumber, event.PaymentAmount, event.OrderStatus, c.webhookSecret),
	}

	sigLower := strings.ToLower(event.Signature)
	for _, raw := range candidates {
		hash := sha256.Sum256([]byte(raw))
		computed := hex.EncodeToString(hash[:])
		if strings.EqualFold(computed, sigLower) {
			return true
		}
	}

	return false
}

// ParseWebhookEvent parses Fawry notification payload
func (c *Client) ParseWebhookEvent(body []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}
	return &event, nil
}

// GetPaymentStatus queries Fawry for status of a merchantRefNumber
func (c *Client) GetPaymentStatus(ctx context.Context, refNum string) (*ChargeResponse, error) {
	sig := c.GenerateStatusSignature(refNum)
	endpoint := fmt.Sprintf("%s/ECommerceWeb/Fawry/payments/status/v2?merchantCode=%s&merchantRefNumber=%s&signature=%s",
		c.baseURL,
		url.QueryEscape(c.merchantCode),
		url.QueryEscape(refNum),
		url.QueryEscape(sig),
	)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("fawry status request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read status response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fawry status query returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chargeResp ChargeResponse
	if err := json.Unmarshal(bodyBytes, &chargeResp); err != nil {
		return nil, fmt.Errorf("failed to decode status response: %w", err)
	}

	return &chargeResp, nil
}
