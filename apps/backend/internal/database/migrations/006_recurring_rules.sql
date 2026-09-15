CREATE TABLE IF NOT EXISTS recurring_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    amount_minor BIGINT NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'EGP',
    frequency TEXT NOT NULL CHECK (frequency IN ('weekly', 'monthly')),
    next_run_date DATE NOT NULL,
    last_generated_date DATE,
    end_date DATE
);

CREATE INDEX IF NOT EXISTS idx_recurring_rules_user ON recurring_rules(user_id);
CREATE INDEX IF NOT EXISTS idx_recurring_rules_user_next_run ON recurring_rules(user_id, next_run_date);

CREATE TABLE IF NOT EXISTS recurring_rule_dedup (
    rule_id UUID NOT NULL REFERENCES recurring_rules(id) ON DELETE CASCADE,
    occurrence_date DATE NOT NULL,
    PRIMARY KEY (rule_id, occurrence_date)
);
