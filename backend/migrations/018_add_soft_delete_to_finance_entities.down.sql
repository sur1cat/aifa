DROP INDEX IF EXISTS idx_transactions_deleted_at;
DROP INDEX IF EXISTS idx_recurring_transactions_deleted_at;

ALTER TABLE transactions
    DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE recurring_transactions
    DROP COLUMN IF EXISTS deleted_at;
