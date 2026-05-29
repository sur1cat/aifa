-- Goals become financial: savings, debt-payoff, purchase, investment.
-- The legacy `target_value INTEGER` (counts of arbitrary units) is
-- replaced by `target_amount DECIMAL(12,2)` (monetary). Existing rows
-- get their target_value copied into target_amount with currency = USD.
-- After the migration the legacy column is dropped to avoid drift.

ALTER TABLE goals
    ADD COLUMN IF NOT EXISTS target_amount  DECIMAL(12,2),
    ADD COLUMN IF NOT EXISTS current_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS currency       TEXT NOT NULL DEFAULT 'USD',
    ADD COLUMN IF NOT EXISTS goal_type      TEXT NOT NULL DEFAULT 'savings';

-- Promote the integer target_value into target_amount when the row
-- looks numeric and the new column is still NULL.
UPDATE goals
SET target_amount = target_value::DECIMAL
WHERE target_amount IS NULL AND target_value IS NOT NULL;

-- The legacy generic `unit` (e.g. "books", "kg") doesn't make sense
-- for money — drop it. `target_value` likewise: superseded by
-- target_amount + currency.
ALTER TABLE goals
    DROP COLUMN IF EXISTS unit,
    DROP COLUMN IF EXISTS target_value;

-- goal_type enum semantics (kept as TEXT for forward extensibility):
--   'savings'    — accumulate money toward a number (emergency fund,
--                  vacation, big purchase). Progress = current_amount
--                  growing toward target_amount.
--   'debt'       — pay down a balance. current_amount tracks how much
--                  has been paid; reaching target_amount = debt cleared.
--   'purchase'   — discrete purchase to be made when current_amount
--                  reaches target_amount.
--   'investment' — track money put into investments. target_amount is
--                  the contribution target (not market value).

CREATE INDEX IF NOT EXISTS idx_goals_type ON goals (goal_type);
