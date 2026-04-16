# Contributing to InkVault

Thanks for your interest! Here's how to get started.

## Development Setup

```bash
git clone https://github.com/your-username/inkvault
cd inkvault
make infra       # Start Postgres + Redis + MinIO
make migrate     # Apply migrations
make seed        # Seed dev data
make backend     # API on :8080
make frontend    # UI on :3000
```

## Running Tests

```bash
# Backend (with race detector — run twice to catch flaky tests)
cd backend && go test -race -count=2 ./...

# Frontend
cd frontend && npm run type-check && npm run lint
```

## Pull Request Guidelines

1. **Tests required** — new features need tests; bug fixes need a regression test
2. **No secrets** — never commit `.env` files or keys
3. **Privacy first** — new data collection requires justification
4. **Security matters** — auth/crypto changes need extra review
5. **One concern per PR** — keep PRs focused

## Code Style

- Go: `gofmt` + `golangci-lint` (run `make lint`)
- TypeScript: ESLint + Prettier
- Commits: conventional commits (`feat:`, `fix:`, `chore:`, `docs:`)

## Reporting Bugs

Open an issue with:
- Steps to reproduce
- Expected vs actual behavior
- Environment (OS, Go/Node version)

Security issues → see [SECURITY.md](SECURITY.md).
