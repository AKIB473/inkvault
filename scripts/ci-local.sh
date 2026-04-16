#!/bin/bash
# ci-local.sh — Run full CI pipeline locally (same as GitHub Actions)
# Usage: bash scripts/ci-local.sh

set -euo pipefail
cd "$(dirname "$0")/.."

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
pass() { echo -e "${GREEN}✅ $1${NC}"; }
fail() { echo -e "${RED}❌ $1${NC}"; exit 1; }
info() { echo -e "${YELLOW}⟳  $1${NC}"; }

echo "======================================"
echo " InkVault CI — Local"
echo "======================================"

# ── Backend ───────────────────────────────
info "Backend: go mod tidy"
cd backend
go mod tidy 2>&1 || fail "go mod tidy failed"
pass "go mod tidy"

info "Backend: go mod verify"
go mod verify 2>&1 || fail "go mod verify failed"
pass "go mod verify"

info "Backend: go vet"
go vet ./... 2>&1 || fail "go vet failed"
pass "go vet"

info "Backend: tests (run 1)"
go test -count=1 ./... 2>&1 || fail "tests failed (run 1)"
pass "tests run 1"

info "Backend: tests (run 2 — catches flaky tests)"
go test -count=1 ./... 2>&1 || fail "tests failed (run 2)"
pass "tests run 2"

info "Backend: tests with race detector"
go test -race -count=1 ./... 2>&1 || fail "race tests failed"
pass "race detector tests"

info "Backend: coverage"
go test -coverprofile=coverage.out ./... 2>&1
go tool cover -func=coverage.out | grep total
pass "coverage report"

info "Backend: build server"
CGO_ENABLED=0 go build -ldflags="-w -s" -trimpath -o /tmp/inkvault-server ./cmd/server 2>&1 \
  || fail "server build failed"
pass "server binary builds"

info "Backend: build migrate"
CGO_ENABLED=0 go build -ldflags="-w -s" -trimpath -o /tmp/inkvault-migrate ./cmd/migrate 2>&1 \
  || fail "migrate build failed"
pass "migrate binary builds"

cd ..

# ── Frontend ──────────────────────────────
info "Frontend: npm ci"
cd frontend
npm ci --frozen-lockfile 2>&1 | tail -5 || fail "npm ci failed"
pass "npm ci"

info "Frontend: type check"
npm run type-check 2>&1 || fail "type check failed"
pass "type check"

info "Frontend: lint"
npm run lint 2>&1 || fail "lint failed"
pass "lint"

info "Frontend: build"
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1 \
NEXT_PUBLIC_APP_URL=http://localhost:3000 \
npm run build 2>&1 | tail -10 || fail "frontend build failed"
pass "frontend build"

cd ..

echo ""
echo "======================================"
echo -e "${GREEN} All checks passed! ✅${NC}"
echo "======================================"
