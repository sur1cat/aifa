CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255),
    phone VARCHAR(20),
    name VARCHAR(255),
    avatar_url TEXT,
    auth_provider VARCHAR(20) NOT NULL, -- google, apple, phone
    provider_id VARCHAR(255) NOT NULL,  -- ID from provider (Google sub, Apple sub, phone number)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_provider_id UNIQUE (auth_provider, provider_id)
);

CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_phone ON users(phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_users_provider ON users(auth_provider, provider_id);

CREATE TABLE otp_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20) NOT NULL,
    code VARCHAR(6) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_phone ON otp_codes(phone, verified);
CREATE INDEX idx_otp_expires ON otp_codes(expires_at);

