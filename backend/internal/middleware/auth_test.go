package middleware_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/you/inkvault/internal/middleware"
)

const testSecret = "super-secret-test-key-32-bytes!!"

// buildApp creates a Fiber app with JWTMiddleware and a test route.
func buildApp(secret string) *fiber.App {
	app := fiber.New(fiber.Config{
		// Return errors as JSON so we can decode them.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(middleware.JWTMiddleware(secret))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"userID":   c.Locals("userID"),
			"username": c.Locals("username"),
			"role":     c.Locals("role"),
		})
	})
	return app
}

// makeToken creates a signed HS256 JWT with the given claims and secret.
func makeToken(secret string, sub, username, role string, exp time.Time) string {
	claims := jwt.MapClaims{
		"sub":      sub,
		"username": username,
		"role":     role,
		"iat":      time.Now().Unix(),
		"exp":      exp.Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString([]byte(secret))
	return signed
}

// ── Tests ─────────────────────────────────────────────────────────────────────

// TestJWTMiddleware_MissingHeader — no Authorization header → 401.
func TestJWTMiddleware_MissingHeader(t *testing.T) {
	app := buildApp(testSecret)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", resp.StatusCode)
	}
}

// TestJWTMiddleware_InvalidFormat — malformed header (no "Bearer" prefix) → 401.
func TestJWTMiddleware_InvalidFormat(t *testing.T) {
	app := buildApp(testSecret)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token abc123")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", resp.StatusCode)
	}
}

// TestJWTMiddleware_InvalidToken — garbage token string → 401.
func TestJWTMiddleware_InvalidToken(t *testing.T) {
	app := buildApp(testSecret)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer this.is.garbage")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", resp.StatusCode)
	}
}

// TestJWTMiddleware_ExpiredToken — valid signature but expired → 401.
func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	app := buildApp(testSecret)
	tok := makeToken(testSecret, "user-1", "alice", "writer", time.Now().Add(-1*time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", resp.StatusCode)
	}
}

// TestJWTMiddleware_WrongSecret — token signed with a different secret → 401.
func TestJWTMiddleware_WrongSecret(t *testing.T) {
	app := buildApp(testSecret)
	tok := makeToken("wrong-secret-entirely", "user-1", "alice", "writer", time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", resp.StatusCode)
	}
}

// TestJWTMiddleware_ValidToken — valid token passes through and claims are extracted.
func TestJWTMiddleware_ValidToken(t *testing.T) {
	app := buildApp(testSecret)
	tok := makeToken(testSecret, "user-42", "bob", "admin", time.Now().Add(15*time.Minute))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if result["userID"] != "user-42" {
		t.Errorf("want userID=user-42, got %v", result["userID"])
	}
	if result["username"] != "bob" {
		t.Errorf("want username=bob, got %v", result["username"])
	}
	if result["role"] != "admin" {
		t.Errorf("want role=admin, got %v", result["role"])
	}
}

// TestJWTMiddleware_BearerCaseInsensitive — "BEARER" prefix should also work.
func TestJWTMiddleware_BearerCaseInsensitive(t *testing.T) {
	app := buildApp(testSecret)
	tok := makeToken(testSecret, "user-7", "carol", "reader", time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "BEARER "+tok)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

// ── RequireRole tests ─────────────────────────────────────────────────────────

func buildRoleApp(secret, minRole string) *fiber.App {
	app := fiber.New()
	app.Use(middleware.JWTMiddleware(secret))
	app.Get("/admin", middleware.RequireRole(minRole), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

// TestRequireRole_Sufficient — user role meets minimum → 200.
func TestRequireRole_Sufficient(t *testing.T) {
	app := buildRoleApp(testSecret, "writer")
	tok := makeToken(testSecret, "u1", "dave", "admin", time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

// TestRequireRole_Insufficient — user role below minimum → 403.
func TestRequireRole_Insufficient(t *testing.T) {
	app := buildRoleApp(testSecret, "admin")
	tok := makeToken(testSecret, "u2", "eve", "reader", time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("want 403, got %d", resp.StatusCode)
	}
}

// ── OptionalAuth tests ────────────────────────────────────────────────────────

func buildOptionalApp(secret string) *fiber.App {
	app := fiber.New()
	app.Use(middleware.OptionalAuth(secret))
	app.Get("/public", func(c *fiber.Ctx) error {
		uid, _ := c.Locals("userID").(string)
		return c.JSON(fiber.Map{"userID": uid, "authed": uid != ""})
	})
	return app
}

// TestOptionalAuth_NoHeader — no header, request still proceeds unauthenticated.
func TestOptionalAuth_NoHeader(t *testing.T) {
	app := buildOptionalApp(testSecret)
	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if result["authed"] != false {
		t.Errorf("want authed=false, got %v", result["authed"])
	}
}

// TestOptionalAuth_ValidToken — valid token sets user context.
func TestOptionalAuth_ValidToken(t *testing.T) {
	app := buildOptionalApp(testSecret)
	tok := makeToken(testSecret, "user-99", "frank", "writer", time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if result["userID"] != "user-99" {
		t.Errorf("want userID=user-99, got %v", result["userID"])
	}
}

// TestOptionalAuth_InvalidToken — invalid token still proceeds unauthenticated.
func TestOptionalAuth_InvalidToken(t *testing.T) {
	app := buildOptionalApp(testSecret)
	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if result["authed"] != false {
		t.Errorf("want authed=false for invalid token, got %v", result["authed"])
	}
}
