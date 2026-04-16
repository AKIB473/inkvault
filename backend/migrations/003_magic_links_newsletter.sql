-- Migration 003: Magic links and newsletter unsubscribe tokens
-- Learned from Ghost's SingleUseTokenProvider:
--   - validity_period: 24h
--   - validity_period_after_usage: 10min (protects against email client pre-fetchers)
--   - max_usage_count: 7 (aggressive pre-fetchers click multiple times)

-- Single-use magic link tokens (newsletter unsubscribe, email verify, etc.)
CREATE TABLE magic_link_tokens (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    -- SHA-256 hash of raw token (never store raw)
    hash            VARCHAR(64) NOT NULL UNIQUE,
    -- What the token is for
    purpose         VARCHAR(50) NOT NULL, -- 'newsletter_unsubscribe', 'email_verify', 'newsletter_confirm'
    -- Subject (subscriber_id or user_id depending on purpose)
    subject_id      UUID        NOT NULL,
    -- Usage tracking (Ghost allows up to max_usage_count uses)
    usage_count     INT         NOT NULL DEFAULT 0,
    max_usage_count INT         NOT NULL DEFAULT 7,
    -- Timing
    expires_at      TIMESTAMPTZ NOT NULL,
    -- 10-min extended window after first use (email pre-fetch protection)
    first_used_at   TIMESTAMPTZ,
    post_use_expires_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_magic_link_tokens_hash ON magic_link_tokens(hash);
CREATE INDEX idx_magic_link_tokens_subject ON magic_link_tokens(subject_id);
CREATE INDEX idx_magic_link_tokens_expires ON magic_link_tokens(expires_at);

-- Newsletter unsubscribe log (audit trail for compliance)
CREATE TABLE unsubscribe_log (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id         UUID        NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    -- Encrypted email (same pattern as users/subscribers)
    email_encrypted TEXT        NOT NULL,
    reason          VARCHAR(50), -- 'user_request', 'bounce', 'spam_complaint', 'admin'
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_unsubscribe_log_blog ON unsubscribe_log(blog_id);

-- Email suppression list (learned from Ghost's email-suppression-list service)
-- Tracks emails that should NEVER receive mail (bounces + spam complaints)
CREATE TABLE email_suppression (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Encrypted email
    email_encrypted TEXT        NOT NULL UNIQUE,
    reason          VARCHAR(50) NOT NULL, -- 'bounce', 'spam', 'unsubscribe', 'manual'
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_suppression_email ON email_suppression(email_encrypted);
