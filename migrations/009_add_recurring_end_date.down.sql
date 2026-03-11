ALTER TABLE recurring_transactions
DROP COLUMN IF EXISTS end_date,
DROP COLUMN IF EXISTS remaining_payments;
