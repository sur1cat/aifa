-- Tasks become financial actions: bill reminders, payment due dates,
-- invoice send-outs, subscription cancellations. The new fields are
-- optional — pure to-do tasks still work.

ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS amount   DECIMAL(12,2),
    ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT 'USD',
    ADD COLUMN IF NOT EXISTS category TEXT,
    ADD COLUMN IF NOT EXISTS kind     TEXT NOT NULL DEFAULT 'todo';

-- kind enum semantics:
--   'todo'    — generic action item; amount/category may still be set
--               for context but no transaction implication
--   'bill'    — money the user OWES (rent, electricity, subscription).
--               amount > 0; on completion the iOS client may emit a
--               matching expense transaction.
--   'income'  — money the user is OWED (invoice, payday). amount > 0;
--               completion may yield an income transaction.

CREATE INDEX IF NOT EXISTS idx_tasks_kind ON tasks (kind);
CREATE INDEX IF NOT EXISTS idx_tasks_category ON tasks (category) WHERE category IS NOT NULL;
