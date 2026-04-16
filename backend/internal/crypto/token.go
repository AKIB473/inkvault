// token.go — Single-use HMAC-signed tokens (inspired by Ghost + WriteFreely)
package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateSecureToken generates a cryptographically random hex token of given byte length.
func GenerateSecureToken(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("token generation: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// SignToken creates an HMAC-SHA256 signature over token+expiry, returning a signed token string.
// Format: "<token>.<expiry_unix>.<signature>"
func SignToken(token string, expiry time.Time, secret []byte) string {
	payload := fmt.Sprintf("%s.%d", token, expiry.Unix())
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s.%d.%s", token, expiry.Unix(), sig)
}

// HashToken returns SHA-256 hash of a token — used to store in DB without exposing raw token.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// GenerateInviteToken creates a single-use invite token (32 bytes).
func GenerateInviteToken() (string, error) {
	return GenerateSecureToken(32)
}

// GeneratePasswordResetToken creates a single-use password reset token (32 bytes).
func GeneratePasswordResetToken() (string, error) {
	return GenerateSecureToken(32)
}

// GenerateRefreshToken creates a refresh token (64 bytes for high entropy).
func GenerateRefreshToken() (string, error) {
	return GenerateSecureToken(64)
}
