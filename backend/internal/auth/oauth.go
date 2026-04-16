// Package auth — OAuth provider support.
// Learned from WriteFreely's OAuthButtons pattern: each provider is opt-in via config.
// We support GitHub and Google by default; more can be added via Generic OAuth.
package auth

import (
	"github.com/you/inkvault/internal/config"
)

// OAuthProviders holds which OAuth providers are enabled (driven by config).
// Passed to the frontend so it knows which buttons to render.
type OAuthProviders struct {
	GitHubEnabled  bool   `json:"github_enabled"`
	GoogleEnabled  bool   `json:"google_enabled"`
	GenericEnabled bool   `json:"generic_enabled"`
	GenericName    string `json:"generic_name,omitempty"`
}

// NewOAuthProviders reads config and returns enabled providers.
func NewOAuthProviders(cfg *config.Config) *OAuthProviders {
	return &OAuthProviders{
		GitHubEnabled:  cfg.GitHubOAuthClientID != "",
		GoogleEnabled:  cfg.GoogleOAuthClientID != "",
		GenericEnabled: cfg.GenericOAuthClientID != "",
		GenericName:    cfg.GenericOAuthName,
	}
}
