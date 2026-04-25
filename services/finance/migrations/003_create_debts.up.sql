-- Долги: я должен кому-то (i_owe) или мне должны (they_owe).

CREATE TABLE IF NOT EXISTS debts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL,
    counterparty TEXT NOT NULL,
    direction    TEXT NOT NULL CHECK (direction IN ('i_owe', 'they_owe')),
    amount       DECIMAL(12,2) NOT NULL CHECK (amount >= 0),
    original_amount DECIMAL(12,2) NOT NULL CHECK (original_amount > 0),
    note         TEXT,
    settled      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_debts_user ON debts (user_id);
CREATE INDEX IF NOT EXISTS idx_debts_user_active ON debts (user_id) WHERE NOT settled;
