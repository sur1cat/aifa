-- Rollback performance indexes

DROP INDEX IF EXISTS idx_habits_user_id;
DROP INDEX IF EXISTS idx_transactions_user_id;
DROP INDEX IF EXISTS idx_transactions_user_date;
DROP INDEX IF EXISTS idx_recurring_transactions_user_id;
DROP INDEX IF EXISTS idx_recurring_transactions_active;
DROP INDEX IF EXISTS idx_tasks_user_id;
DROP INDEX IF EXISTS idx_tasks_user_due_date;
DROP INDEX IF EXISTS idx_habit_completions_habit_date;
DROP INDEX IF EXISTS idx_goals_user_id;
DROP INDEX IF EXISTS idx_savings_goals_user_id;
DROP INDEX IF EXISTS idx_device_tokens_user_id;
DROP INDEX IF EXISTS idx_otp_codes_phone;
