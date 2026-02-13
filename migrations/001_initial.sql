-- Pixshift initial schema
-- Run: psql $DATABASE_URL < migrations/001_initial.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT UNIQUE NOT NULL,
    password_hash   TEXT,                          -- NULL for OAuth-only users
    name            TEXT NOT NULL DEFAULT '',
    provider        TEXT NOT NULL DEFAULT 'email', -- email, google, github
    tier            TEXT NOT NULL DEFAULT 'free',  -- free, pro
    stripe_customer_id    TEXT,
    stripe_subscription_id TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
    id          TEXT PRIMARY KEY,                  -- random token
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS api_keys (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash    TEXT NOT NULL,                     -- SHA-256 of full key
    prefix      TEXT NOT NULL,                     -- first 8 chars for display
    name        TEXT NOT NULL DEFAULT 'Default',
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);

CREATE TABLE IF NOT EXISTS conversions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    api_key_id      UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    input_format    TEXT NOT NULL,
    output_format   TEXT NOT NULL,
    input_size      BIGINT NOT NULL,
    output_size     BIGINT NOT NULL,
    duration_ms     INT NOT NULL,
    params          JSONB,
    source          TEXT NOT NULL DEFAULT 'api',   -- api, web, mcp
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_conversions_user_id ON conversions(user_id);
CREATE INDEX idx_conversions_created ON conversions(created_at);

CREATE TABLE IF NOT EXISTS daily_usage (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date        DATE NOT NULL DEFAULT CURRENT_DATE,
    count       INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, date)
);
