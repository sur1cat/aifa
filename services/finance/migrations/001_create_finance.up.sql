-- Consolidated finance schema: transactions, recurring_transactions, savings_goals.
-- No cross-schema FK to users (cleanup via user.deleted NATS event).

CREATE TABLE IF NOT EXISTS transactions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    title       TEXT NOT NULL,
    amount      DECIMAL(12,2) NOT NULL,
    type        TEXT NOT NULL DEFAULT 'expense',
    category    TEXT NOT NULL DEFAULT '',
    date        DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_user ON transactions (user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions (date);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions (type);

CREATE TABLE IF NOT EXISTS recurring_transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL,
    title               TEXT NOT NULL,
    amount              DECIMAL(12,2) NOT NULL,
    type                TEXT NOT NULL DEFAULT 'expense',
    category            TEXT NOT NULL DEFAULT '',
    frequency           TEXT NOT NULL DEFAULT 'monthly',
    start_date          DATE NOT NULL DEFAULT CURRENT_DATE,
    next_date           DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date            DATE,
    remaining_payments  INTEGER,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recurring_user ON recurring_transactions (user_id);
CREATE INDEX IF NOT EXISTS idx_recurring_next_date ON recurring_transactions (next_date);
CREATE INDEX IF NOT EXISTS idx_recurring_active ON recurring_transactions (is_active);

CREATE TABLE IF NOT EXISTS savings_goals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE,
    monthly_target  DECIMAL(12,2) NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_savings_goals_user ON savings_goals (user_id);
