-- Migration 001: Users, Roles, Sessions
-- Designed with privacy-first principles:
-- - Emails encrypted at rest (stored as hex ciphertext, never plaintext)
-- - bcrypt password hashes only
-- - Hard delete support (no soft-delete for GDPR compliance)

-- Roles enum
CREATE TYPE user_role AS ENUM ('owner', 'admin', 'editor', 'writer', 'reader');

-- Users table
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(30)  NOT NULL UNIQUE,
    display_name    VARCHAR(100) NOT NULL DEFAULT '',
    -- AES-256-GCM encrypted email (hex-encoded ciphertext) — NEVER plaintext
    email_encrypted TEXT         NOT NULL UNIQUE,
    -- bcrypt hash (cost 12), never raw password
    password_hash   VARCHAR(255) NOT NULL,
    role            user_role    NOT NULL DEFAULT 'writer',
    -- Bitmask: 0=active, 1=silenced, 2=banned
    status          SMALLINT     NOT NULL DEFAULT 0,
    two_fa_enabled  BOOLEAN      NOT NULL DEFAULT FALSE,
    avatar_url      TEXT,
    bio             TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
    -- No deleted_at: we HARD DELETE for GDPR compliance
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);

-- Sessions table (refresh tokens)
-- Stores device info for suspicious login detection (Ghost pattern)
CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- SHA-256 of refresh token — raw token never stored
    refresh_hash    VARCHAR(64)  NOT NULL UNIQUE,
    device_info     VARCHAR(50)  NOT NULL DEFAULT 'Unknown',
    ip_address      INET         NOT NULL,
    user_agent      TEXT,
    last_seen_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ  NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_hash ON sessions(refresh_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Single-use tokens (password reset, email verify, invites, 2FA)
-- Ghost: tokens are base64+HMAC, stored as hash only
CREATE TYPE token_type AS ENUM (
    'password_reset',
    'email_verify',
    'invite',
    'two_fa',
    'delete_confirm'
);

CREATE TABLE tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- SHA-256 of raw token — raw token returned once at creation, never stored
    hash            VARCHAR(64)  NOT NULL UNIQUE,
    user_id         UUID         REFERENCES users(id) ON DELETE CASCADE,
    type            token_type   NOT NULL,
    used_at         TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ  NOT NULL,
    -- For invite tokens
    invited_email   TEXT,
    created_by      UUID         REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tokens_hash ON tokens(hash);
CREATE INDEX idx_tokens_user_id ON tokens(user_id);
CREATE INDEX idx_tokens_type ON tokens(type);
CREATE INDEX idx_tokens_expires_at ON tokens(expires_at);

-- Audit log — every significant write action recorded
-- Learned from Ghost's security-first approach
CREATE TABLE audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id        UUID         REFERENCES users(id) ON DELETE SET NULL,
    actor_role      user_role,
    action          VARCHAR(100) NOT NULL, -- e.g. "post.publish", "user.delete"
    resource_type   VARCHAR(50),           -- "post", "user", "blog"
    resource_id     UUID,
    ip_address      INET,
    user_agent      TEXT,
    meta            JSONB,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_actor_id ON audit_log(actor_id);
CREATE INDEX idx_audit_action ON audit_log(action);
CREATE INDEX idx_audit_resource ON audit_log(resource_type, resource_id);
CREATE INDEX idx_audit_created_at ON audit_log(created_at);

-- updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
