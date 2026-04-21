DROP INDEX IF EXISTS idx_tasks_category;
DROP INDEX IF EXISTS idx_tasks_kind;

ALTER TABLE tasks
    DROP COLUMN IF EXISTS kind,
    DROP COLUMN IF EXISTS category,
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS amount;
