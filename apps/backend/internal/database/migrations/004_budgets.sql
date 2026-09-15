CREATE TABLE IF NOT EXISTS budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    monthly_limit_minor BIGINT NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'EGP',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS budget_members (
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('owner', 'member')),
    PRIMARY KEY (budget_id, user_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS one_owner_per_budget ON budget_members(budget_id) WHERE role = 'owner';
CREATE INDEX IF NOT EXISTS idx_budgets_category ON budgets(category_id);
CREATE INDEX IF NOT EXISTS idx_budget_members_user ON budget_members(user_id);
CREATE INDEX IF NOT EXISTS idx_budget_members_budget ON budget_members(budget_id);
