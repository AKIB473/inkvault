# 🖋️ InkVault

> **Secure, privacy-first blog publishing platform** — built with Go, Next.js, and a deep respect for your data.

[![CI](https://github.com/AKIB473/inkvault/workflows/CI/badge.svg)](https://github.com/AKIB473/inkvault/actions)
[![Go Version](https://img.shields.io/badge/go-1.22-blue)](https://golang.org)
[![Next.js](https://img.shields.io/badge/next.js-14-black)](https://nextjs.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## What is InkVault?

InkVault is a full-stack, self-hostable blogging platform that puts privacy first. Inspired by deep code study of [Ghost](https://github.com/TryGhost/Ghost) (52k ⭐), [WriteFreely](https://github.com/writefreely/writefreely) (5.1k ⭐), and [Comma](https://github.com/Evernomic/comma) (218 ⭐) — we took the best security and privacy patterns from each.

**Your email is encrypted at rest. There are no trackers. You own your content — export it any time.**

---

## Screenshots

> `make infra && make migrate && make seed && make backend & make frontend`
> 
> Visit http://localhost:3000 · Login: `admin` / `devpassword123`

---

## Features

| Category | Features |
|---|---|
| ✍️ **Writing** | Rich Tiptap editor, Markdown mode, code highlighting, image uploads |
| 📝 **Publishing** | Draft → Publish workflow, scheduled posts, revision history |
| 🔒 **Privacy** | Public / Unlisted / Members-only / Private / Password-protected |
| 🔐 **Security** | AES-256-GCM email encryption, bcrypt-12, JWT+HttpOnly cookies, rate limiting |
| 📡 **Feeds** | RSS 2.0 + Atom 1.0 for every blog |
| 🔍 **Search** | PostgreSQL full-text search with ranking |
| 🌐 **Multi-blog** | Unlimited blogs per account, custom domains, pen names |
| 👥 **Teams** | Owner → Admin → Editor → Writer → Reader roles |
| 📤 **GDPR** | One-click export (JSON), hard-delete with full cascade |
| 🚀 **Performance** | Go backend, Redis caching, Docker multi-stage builds |

---

## Architecture

```
inkvault/
├── backend/                    # Go 1.22 + Fiber v2
│   ├── cmd/server/             # Entry point
│   ├── cmd/migrate/            # Migration runner
│   ├── cmd/seed/               # Dev seed data
│   └── internal/
│       ├── auth/               # JWT, bcrypt, 2FA, invites, password reset
│       ├── crypto/             # AES-256-GCM, HMAC tokens
│       ├── domain/             # Core entities
│       ├── posts/              # Post & blog CRUD
│       ├── rss/                # RSS 2.0 + Atom feeds
│       ├── media/              # S3/MinIO uploads
│       ├── email/              # Resend transactional email
│       ├── health/             # Health check endpoint
│       ├── middleware/         # JWT, rate limit, security headers
│       └── repository/postgres/ # Postgres implementation
├── frontend/                   # Next.js 14 App Router
│   ├── app/
│   │   ├── (auth)/             # Login, register, forgot password
│   │   ├── (dashboard)/        # Dashboard, settings
│   │   ├── (editor)/           # Tiptap post editor
│   │   └── (reader)/           # Public blog views
│   ├── components/
│   │   ├── ui/                 # Design system (Button, Input, Badge, Card)
│   │   ├── editor/             # Tiptap + toolbar + image upload
│   │   └── blog/               # Blog-specific components
│   └── lib/                    # API client, utils
├── docker-compose.yml          # Dev infrastructure
├── docker-compose.prod.yml     # Production stack (with Traefik + Let's Encrypt)
└── .github/workflows/          # CI + Release pipelines
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.22 + Fiber v2 |
| Frontend | Next.js 14 + TypeScript + Tailwind CSS |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Editor | Tiptap v2 |
| Media | MinIO (S3-compatible) |
| Email | Resend API |
| Container | Docker + Traefik |
| CI/CD | GitHub Actions |

---

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.22+ (for local dev)
- Node.js 20+ (for local dev)

### 1. Clone
```bash
git clone https://github.com/AKIB473/inkvault
cd inkvault
```

### 2. Configure
```bash
cp backend/.env.example backend/.env
# Edit backend/.env — at minimum set:
#   EMAIL_ENCRYPTION_KEY (32 random bytes, hex-encoded)
#   JWT_SECRET (32+ random chars)
```

Generate secure keys:
```bash
openssl rand -hex 32   # EMAIL_ENCRYPTION_KEY
openssl rand -base64 32 # JWT_SECRET
```

### 3. Start
```bash
make infra      # Postgres + Redis + MinIO
make migrate    # Run 4 SQL migrations
make seed       # Create admin user + sample blog
make backend    # API on :8080
make frontend   # UI on :3000
```

Visit **http://localhost:3000**  
Login: `admin` / `devpassword123`

---

## Security Architecture

| Concern | Implementation | Source |
|---|---|---|
| Passwords | bcrypt cost-12 | OWASP |
| Email at rest | AES-256-GCM, key in env | WriteFreely |
| Access tokens | JWT HS256, 15min TTL | Ghost |
| Refresh tokens | SHA-256 hashed, HttpOnly cookie, 7d | Ghost |
| Rate limiting | Redis INCR/EXPIRE, 5/hr/IP on auth | Ghost |
| Reset tokens | HMAC-signed, embeds pw hash (auto-invalidates) | Ghost |
| Spam prevention | Honeypot field (invisible to humans) | WriteFreely |
| Magic links | 7-use max (email pre-fetch protection) | Ghost |
| Audit log | Every write: actor, action, IP | Ghost |
| Hard delete | FK CASCADE, no soft delete | WriteFreely |
| HTTP headers | HSTS, CSP, X-Frame-Options, Referrer-Policy | OWASP |
| Docker | `scratch` image, uid 65534 (nobody) | Best practice |

---

## API Reference

```
POST   /api/v1/auth/register            Register (with honeypot)
POST   /api/v1/auth/login               Login → JWT + HttpOnly cookie
POST   /api/v1/auth/refresh             Rotate refresh token
POST   /api/v1/auth/logout              Clear session
POST   /api/v1/auth/logout-all          Revoke all devices
POST   /api/v1/auth/forgot-password     Always returns 200 (no user enumeration)
POST   /api/v1/auth/reset-password      HMAC token validation
GET    /api/v1/auth/username-check      Live availability check

GET    /api/v1/blogs/:slug              Public blog info
GET    /api/v1/blogs/:slug/posts        Blog's published posts
GET    /api/v1/blogs/:slug/posts/:slug  Single post (visibility checked)
GET    /api/v1/blogs/:slug/feed.xml     RSS 2.0 feed
GET    /api/v1/blogs/:slug/atom.xml     Atom feed
GET    /api/v1/search?q=               Full-text search

POST   /api/v1/blogs                   Create blog [writer+]
POST   /api/v1/posts                   Create post [writer+]
PATCH  /api/v1/posts/:id               Update post [writer+]
DELETE /api/v1/posts/:id               Delete post [writer+]
GET    /api/v1/posts/:id/revisions     Post history [auth]
POST   /api/v1/media                   Upload media [writer+]

GET    /api/v1/me/blogs                My blogs
GET    /api/v1/me/posts                My posts
GET    /api/v1/me/export/posts.json    GDPR export

GET    /api/v1/admin/users             User list [admin+]
GET    /api/v1/admin/audit             Audit log [admin+]
GET    /api/v1/health                  Health check
```

---

## Production Deploy

```bash
# Set required env vars
export POSTGRES_PASSWORD=strong_random_password
export REDIS_PASSWORD=strong_random_password
export ACME_EMAIL=you@example.com

# Deploy (builds images, runs migrations, starts Traefik with auto-SSL)
docker compose -f docker-compose.prod.yml up -d
```

Traefik auto-provisions Let's Encrypt SSL. Internal network isolation keeps Postgres and Redis unreachable from outside.

---

## Running Tests

```bash
cd backend

# All tests
go test ./...

# With race detector (recommended — run twice to catch flaky tests)
go test -race -count=2 ./...

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Test suites cover:
- `crypto/` — AES-GCM roundtrip, nonce randomness, tamper detection, HMAC tokens
- `domain/` — User status bitmask, role checks, token validity, public serialization
- `auth/` — Username normalization, reserved names, 30-char limit
- `posts/` — Slugify, reading time, tag deduplication, visibility parsing
- `rss/` — XML escaping, double-encode prevention
- `users/` — CanPerform() permission matrix, magic link config
- `media/` — MIME sniffing, extension mapping, size limits, allowlist
- `apierr/` — Status codes, WithMessage immutability

---

## License

MIT — see [LICENSE](LICENSE)

---

## Acknowledgements

Built by studying the best open-source platforms:
- **[Ghost](https://github.com/TryGhost/Ghost)** — Security patterns, permission model, token design
- **[WriteFreely](https://github.com/writefreely/writefreely)** — Email encryption, privacy-first DB design
- **[Comma](https://github.com/Evernomic/comma)** — Modern Next.js patterns, Tiptap editor, SEO fields
