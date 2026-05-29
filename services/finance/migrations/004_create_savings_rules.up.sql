-- Правила автоматических накоплений.
-- period: 'monthly' | 'on_income'
-- alert_rules: суточные/месячные лимиты предупреждений

CREATE TABLE IF NOT EXISTS savings_rules (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL,
    kind       TEXT NOT NULL CHECK (kind IN ('monthly_savings', 'on_income_savings', 'spending_alert')),
    amount     DECIMAL(12,2) NOT NULL CHECK (amount > 0),
    period     TEXT,           -- monthly | on_income | daily
    goal_title TEXT,
    active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_savings_rules_user ON savings_rules (user_id);
CREATE INDEX IF NOT EXISTS idx_savings_rules_active ON savings_rules (user_id, kind) WHERE active;
