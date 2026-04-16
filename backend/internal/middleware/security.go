package middleware

import "github.com/gofiber/fiber/v2"

// SecurityHeaders adds hardened HTTP security headers to all responses.
// Learned from Ghost + OWASP recommendations.
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Prevent MIME sniffing
		c.Set("X-Content-Type-Options", "nosniff")
		// Prevent clickjacking
		c.Set("X-Frame-Options", "DENY")
		// XSS protection (legacy browsers)
		c.Set("X-XSS-Protection", "1; mode=block")
		// Referrer policy — don't leak URLs
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// HSTS — force HTTPS (1 year, include subdomains)
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// Permissions policy — disable unnecessary browser features
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		// Content Security Policy
		c.Set("Content-Security-Policy",
			"default-src 'self'; "+
			"script-src 'self' 'unsafe-inline'; "+ // Tiptap needs inline scripts
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data: blob: https:; "+
			"font-src 'self'; "+
			"connect-src 'self'; "+
			"frame-ancestors 'none';")

		return c.Next()
	}
}

// AuditLogger records significant operations to the audit log.
// Attached to write endpoints (POST, PUT, PATCH, DELETE).
func AuditLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Run handler first
		err := c.Next()

		// Log only on successful writes
		method := c.Method()
		if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
			if c.Response().StatusCode() < 400 {
				userID, _ := c.Locals("userID").(string)
				// Audit logging is async — don't fail the request for it
				go func() {
					_ = userID // TODO: write to audit_log table
					// log: actor=userID, method=method, path=c.Path(), ip=c.IP()
				}()
			}
		}

		return err
	}
}
