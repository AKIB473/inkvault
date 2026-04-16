// Package health provides a production-grade health check endpoint.
// Checks: DB connectivity, Redis ping, and reports version/uptime.
package health

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var startTime = time.Now()

type Handler struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewHandler(db *pgxpool.Pool, rdb *redis.Client) *Handler {
	return &Handler{db: db, rdb: rdb}
}

type HealthStatus struct {
	Status   string            `json:"status"` // "ok" | "degraded" | "down"
	Version  string            `json:"version"`
	Uptime   string            `json:"uptime"`
	Checks   map[string]string `json:"checks"`
}

// Check godoc — GET /api/v1/health
func (h *Handler) Check(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	checks := map[string]string{}
	overall := "ok"

	// DB check
	if err := h.db.Ping(ctx); err != nil {
		checks["postgres"] = "down: " + err.Error()
		overall = "down"
	} else {
		checks["postgres"] = "ok"
	}

	// Redis check
	if err := h.rdb.Ping(ctx).Err(); err != nil {
		checks["redis"] = "degraded: " + err.Error()
		if overall == "ok" {
			overall = "degraded"
		}
	} else {
		checks["redis"] = "ok"
	}

	status := &HealthStatus{
		Status:  overall,
		Version: version,
		Uptime:  time.Since(startTime).Round(time.Second).String(),
		Checks:  checks,
	}

	code := fiber.StatusOK
	if overall == "down" {
		code = fiber.StatusServiceUnavailable
	}

	return c.Status(code).JSON(status)
}

// version is injected at build time via ldflags.
var version = "dev"
