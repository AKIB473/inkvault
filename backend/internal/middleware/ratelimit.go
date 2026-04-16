// Package middleware provides HTTP middleware for Fiber.
package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimiter creates a Redis-backed rate limiter middleware.
// Ghost uses 5 attempts/hr/IP for auth endpoints — we do the same.
func RateLimiter(rdb *redis.Client, limit int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := fmt.Sprintf("ratelimit:%s:%s", c.Path(), ip)

		ctx := c.Context()

		// Increment and set expiry atomically
		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)

		if _, err := pipe.Exec(ctx); err != nil {
			// Fail open — don't block users if Redis is down
			return c.Next()
		}

		count := incr.Val()
		if count > int64(limit) {
			retryAfter := window.Seconds()
			c.Set("Retry-After", fmt.Sprintf("%.0f", retryAfter))
			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Set("X-RateLimit-Remaining", "0")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "too_many_requests",
				"message": fmt.Sprintf("Rate limit exceeded. Try again in %.0f seconds.", retryAfter),
			})
		}

		remaining := int64(limit) - count
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		return c.Next()
	}
}
