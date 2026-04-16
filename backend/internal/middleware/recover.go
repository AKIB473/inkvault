package middleware

// Custom panic recovery — logs stack trace, returns 500, never leaks internals.

import (
	"fmt"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// PanicRecovery catches panics, logs them with a stack trace, and returns 500.
// Never exposes internal error details to the client.
func PanicRecovery() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				log.Error().
					Str("panic", fmt.Sprintf("%v", r)).
					Bytes("stack", stack).
					Str("path", c.Path()).
					Str("method", c.Method()).
					Msg("panic recovered")

				err = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "server_error",
					"message": "An unexpected error occurred. Our team has been notified.",
				})
			}
		}()
		return c.Next()
	}
}
