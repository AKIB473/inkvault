#!/bin/bash
set -euo pipefail

INKVAULT=/root/.openclaw/workspace/inkvault
BACKEND=$INKVAULT/backend
FRONTEND=$INKVAULT/frontend

echo ""
echo "════════════════════════════════════════"
echo "  InkVault Build & Deploy Script"
echo "════════════════════════════════════════"
echo ""

# ── STEP 1: Backend tests ────────────────────────────────────────────────────
echo "=== STEP 1: Backend tests ==="
cd $BACKEND
go clean -testcache
go test -race -count=1 ./... 2>&1
echo "✅ All backend tests passed"

# ── STEP 3: Build Go binaries ────────────────────────────────────────────────
echo ""
echo "=== STEP 3: Build Go binaries ==="
CGO_ENABLED=0 go build -ldflags="-w -s" -trimpath -o /tmp/final-server ./cmd/server
echo "  ✅ /tmp/final-server built"
CGO_ENABLED=0 go build -ldflags="-w -s" -trimpath -o /tmp/final-migrate ./cmd/migrate
echo "  ✅ /tmp/final-migrate built"

# ── STEP 5: Frontend install and build ──────────────────────────────────────
echo ""
echo "=== STEP 5: Frontend install and build ==="
cd $FRONTEND
npm install 2>&1 | tail -5
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1 \
  NEXT_PUBLIC_APP_URL=http://localhost:3000 \
  NEXT_PUBLIC_APP_NAME=InkVault \
  npm run build 2>&1
echo "✅ Frontend build complete"

# ── STEP 6: Backend tests again ──────────────────────────────────────────────
echo ""
echo "=== STEP 6: Backend tests (second run) ==="
cd $BACKEND
go test -race -count=1 ./... 2>&1
echo "✅ All backend tests stable"

# ── STEP 7: Commit and push ──────────────────────────────────────────────────
echo ""
echo "=== STEP 7: Commit and push ==="
cd $INKVAULT
git add -A
if git diff --cached --quiet; then
  echo "  (nothing to commit — already clean)"
else
  git commit -m "feat: complete InkVault v1.0.0 - all tests pass, frontend builds clean"
  echo "  ✅ Committed"
fi

GH_TOKEN=$(gh auth token 2>/dev/null || echo "")
if [ -n "$GH_TOKEN" ]; then
  git remote set-url origin "https://AKIB473:${GH_TOKEN}@github.com/AKIB473/inkvault.git"
  git push -u origin main 2>&1
  git remote set-url origin "https://github.com/AKIB473/inkvault.git"
  echo "  ✅ Pushed to GitHub"
else
  echo "  ⚠️  No GH_TOKEN available, skipping push"
fi

# ── STEP 8: Final verify ─────────────────────────────────────────────────────
echo ""
echo "=== STEP 8: Final verify ==="
gh api repos/AKIB473/inkvault --jq '{name:.name, updated:.updated_at, url:.html_url}' 2>&1 || true
echo ""
git log --oneline -5
echo ""
echo "════════════════════════════════════════"
echo "  ✅ ALL DONE!"
echo "════════════════════════════════════════"
