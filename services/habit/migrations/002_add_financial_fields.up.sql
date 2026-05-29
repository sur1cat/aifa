-- Habits become financial habits: tracking, saving, or spending-control
-- behaviors with budgetary impact. Existing rows keep working: new columns
-- are optional (nullable / sensible defaults).

ALTER TABLE habits
    ADD COLUMN IF NOT EXISTS currency           TEXT NOT NULL DEFAULT 'USD',
    ADD COLUMN IF NOT EXISTS financial_category TEXT,
    ADD COLUMN IF NOT EXISTS expected_amount    DECIMAL(12,2),
    ADD COLUMN IF NOT EXISTS kind               TEXT NOT NULL DEFAULT 'tracking';

-- kind enum semantics:
--   'tracking'    — habit improves financial awareness (no direct money impact)
--   'saving'      — completion adds expected_amount to savings (positive impact)
--   'spending'    — completion controls/limits an expense category (the value
--                   represents the budget cap for the period; staying within
--                   counts as completion)

CREATE INDEX IF NOT EXISTS idx_habits_financial_category ON habits (financial_category)
    WHERE financial_category IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_habits_kind ON habits (kind);
