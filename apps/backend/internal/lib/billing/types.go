package billing

type ChargeRequest struct {
	MerchantCode      string       `json:"merchantCode"`
	MerchantRefNum    string       `json:"merchantRefNum"`
	CustomerProfileID string       `json:"customerProfileId,omitempty"`
	PaymentMethod     string       `json:"paymentMethod"`
	CustomerName      string       `json:"customerName,omitempty"`
	CustomerMobile    string       `json:"customerMobile"`
	CustomerEmail     string       `json:"customerEmail"`
	Amount            float64      `json:"amount"`
	CurrencyCode      string       `json:"currencyCode"`
	PaymentExpiry     int64        `json:"paymentExpiry,omitempty"`
	Description       string       `json:"description"`
	Language          string       `json:"language"`
	ChargeItems       []ChargeItem `json:"chargeItems"`
	OrderWebHookUrl   string       `json:"orderWebHookUrl"`
	Signature         string       `json:"signature"`
}

type ChargeItem struct {
	ItemID      string  `json:"itemId"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

type ChargeResponse struct {
	Type              string  `json:"type"`
	ReferenceNumber   string  `json:"referenceNumber"`
	MerchantRefNumber string  `json:"merchantRefNumber"`
	OrderAmount       float64 `json:"orderAmount"`
	PaymentAmount     float64 `json:"paymentAmount"`
	FawryFees         float64 `json:"fawryFees"`
	PaymentMethod     string  `json:"paymentMethod"`
	OrderStatus       string  `json:"orderStatus"`
	PaymentTime       int64   `json:"paymentTime"`
	CustomerMobile    string  `json:"customerMobile"`
	CustomerMail      string  `json:"customerMail"`
	CustomerProfileID string  `json:"customerProfileId"`
	Signature         string  `json:"signature"`
	StatusCode        int     `json:"statusCode"`
	StatusDescription string  `json:"statusDescription"`
}

type WebhookEvent struct {
	FawryRefNumber    string  `json:"fawryRefNumber"`
	MerchantRefNumber string  `json:"merchantRefNumber"`
	OrderAmount       float64 `json:"orderAmount"`
	PaymentAmount     float64 `json:"paymentAmount"`
	OrderStatus       string  `json:"orderStatus"`
	PaymentMethod     string  `json:"paymentMethod"`
	PaymentTime       int64   `json:"paymentTime"`
	CustomerMobile    string  `json:"customerMobile"`
	CustomerMail      string  `json:"customerMail"`
	CustomerProfileID string  `json:"customerProfileId"`
	StatusCode        int     `json:"statusCode"`
	StatusDescription string  `json:"statusDescription"`
	Signature         string  `json:"signature"`
}

const (
	OrderStatusPaid     = "PAID"
	OrderStatusNew      = "NEW"
	OrderStatusCanceled = "CANCELED"
	OrderStatusRefunded = "REFUNDED"
	OrderStatusExpired  = "EXPIRED"
	OrderStatusFailed   = "FAILED"
)
