package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/omar-shahieen/moneyflow/internal/lib/billing"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type BillingEvent struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	FawryRefNumber    string          `json:"fawry_ref_number" db:"fawry_ref_number"`
	MerchantRefNumber string          `json:"merchant_ref_number" db:"merchant_ref_number"`
	OrderStatus       string          `json:"order_status" db:"order_status"`
	Payload           json.RawMessage `json:"payload" db:"payload"`
	ProcessedAt       time.Time       `json:"processed_at" db:"processed_at"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
}

type BillingEventRepo struct {
	server *server.Server
}

func NewBillingEventRepository(server *server.Server) *BillingEventRepo {
	return &BillingEventRepo{server: server}
}

// RecordEvent records a webhook event into the billing_events idempotency ledger.
// Returns created=false if the event was already recorded (duplicate delivery).
func (r *BillingEventRepo) RecordEvent(ctx context.Context, e *billing.WebhookEvent) (bool, error) {
	payloadBytes, err := json.Marshal(e)
	if err != nil {
		return false, fmt.Errorf("failed to marshal webhook event payload: %w", err)
	}

	stmt := `
		INSERT INTO billing_events (id, fawry_ref_number, merchant_ref_number, order_status, payload, processed_at, created_at)
		VALUES (gen_random_uuid(), @fawry_ref_number, @merchant_ref_number, @order_status, @payload, NOW(), NOW())
		ON CONFLICT (fawry_ref_number) DO NOTHING
		RETURNING id
	`

	var eventID uuid.UUID
	err = r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"fawry_ref_number":    e.FawryRefNumber,
		"merchant_ref_number": e.MerchantRefNumber,
		"order_status":        e.OrderStatus,
		"payload":             payloadBytes,
	}).Scan(&eventID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Already exists, duplicate event
			return false, nil
		}
		return false, fmt.Errorf("failed to insert billing event: %w", err)
	}

	return true, nil
}

// GetByMerchantRef fetches the latest billing event for a merchant reference number
func (r *BillingEventRepo) GetByMerchantRef(ctx context.Context, merchantRefNum string) (*BillingEvent, error) {
	stmt := `
		SELECT
			id, fawry_ref_number, merchant_ref_number, order_status, payload, processed_at, created_at
		FROM
			billing_events
		WHERE
			merchant_ref_number = @merchant_ref_number
		ORDER BY
			created_at DESC
		LIMIT 1
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"merchant_ref_number": merchantRefNum,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get billing event by merchant_ref_number=%s: %w", merchantRefNum, err)
	}

	event, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[BillingEvent])
	if err != nil {
		return nil, fmt.Errorf("failed to collect billing event: %w", err)
	}

	return &event, nil
}

// GetByFawryRef fetches a billing event by Fawry reference number
func (r *BillingEventRepo) GetByFawryRef(ctx context.Context, fawryRefNum string) (*BillingEvent, error) {
	stmt := `
		SELECT
			id, fawry_ref_number, merchant_ref_number, order_status, payload, processed_at, created_at
		FROM
			billing_events
		WHERE
			fawry_ref_number = @fawry_ref_number
		LIMIT 1
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"fawry_ref_number": fawryRefNum,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get billing event by fawry_ref_number=%s: %w", fawryRefNum, err)
	}

	event, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[BillingEvent])
	if err != nil {
		return nil, fmt.Errorf("failed to collect billing event: %w", err)
	}

	return &event, nil
}
