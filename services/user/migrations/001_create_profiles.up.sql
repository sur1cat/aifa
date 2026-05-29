CREATE TABLE IF NOT EXISTS profiles (
    id          UUID PRIMARY KEY,
    email       TEXT,
    name        TEXT,
    avatar_url  TEXT,
    locale      TEXT NOT NULL DEFAULT 'en',
    timezone    TEXT NOT NULL DEFAULT 'UTC',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_profiles_email ON profiles (email) WHERE email IS NOT NULL;
