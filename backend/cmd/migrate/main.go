// Migration runner — reads SQL files from migrations/ and applies them in order.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	// Create migrations tracking table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`)
	if err != nil {
		log.Fatalf("create migrations table: %v", err)
	}

	// Find migration files
	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		log.Fatalf("glob: %v", err)
	}
	sort.Strings(files)

	applied := 0
	for _, f := range files {
		version := filepath.Base(f)

		// Check if already applied
		var exists bool
		_ = pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, version,
		).Scan(&exists)

		if exists {
			fmt.Printf("  ✓ %s (already applied)\n", version)
			continue
		}

		// Read and execute
		sql, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("read %s: %v", f, err)
		}

		// Skip comment-only files
		content := strings.TrimSpace(string(sql))
		if content == "" || strings.HasPrefix(content, "--") && !strings.Contains(content, ";") {
			fmt.Printf("  - %s (skipped, empty)\n", version)
			continue
		}

		start := time.Now()
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			log.Fatalf("apply %s: %v", version, err)
		}

		// Record as applied
		_, _ = pool.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, version)

		fmt.Printf("  ✅ %s (%.0fms)\n", version, float64(time.Since(start).Milliseconds()))
		applied++
	}

	if applied == 0 {
		fmt.Println("Already up to date.")
	} else {
		fmt.Printf("\nApplied %d migration(s).\n", applied)
	}
}
