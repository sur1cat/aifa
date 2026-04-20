-- Performance indexes for common queries

-- Habits by user (used in every habits list query)
CREATE INDEX IF NOT EXISTS idx_habits_user_id ON habits(user_id);

-- Transactions by user (used in every transactions list query)
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);

-- Transactions by user and date (used for monthly queries)
CREATE INDEX IF NOT EXISTS idx_transactions_user_date ON transactions(user_id, date);

-- Recurring transactions by user
CREATE INDEX IF NOT EXISTS idx_recurring_transactions_user_id ON recurring_transactions(user_id);

-- Recurring transactions active lookup (used for processing)
CREATE INDEX IF NOT EXISTS idx_recurring_transactions_active ON recurring_transactions(user_id, is_active) WHERE is_active = true;

-- Tasks by user
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);

-- Tasks by user and due date (used for daily queries)
CREATE INDEX IF NOT EXISTS idx_tasks_user_due_date ON tasks(user_id, due_date);

-- Habit completions lookup (used for streak calculation)
CREATE INDEX IF NOT EXISTS idx_habit_completions_habit_date ON habit_completions(habit_id, completed_date);

-- Goals by user
CREATE INDEX IF NOT EXISTS idx_goals_user_id ON goals(user_id);

-- Savings goals by user
CREATE INDEX IF NOT EXISTS idx_savings_goals_user_id ON savings_goals(user_id);

-- Device tokens by user (for push notifications)
CREATE INDEX IF NOT EXISTS idx_device_tokens_user_id ON device_tokens(user_id);

-- OTP codes lookup (used for verification)
CREATE INDEX IF NOT EXISTS idx_otp_codes_phone ON otp_codes(phone, verified, expires_at);
