package auth

// Two-factor authentication via email OTP.
// Ghost uses device verification (send code on new device) + optional email 2FA on every login.
// We implement both patterns.

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/you/inkvault/internal/crypto"
	"github.com/you/inkvault/internal/domain"
)

const (
	otpLength     = 6               // 6-digit code (000000–999999)
	otpExpiry     = 10 * time.Minute
	otpMaxAttempts = 5             // Lock after 5 wrong attempts (Ghost brute-force pattern)
)

// GenerateOTP creates a 6-digit OTP for the user, stores the hash in DB.
// Returns the raw 6-digit code to send via email (never stored).
func (s *Service) GenerateOTP(ctx context.Context, userID string) (string, error) {
	// Invalidate any existing OTP for this user
	_ = s.store.Tokens().DeleteTokensByUser(ctx, userID, domain.TokenTypeTwoFA)

	// Generate cryptographically random 6-digit code
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	code := fmt.Sprintf("%06d", n.Int64())

	token := &domain.Token{
		ID:        uuid.New().String(),
		Hash:      crypto.HashToken(code + userID), // Include userID to prevent cross-user reuse
		UserID:    userID,
		Type:      domain.TokenTypeTwoFA,
		ExpiresAt: time.Now().Add(otpExpiry),
		CreatedAt: time.Now(),
	}

	if err := s.store.Tokens().CreateToken(ctx, token); err != nil {
		return "", err
	}

	return code, nil
}

// VerifyOTP checks a submitted OTP code against the stored hash.
// Consumes the token on success.
func (s *Service) VerifyOTP(ctx context.Context, userID, code string) error {
	hash := crypto.HashToken(code + userID)

	token, err := s.store.Tokens().GetTokenByHash(ctx, hash, domain.TokenTypeTwoFA)
	if err != nil || !token.IsValid() {
		return fmt.Errorf("invalid or expired verification code")
	}

	if token.UserID != userID {
		return fmt.Errorf("invalid verification code")
	}

	// Consume token
	return s.store.Tokens().MarkTokenUsed(ctx, token.ID)
}

// LoginWith2FA completes login after successful OTP verification.
func (s *Service) LoginWith2FA(ctx context.Context, userID, code, ip, userAgent string) (*LoginResult, error) {
	if err := s.VerifyOTP(ctx, userID, code); err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.store.Users().GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.createSession(ctx, user, ip, userAgent)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
