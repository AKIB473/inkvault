#!/bin/bash
# publish-github.sh — Initialize git repo and publish to GitHub
# Usage: bash scripts/publish-github.sh [repo-name]
# Requires: gh CLI authenticated

set -euo pipefail
cd "$(dirname "$0")/.."

REPO_NAME="${1:-inkvault}"
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

echo -e "${YELLOW}Publishing InkVault to GitHub as: ${REPO_NAME}${NC}"

# ── Git init ──────────────────────────────
if [ ! -d ".git" ]; then
  git init
  git branch -M main
  echo "git init done"
fi

# ── Add all files ─────────────────────────
git add -A
git status --short | head -30

# ── Commit ────────────────────────────────
git commit -m "feat: initial InkVault production release

A secure, privacy-first blog publishing platform.

## Stack
- Backend: Go 1.22 + Fiber v2
- Frontend: Next.js 14 + TypeScript
- Database: PostgreSQL 16
- Cache: Redis 7
- Editor: Tiptap v2
- Media: S3/MinIO
- Containerization: Docker + Docker Compose

## Security & Privacy
- AES-256-GCM encrypted emails at rest
- bcrypt cost-12 password hashing
- JWT (15min) + rotating HttpOnly refresh tokens (7d)
- Redis-backed rate limiting (5 auth/hr/IP)
- Honeypot spam prevention
- Single-use HMAC-signed tokens
- Hard delete (GDPR cascade)
- Role system: owner > admin > editor > writer > reader
- Audit log on all writes
- Security headers: HSTS, CSP, X-Frame-Options

## Features
- Multi-blog per account (pen names)
- Rich Tiptap editor + Markdown mode
- Post visibility: public/unlisted/members/private/password
- RSS 2.0 + Atom feeds
- Full-text search (Postgres FTS)
- Post revision history
- Media uploads (S3/MinIO)
- Email 2FA (OTP)
- Data export (GDPR)
- ActivityPub ready

Research-backed from Ghost (52k★), WriteFreely (5.1k★), Comma (218★)" \
  || git commit --allow-empty -m "chore: ci trigger"

# ── Create GitHub repo ────────────────────
echo "Creating GitHub repository: ${REPO_NAME}"
gh repo create "${REPO_NAME}" \
  --public \
  --description "🖋️ A secure, privacy-first blog & article publishing platform. Go + Next.js." \
  --homepage "https://github.com/$(gh api user --jq .login)/${REPO_NAME}" \
  2>/dev/null || echo "Repo may already exist, continuing..."

# ── Add remote & push ─────────────────────
GITHUB_USER=$(gh api user --jq .login)
REMOTE="https://github.com/${GITHUB_USER}/${REPO_NAME}.git"

git remote remove origin 2>/dev/null || true
git remote add origin "${REMOTE}"

echo "Pushing to ${REMOTE}..."
git push -u origin main --force

echo ""
echo -e "${GREEN}✅ Published! https://github.com/${GITHUB_USER}/${REPO_NAME}${NC}"

# ── Add topics ────────────────────────────
gh repo edit "${REPO_NAME}" \
  --add-topic "blog" \
  --add-topic "golang" \
  --add-topic "nextjs" \
  --add-topic "privacy" \
  --add-topic "security" \
  --add-topic "self-hosted" \
  --add-topic "postgresql" \
  --add-topic "typescript" \
  2>/dev/null || true

echo -e "${GREEN}✅ Done! Check: https://github.com/${GITHUB_USER}/${REPO_NAME}${NC}"
