// Package domain defines core business entities.
package domain

import (
	"time"
)

// Role represents a user's permission level (learned from Ghost's role hierarchy).
type Role string

const (
	RoleOwner      Role = "owner"      // Full platform control
	RoleAdmin      Role = "admin"      // Manage users, all content
	RoleEditor     Role = "editor"     // Edit any post
	RoleWriter     Role = "writer"     // Create/edit own posts
	RoleReader     Role = "reader"     // Read members-only content
)

// UserStatus bitmask (learned from WriteFreely).
type UserStatus int

const (
	UserStatusActive   UserStatus = 0
	UserStatusSilenced UserStatus = 1 << 0
	UserStatusBanned   UserStatus = 1 << 1
)

// User is the core user entity.
// NOTE: Email is NEVER stored in plaintext — only encrypted bytes go to DB.
// Use EmailDecrypted() after calling SetDecryptedEmail().
type User struct {
	ID             string     `json:"id"`
	Username       string     `json:"username"`
	DisplayName    string     `json:"display_name"`
	EmailEncrypted string     `json:"-"`         // AES-256-GCM encrypted, never in JSON
	PasswordHash   string     `json:"-"`         // bcrypt, never in JSON
	Role           Role       `json:"role"`
	Status         UserStatus `json:"status"`
	TwoFAEnabled   bool       `json:"two_fa_enabled"`
	AvatarURL      string     `json:"avatar_url,omitempty"`
	Bio            string     `json:"bio,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"-"` // Hard delete: we actually DELETE the row

	// Transient — set at request time, never persisted
	clearEmail string
}

// SetDecryptedEmail caches the decrypted email in memory (never serialized).
func (u *User) SetDecryptedEmail(email string) {
	u.clearEmail = email
}

// Email returns the decrypted email if available, empty string otherwise.
// Never serialize this — it's transient only.
func (u *User) Email() string {
	return u.clearEmail
}

func (u *User) IsActive() bool {
	return u.Status&UserStatusBanned == 0 && u.Status&UserStatusSilenced == 0
}

func (u *User) IsSilenced() bool {
	return u.Status&UserStatusSilenced != 0
}

func (u *User) IsBanned() bool {
	return u.Status&UserStatusBanned != 0
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleOwner || u.Role == RoleAdmin
}

func (u *User) CanEdit() bool {
	return u.Role == RoleOwner || u.Role == RoleAdmin || u.Role == RoleEditor || u.Role == RoleWriter
}

// PublicUser is the safe-to-expose user object (no sensitive fields).
type PublicUser struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Bio         string `json:"bio,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToPublic strips sensitive fields.
func (u *User) ToPublic() *PublicUser {
	return &PublicUser{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		Bio:         u.Bio,
		CreatedAt:   u.CreatedAt,
	}
}
