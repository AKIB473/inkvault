package domain

import "time"

// TokenType distinguishes different single-use token purposes.
type TokenType string

const (
	TokenTypePasswordReset  TokenType = "password_reset"
	TokenTypeEmailVerify    TokenType = "email_verify"
	TokenTypeInvite         TokenType = "invite"
	TokenTypeTwoFA          TokenType = "two_fa"
	TokenTypeDeleteConfirm  TokenType = "delete_confirm"
)

// Token represents a single-use secure token (learned from Ghost).
// Only the SHA-256 hash is stored — raw token is only returned once at creation.
type Token struct {
	ID        string    `json:"id"`
	Hash      string    `json:"-"`          // SHA-256 of raw token, stored in DB
	UserID    string    `json:"user_id"`
	Type      TokenType `json:"type"`
	UsedAt    *time.Time `json:"used_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	// For invite tokens
	InvitedEmail string `json:"invited_email,omitempty"`
	CreatedBy    string `json:"created_by,omitempty"`
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t *Token) IsUsed() bool {
	return t.UsedAt != nil
}

func (t *Token) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed()
}

// AuditLog records every significant write action (learned from Ghost's audit trail).
type AuditLog struct {
	ID           string    `json:"id"`
	ActorID      string    `json:"actor_id"`
	ActorRole    Role      `json:"actor_role"`
	Action       string    `json:"action"`       // "post.create", "user.delete", etc.
	ResourceType string    `json:"resource_type"` // "post", "user", "blog"
	ResourceID   string    `json:"resource_id"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	Meta         map[string]interface{} `json:"meta,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Session stores device info alongside refresh tokens (learned from Ghost device verification).
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	RefreshHash  string    `json:"-"`   // SHA-256 of refresh token
	DeviceInfo   string    `json:"device_info"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	LastSeenAt   time.Time `json:"last_seen_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}
