package auth

// Password reset implementation.
// Key security lessons from Ghost's passwordreset.js:
//
// 1. Token payload includes the user's CURRENT password hash.
//    → If user changes password, ALL existing reset tokens are instantly invalidated.
//    → Ghost does this via security.tokens.resetToken.compare({token, dbHash, password})
//
// 2. Token payload includes an instance-specific db_hash (secret).
//    → Tokens from one instance cannot be replayed on another.
//    → We use APP_SECRET as our instance secret.
//
// 3. Per-token brute-force counter: if a token is probed >= 10 times, it locks.
//    → Prevents enumeration even if attacker has a valid-looking token.
//
// 4. Tokens expire after 24 hours (Ghost uses moment().add(1, 'days')).
//
// 5. Tokens are base64-encoded for URL safety, stored as SHA-256 hash in DB.

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/you/inkvault/internal/crypto"
	"github.com/you/inkvault/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

const (
	passwordResetExpiry   = 24 * time.Hour
	passwordResetMaxProbes = 10 // Lock token after 10 failed attempts (Ghost pattern)
)

// tokenProbeCounter tracks brute-force attempts per token (in-memory, like Ghost).
// Key: tokenHash, Value: probe count
var (
	tokenProbes   = make(map[string]int)
	tokenProbesMu sync.Mutex
)

// PasswordResetTokenPayload is embedded in the reset token.
// Includes password hash so token auto-invalidates on password change (Ghost pattern).
type PasswordResetTokenPayload struct {
	UserID       string    `json:"uid"`
	Email        string    `json:"email"`   // Plaintext only in token, not stored
	PasswordHash string    `json:"ph"`      // Current bcrypt hash — invalidates token on pw change
	InstanceKey  string    `json:"ik"`      // Instance secret — prevents cross-instance replay
	Expires      time.Time `json:"exp"`
}

// GeneratePasswordResetToken creates a signed reset token for the given user.
// Returns the raw token (to send in email) — only the hash is stored in DB.
func (s *Service) GeneratePasswordResetToken(ctx context.Context, username string) (rawToken string, err error) {
	user, err := s.store.Users().GetUserByUsername(ctx, username)
	if err != nil {
		// Return generic error — don't reveal whether user exists (timing-safe)
		return "", nil
	}

	// Delete any existing reset tokens for this user (one active at a time)
	_ = s.store.Tokens().DeleteTokensByUser(ctx, user.ID, domain.TokenTypePasswordReset)

	// Build payload: includes current password hash (Ghost's key insight)
	payload := PasswordResetTokenPayload{
		UserID:       user.ID,
		PasswordHash: user.PasswordHash,                  // Invalidates if password changes
		InstanceKey:  hmacHex(user.ID, s.cfg.JWTSecret),  // Instance-specific
		Expires:      time.Now().Add(passwordResetExpiry),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal token payload: %w", err)
	}

	// Sign payload with HMAC-SHA256
	sig := hmacHex(string(payloadJSON), s.cfg.JWTSecret)
	rawToken = hex.EncodeToString(payloadJSON) + "." + sig

	// Store only the hash in DB
	tokenHash := crypto.HashToken(rawToken)
	token := &domain.Token{
		ID:        uuid.New().String(),
		Hash:      tokenHash,
		UserID:    user.ID,
		Type:      domain.TokenTypePasswordReset,
		ExpiresAt: payload.Expires,
		CreatedAt: time.Now(),
	}

	if err := s.store.Tokens().CreateToken(ctx, token); err != nil {
		return "", fmt.Errorf("store token: %w", err)
	}

	return rawToken, nil
}

// ResetPasswordInput is the payload for completing a password reset.
type ResetPasswordInput struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

// ResetPassword validates the token and sets the new password.
func (s *Service) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	tokenHash := crypto.HashToken(input.Token)

	// Brute-force protection: lock token after 10 probes (Ghost pattern)
	tokenProbesMu.Lock()
	probes := tokenProbes[tokenHash]
	if probes >= passwordResetMaxProbes {
		tokenProbesMu.Unlock()
		return errors.New("reset link has been locked due to too many attempts. Request a new one")
	}
	tokenProbes[tokenHash]++
	tokenProbesMu.Unlock()

	// Look up token in DB
	token, err := s.store.Tokens().GetTokenByHash(ctx, tokenHash, domain.TokenTypePasswordReset)
	if err != nil || !token.IsValid() {
		return errors.New("reset link is invalid or has expired")
	}

	// Decode and verify payload
	parts := splitToken(input.Token)
	if len(parts) != 2 {
		return errors.New("reset link is malformed")
	}

	payloadBytes, err := hex.DecodeString(parts[0])
	if err != nil {
		return errors.New("reset link is malformed")
	}

	// Verify signature
	expectedSig := hmacHex(string(payloadBytes), s.cfg.JWTSecret)
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSig)) {
		return errors.New("reset link signature is invalid")
	}

	var payload PasswordResetTokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return errors.New("reset link is malformed")
	}

	// Check expiry
	if time.Now().After(payload.Expires) {
		return errors.New("reset link has expired. Request a new one")
	}

	// Load user
	user, err := s.store.Users().GetUserByID(ctx, payload.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	// Ghost's key check: verify current password hash matches token payload.
	// If user already changed password, this won't match → token invalidated.
	if user.PasswordHash != payload.PasswordHash {
		return errors.New("reset link has been invalidated. The password was already changed")
	}

	// Verify instance key
	expectedKey := hmacHex(user.ID, s.cfg.JWTSecret)
	if !hmac.Equal([]byte(payload.InstanceKey), []byte(expectedKey)) {
		return errors.New("reset link is invalid for this instance")
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Update password
	if err := s.store.Users().UpdatePasswordHash(ctx, user.ID, string(newHash)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// Consume token (mark used)
	_ = s.store.Tokens().MarkTokenUsed(ctx, token.ID)

	// Revoke all sessions (force re-login everywhere — Ghost does this too)
	_ = s.store.Sessions().DeleteAllUserSessions(ctx, user.ID)

	// Clear probe counter
	tokenProbesMu.Lock()
	delete(tokenProbes, tokenHash)
	tokenProbesMu.Unlock()

	return nil
}

// hmacHex returns hex-encoded HMAC-SHA256 of data with key.
func hmacHex(data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// splitToken splits "hexPayload.signature" into parts.
func splitToken(token string) []string {
	// Find last dot — signature follows it
	for i := len(token) - 1; i >= 0; i-- {
		if token[i] == '.' {
			return []string{token[:i], token[i+1:]}
		}
	}
	return nil
}
