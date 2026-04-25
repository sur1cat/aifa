-- Minimal identity record owned by auth-service.
-- Profile data (email, name, avatar) lives in user-service and is
-- propagated there via the user.provisioned NATS event.
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auth_provider   TEXT NOT NULL,
    provider_id     TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_provider_unique UNIQUE (auth_provider, provider_id)
);

CREATE INDEX IF NOT EXISTS idx_users_provider ON users (auth_provider, provider_id);
