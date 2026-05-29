-- Бюджеты по категориям: лимит расходов в месяц.
-- Один бюджет на пользователя + категорию.

CREATE TABLE IF NOT EXISTS budgets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    category    TEXT NOT NULL,
    monthly_limit DECIMAL(12,2) NOT NULL CHECK (monthly_limit > 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, category)
);

CREATE INDEX IF NOT EXISTS idx_budgets_user ON budgets (user_id);
