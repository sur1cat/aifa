-- Add end_date and remaining_payments to recurring_transactions
ALTER TABLE recurring_transactions
ADD COLUMN end_date DATE DEFAULT NULL,
ADD COLUMN remaining_payments INTEGER DEFAULT NULL;

-- Add comment
COMMENT ON COLUMN recurring_transactions.end_date IS 'Optional end date for loans/subscriptions';
COMMENT ON COLUMN recurring_transactions.remaining_payments IS 'Optional remaining payments count for loans';
