package auth

// Learned from WriteFreely's handleUsernameCheck + getValidUsername patterns.
// Provides live username availability checking and slug sanitization.

import (
	"context"
	"regexp"
	"strings"
	"unicode"
)

var (
	// Only allow alphanumeric + underscore, 3-30 chars
	validUsernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

	// Reserved usernames that cannot be registered (route conflicts + abuse prevention)
	reservedUsernames = map[string]bool{
		"admin": true, "api": true, "app": true, "auth": true,
		"blog": true, "dashboard": true, "docs": true, "editor": true,
		"explore": true, "feed": true, "health": true, "help": true,
		"home": true, "login": true, "logout": true, "me": true,
		"new": true, "post": true, "privacy": true, "register": true,
		"rss": true, "search": true, "settings": true, "signup": true,
		"static": true, "support": true, "terms": true, "write": true,
		"writing": true, "www": true, "404": true, "500": true,
	}
)

// UsernameCheckResult is the response for the username availability endpoint.
type UsernameCheckResult struct {
	Available bool   `json:"available"`
	Slug      string `json:"slug,omitempty"`
	Error     string `json:"error,omitempty"`
}

// CheckUsernameAvailability validates and checks if a username is available.
// Returns the normalized slug and availability status.
// Learned from WriteFreely's getValidUsername — normalize first, then check DB.
func (s *Service) CheckUsernameAvailability(ctx context.Context, username string) UsernameCheckResult {
	slug := normalizeUsername(username)

	if slug == "" {
		msg := "Invalid username"
		if username != "" {
			msg += " — must contain at least 2 letters or numbers"
		}
		return UsernameCheckResult{Available: false, Error: msg}
	}

	if reservedUsernames[slug] {
		return UsernameCheckResult{Available: false, Error: "That username is reserved"}
	}

	// Check DB
	_, err := s.store.Users().GetUserByUsername(ctx, slug)
	if err != nil {
		// Not found = available
		return UsernameCheckResult{Available: true, Slug: slug}
	}

	return UsernameCheckResult{Available: false, Error: "Username is already taken"}
}

// normalizeUsername lowercases and strips invalid characters.
// Returns empty string if the result is too short to be valid.
func normalizeUsername(input string) string {
	// Lowercase
	s := strings.ToLower(strings.TrimSpace(input))

	// Remove anything that isn't alphanumeric or underscore
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(r)
		}
	}
	result := b.String()

	// Must match pattern and be at least 3 chars
	if !validUsernameRe.MatchString(result) {
		return ""
	}

	return result
}
