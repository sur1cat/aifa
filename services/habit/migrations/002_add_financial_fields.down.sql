DROP INDEX IF EXISTS idx_habits_kind;
DROP INDEX IF EXISTS idx_habits_financial_category;

ALTER TABLE habits
    DROP COLUMN IF EXISTS kind,
    DROP COLUMN IF EXISTS expected_amount,
    DROP COLUMN IF EXISTS financial_category,
    DROP COLUMN IF EXISTS currency;
