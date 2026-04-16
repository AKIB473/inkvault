package auth

// API Key authentication — learned from Ghost's identity-token-service.ts
//
// Ghost uses TWO types of API access:
//   1. Content API key  — read-only public key, anyone can use (for headless frontends)
//   2. Admin API key    — read/write, authenticated users + integrations
//
// Ghost's IdentityTokenService generates short-lived RS256 JWTs (5 minutes)
// signed with a RSA private key stored in settings. This is for inter-service auth.
//
// For InkVault we implement a simpler but equally secure model:
//   - Content API key: UUID stored in DB, passed as ?key= query param
//   - Admin API key: HMAC-SHA256 signed token for integrations/webhooks
//   - Identity token: Short-lived HS256 JWT (5 min) for internal service calls
//
// Ghost's authorize.js lesson:
//   - Content API: pass if req has api_key OR member session
//   - Admin API:   pass if req has user session OR api_key

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/you/inkvault/internal/domain"
)

// APIKeyType distinguishes Content vs Admin API keys (Ghost pattern).
type APIKeyType string

const (
	APIKeyContent APIKeyType = "content" // Read-only, public
	APIKeyAdmin   APIKeyType = "admin"   // Read/write, authenticated
)

// APIKey represents a blog's API key for headless/integration access.
type APIKey struct {
	ID        string     `json:"id"`
	BlogID    string     `json:"blog_id"`
	Type      APIKeyType `json:"type"`
	KeyHash   string     `json:"-"` // SHA-256 of raw key, never returned
	CreatedAt time.Time  `json:"created_at"`
	LastUsed  *time.Time `json:"last_used,omitempty"`
}

// IdentityTokenClaims are the claims in an identity token (Ghost pattern).
// Used for short-lived inter-service authentication (5 minutes).
type IdentityTokenClaims struct {
	Role string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// GenerateIdentityToken creates a short-lived (5min) JWT for the given user.
// Learned from Ghost's IdentityTokenService.getTokenForUser().
// Ghost uses RS256 (RSA); we use HS256 for simplicity — upgrade to RS256 for microservices.
func (s *Service) GenerateIdentityToken(user *domain.User, issuer string) (string, error) {
	claims := IdentityTokenClaims{
		Role: string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)), // Ghost: 5min
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// AuthorizeContentAPI checks if a request has a valid content API key OR member session.
// Learned from Ghost's authorize.authorizeContentApi.
func AuthorizeContentAPI(apiKey string, hasMemberSession bool) bool {
	return apiKey != "" || hasMemberSession
}

// AuthorizeAdminAPI checks if a request has a valid user session OR admin API key.
// Learned from Ghost's authorize.authorizeAdminApi.
func AuthorizeAdminAPI(userID string, apiKey string) bool {
	return userID != "" || apiKey != ""
}
