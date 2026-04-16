-- Migration 002: Blogs (Collections) and Posts
-- Blog = Collection pattern from WriteFreely: each user can have multiple blogs
-- Post visibility bitmask also from WriteFreely

-- Post visibility enum
-- 0=public, 1=unlisted, 2=members_only, 3=private, 4=password_protected
CREATE TYPE post_visibility AS ENUM ('public', 'unlisted', 'members_only', 'private', 'password_protected');

-- Post status enum
CREATE TYPE post_status AS ENUM ('draft', 'scheduled', 'published', 'archived');

-- Blogs table (= Collections in WriteFreely)
-- Each user can have multiple blogs (pen names supported)
CREATE TABLE blogs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id        UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slug            VARCHAR(100)    NOT NULL UNIQUE,
    title           VARCHAR(255)    NOT NULL,
    description     TEXT            NOT NULL DEFAULT '',
    -- Custom domain (e.g. myblog.com) — from Comma
    domain          VARCHAR(255)    UNIQUE,
    visibility      post_visibility NOT NULL DEFAULT 'public',
    -- bcrypt hash for password-protected blogs
    password_hash   VARCHAR(255),
    style_sheet     TEXT,           -- Custom CSS
    language        VARCHAR(10)     NOT NULL DEFAULT 'en',
    rtl             BOOLEAN         NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_blogs_owner_id ON blogs(owner_id);
CREATE INDEX idx_blogs_slug ON blogs(slug);
CREATE INDEX idx_blogs_domain ON blogs(domain);
CREATE INDEX idx_blogs_visibility ON blogs(visibility);

CREATE TRIGGER blogs_updated_at
    BEFORE UPDATE ON blogs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Posts table
CREATE TABLE posts (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id         UUID            NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    author_id       UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           VARCHAR(500)    NOT NULL DEFAULT '',
    slug            VARCHAR(500)    NOT NULL,
    excerpt         TEXT            NOT NULL DEFAULT '',
    -- Tiptap JSON or raw Markdown
    content         TEXT            NOT NULL DEFAULT '',
    content_type    VARCHAR(20)     NOT NULL DEFAULT 'tiptap', -- 'tiptap' | 'markdown'
    status          post_status     NOT NULL DEFAULT 'draft',
    visibility      post_visibility NOT NULL DEFAULT 'public',
    password_hash   VARCHAR(255),   -- For password-protected posts
    is_pinned       BOOLEAN         NOT NULL DEFAULT FALSE,
    -- Tags stored as text array (Comma pattern — simple and fast for FTS)
    tags            TEXT[]          NOT NULL DEFAULT '{}',
    cover_image     TEXT,
    -- SEO fields (from Comma)
    seo_title       VARCHAR(255),
    seo_description VARCHAR(500),
    og_image        TEXT,
    canonical_url   TEXT,
    -- Metrics
    views           BIGINT          NOT NULL DEFAULT 0,
    reading_time_mins INT           NOT NULL DEFAULT 0,
    -- Timestamps
    published_at    TIMESTAMPTZ,
    scheduled_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    -- Newsletter tracking (from Comma)
    last_newsletter_sent_at TIMESTAMPTZ,

    UNIQUE(blog_id, slug)
);

CREATE INDEX idx_posts_blog_id ON posts(blog_id);
CREATE INDEX idx_posts_author_id ON posts(author_id);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_visibility ON posts(visibility);
CREATE INDEX idx_posts_published_at ON posts(published_at DESC);
CREATE INDEX idx_posts_tags ON posts USING GIN(tags);
-- Full-text search index (learned from Ghost + GoBlog FTS5 pattern)
CREATE INDEX idx_posts_fts ON posts
    USING GIN(to_tsvector('english', coalesce(title,'') || ' ' || coalesce(excerpt,'') || ' ' || coalesce(content,'')));

CREATE TRIGGER posts_updated_at
    BEFORE UPDATE ON posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Post revisions — full history (from Ghost + goblog)
CREATE TABLE post_revisions (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID        NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    content     TEXT        NOT NULL,
    edited_by   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_post_revisions_post_id ON post_revisions(post_id);
CREATE INDEX idx_post_revisions_created_at ON post_revisions(created_at DESC);

-- Newsletter subscribers (from Comma)
CREATE TABLE subscribers (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    blog_id         UUID        NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    -- Encrypted email (same pattern as users.email_encrypted)
    email_encrypted TEXT        NOT NULL,
    confirmed       BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(blog_id, email_encrypted)
);

CREATE INDEX idx_subscribers_blog_id ON subscribers(blog_id);

-- Media uploads
CREATE TABLE media (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    uploader_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blog_id         UUID        REFERENCES blogs(id) ON DELETE SET NULL,
    filename        VARCHAR(255) NOT NULL,
    mime_type       VARCHAR(100) NOT NULL,
    size_bytes      BIGINT      NOT NULL,
    -- S3/MinIO object key
    storage_key     TEXT        NOT NULL,
    -- Public URL (CDN or direct)
    public_url      TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_uploader_id ON media(uploader_id);
CREATE INDEX idx_media_blog_id ON media(blog_id);
