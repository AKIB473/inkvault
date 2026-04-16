package middleware

// Structured request logger using zerolog.
// Production-grade: logs method, path, status, latency, request ID.
// Never logs Authorization headers or request bodies (privacy).

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// RequestLogger attaches a request ID and logs each request.
func RequestLogger(logger zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Generate request ID for tracing
		reqID := uuid.New().String()[:8]
		c.Set("X-Request-ID", reqID)
		c.Locals("requestID", reqID)

		err := c.Next()

		// Log after response — never log sensitive headers
		latency := time.Since(start)
		status := c.Response().StatusCode()

		event := logger.Info()
		if status >= 500 {
			event = logger.Error()
		} else if status >= 400 {
			event = logger.Warn()
		}

		event.
			Str("req_id", reqID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Dur("latency", latency).
			Str("ip", c.IP()).
			// Never log: Authorization, Cookie, X-API-Key
			Msg("request")

		return err
	}
}

// InitLogger configures zerolog for production (JSON) or development (pretty).
func InitLogger(env string) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if env == "development" {
		return log.Output(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	}
	// Production: JSON logs for log aggregators (Loki, Datadog, etc.)
	return zerolog.New(nil).With().Timestamp().Logger()
}
