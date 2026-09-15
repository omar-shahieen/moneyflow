-- Transactions: link back to the billing event that created them
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS billing_provider TEXT DEFAULT 'manual',
    ADD COLUMN IF NOT EXISTS billing_ref_id TEXT,
    ADD COLUMN IF NOT EXISTS billing_status TEXT DEFAULT 'manual';

CREATE INDEX IF NOT EXISTS idx_transactions_billing_ref ON transactions(billing_ref_id);

-- Subscriptions: pending-expiry + end-of-period cancellation + dunning state
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_status_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_status_check 
    CHECK (status IN ('incomplete', 'pending', 'active', 'past_due', 'cancelled', 'canceled', 'expired'));

ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS failed_attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS next_billing_at TIMESTAMPTZ;

-- Partial index: reconciliation/dunning jobs only ever scan non-terminal rows
CREATE INDEX IF NOT EXISTS idx_subscriptions_pending
    ON subscriptions(status)
    WHERE status IN ('pending', 'past_due');

-- Billing events: append-only idempotency ledger + audit trail.
CREATE TABLE IF NOT EXISTS billing_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fawry_ref_number    TEXT NOT NULL UNIQUE,
    merchant_ref_number TEXT NOT NULL,
    order_status        TEXT NOT NULL,
    payload             JSONB NOT NULL,
    processed_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_billing_events_merchant_ref ON billing_events(merchant_ref_number);
