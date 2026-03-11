CREATE TABLE recurring_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    amount DECIMAL(12, 2) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'expense', -- income, expense
    category VARCHAR(100) NOT NULL DEFAULT '',
    frequency VARCHAR(50) NOT NULL DEFAULT 'monthly', -- weekly, biweekly, monthly, quarterly, yearly
    start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    next_date DATE NOT NULL DEFAULT CURRENT_DATE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recurring_transactions_user ON recurring_transactions(user_id);
CREATE INDEX idx_recurring_transactions_next_date ON recurring_transactions(next_date);
CREATE INDEX idx_recurring_transactions_active ON recurring_transactions(is_active);
