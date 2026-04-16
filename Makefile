.PHONY: infra stop backend frontend migrate seed test test-race lint build docker-build clean

# ── Infrastructure ────────────────────────────────────────────────────────────
infra:
	docker compose up -d postgres redis minio
	@echo "✅ Infrastructure running"
	@echo "   Postgres: localhost:5432"
	@echo "   Redis:    localhost:6379"
	@echo "   MinIO:    localhost:9000 (console: localhost:9001)"

stop:
	docker compose down

# ── Dev servers ───────────────────────────────────────────────────────────────
backend:
	cd backend && go run cmd/server/main.go

frontend:
	cd frontend && npm run dev

# ── Database ──────────────────────────────────────────────────────────────────
migrate:
	cd backend && go run cmd/migrate/main.go

seed:
	cd backend && go run cmd/seed/main.go

# ── Tests ─────────────────────────────────────────────────────────────────────
test:
	@echo "=== Backend Tests ==="
	cd backend && go test ./... -v
	@echo ""
	@echo "=== Frontend Type Check ==="
	cd frontend && npm run type-check

# Run tests twice with race detector (catches concurrency bugs)
test-race:
	@echo "=== Backend Tests (race detector, 2 runs) ==="
	cd backend && go test -race -count=2 -coverprofile=coverage.out ./...
	@echo ""
	cd backend && go tool cover -func=coverage.out | grep total

# Single package
test-pkg:
	cd backend && go test -v -run $(RUN) ./internal/$(PKG)/...

# ── Linting ───────────────────────────────────────────────────────────────────
lint:
	cd backend && golangci-lint run --timeout=5m
	cd frontend && npm run lint

vet:
	cd backend && go vet ./...

# ── Build ─────────────────────────────────────────────────────────────────────
build:
	cd backend && CGO_ENABLED=0 go build -ldflags="-w -s" -trimpath \
		-o bin/inkvault-server ./cmd/server
	cd backend && CGO_ENABLED=0 go build -ldflags="-w -s" -trimpath \
		-o bin/inkvault-migrate ./cmd/migrate
	cd frontend && npm run build
	@echo "✅ Build complete"

docker-build:
	docker build -t inkvault/backend:dev ./backend
	docker build -t inkvault/frontend:dev ./frontend
	@echo "✅ Docker images built"

# ── sqlc code generation ──────────────────────────────────────────────────────
sqlc:
	cd backend && sqlc generate

# ── Cleanup ───────────────────────────────────────────────────────────────────
clean:
	cd backend && rm -rf bin/ coverage.out
	cd frontend && rm -rf .next/ out/

# ── Production deploy ─────────────────────────────────────────────────────────
deploy:
	docker compose -f docker-compose.prod.yml up -d --build
	@echo "✅ Production deploy started"

deploy-down:
	docker compose -f docker-compose.prod.yml down

# ── Help ──────────────────────────────────────────────────────────────────────
help:
	@echo "InkVault Development Commands"
	@echo ""
	@echo "  make infra       Start Postgres + Redis + MinIO"
	@echo "  make migrate     Run DB migrations"
	@echo "  make seed        Seed dev data"
	@echo "  make backend     Start Go API server"
	@echo "  make frontend    Start Next.js dev server"
	@echo "  make test        Run all tests"
	@echo "  make test-race   Run tests with race detector (2x)"
	@echo "  make lint        Lint all code"
	@echo "  make build       Build production binaries"
	@echo "  make docker-build Build Docker images"
	@echo "  make deploy      Deploy production stack"
	@echo "  make clean       Clean build artifacts"
