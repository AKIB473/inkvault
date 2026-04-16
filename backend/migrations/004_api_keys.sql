-- Migration 004: API Keys for headless/integration access
-- Learned from Ghost's Content API key vs Admin API key split.
--
-- Content API key: read-only public key (headless frontends, RSS readers)
-- Admin API key:   read/write for integrations and automations
--
-- Ghost's authorize.js:
--   Content API: valid if has api_key OR member session cookie
--   Admin API:   valid if has user session OR api_key

CREATE TYPE api_key_type AS ENUM ('content', 'admin');

CREATE TABLE api_keys (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id     UUID         NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    type        api_key_type NOT NULL DEFAULT 'content',
    -- SHA-256 hash of the raw key — raw key shown once at creation
    key_hash    VARCHAR(64)  NOT NULL UNIQUE,
    -- Human-readable label for the integration
    label       VARCHAR(255) NOT NULL DEFAULT 'Untitled Integration',
    last_used_at TIMESTAMPTZ,
    created_by  UUID         REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_blog_id ON api_keys(blog_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_type ON api_keys(type);

-- Webhooks — outbound event notifications for integrations
-- (Ghost has a full webhooks service)
CREATE TABLE webhooks (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id         UUID        NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    event           VARCHAR(100) NOT NULL, -- 'post.published', 'subscriber.added', etc.
    target_url      TEXT        NOT NULL,
    -- Secret for HMAC-SHA256 signature on webhook payload
    secret_hash     VARCHAR(64) NOT NULL,
    active          BOOLEAN     NOT NULL DEFAULT TRUE,
    last_triggered_at TIMESTAMPTZ,
    last_status_code  INT,
    created_by      UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_blog_id ON webhooks(blog_id);
CREATE INDEX idx_webhooks_event ON webhooks(event);
CREATE INDEX idx_webhooks_active ON webhooks(active);

CREATE TRIGGER webhooks_updated_at
    BEFORE UPDATE ON webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
