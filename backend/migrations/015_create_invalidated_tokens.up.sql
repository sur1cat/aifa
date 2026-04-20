-- Invalidated refresh tokens (for logout)
CREATE TABLE invalidated_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    token_hash VARCHAR(64) NOT NULL UNIQUE, -- SHA256 hash of the token
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL, -- When the original token expires
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invalidated_tokens_hash ON invalidated_tokens(token_hash);
CREATE INDEX idx_invalidated_tokens_expires ON invalidated_tokens(expires_at);

-- Cleanup job: DELETE FROM invalidated_tokens WHERE expires_at < NOW();
