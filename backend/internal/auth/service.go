// Package auth handles authentication and authorization logic.
package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/you/inkvault/internal/config"
	"github.com/you/inkvault/internal/crypto"
	"github.com/you/inkvault/internal/domain"
	"github.com/you/inkvault/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserBanned         = errors.New("account is banned")
	ErrUserSilenced       = errors.New("account is silenced")
	ErrTokenInvalid       = errors.New("token is invalid or expired")
	ErrHoneypotTriggered  = errors.New("registration rejected")
	ErrUsernameExists     = errors.New("username already taken")
	ErrRegistrationClosed = errors.New("registration is currently closed")
	Err2FARequired        = errors.New("two-factor authentication required")
)

const bcryptCost = 12 // Higher than Ghost's default 10

// Service handles all auth operations.
type Service struct {
	cfg       *config.Config
	store     repository.Store
	encryptor *crypto.EmailEncryptor
}

func NewService(cfg *config.Config, store repository.Store, enc *crypto.EmailEncryptor) *Service {
	return &Service{cfg: cfg, store: store, encryptor: enc}
}

// RegisterInput is the registration payload.
// Honeypot field (Fullname) is invisible to humans — bots fill it in.
type RegisterInput struct {
	Username   string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Password   string `json:"password" validate:"required,min=8,max=128"`
	Email      string `json:"email" validate:"required,email"`
	InviteCode string `json:"invite_code"`
	Fullname   string `json:"fullname"` // HONEYPOT — must be empty
}

// Register creates a new user account.
func (s *Service) Register(ctx context.Context, input RegisterInput) (*domain.User, string, error) {
	// Honeypot check — bots fill the hidden "fullname" field
	if strings.TrimSpace(input.Fullname) != "" {
		return nil, "", ErrHoneypotTriggered
	}

	// Check registration policy
	if !s.cfg.OpenRegistration && !s.cfg.InviteOnly {
		return nil, "", ErrRegistrationClosed
	}

	// If invite-only, validate invite token
	if s.cfg.InviteOnly {
		if err := s.validateAndConsumeInvite(ctx, input.InviteCode); err != nil {
			return nil, "", err
		}
	}

	// Hash password (bcrypt cost 12)
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcryptCost)
	if err != nil {
		return nil, "", err
	}

	// Encrypt email at rest (AES-256-GCM)
	encEmail, err := s.encryptor.Encrypt(strings.ToLower(input.Email))
	if err != nil {
		return nil, "", err
	}

	user := &domain.User{
		ID:             uuid.New().String(),
		Username:       strings.ToLower(input.Username),
		DisplayName:    input.Username,
		EmailEncrypted: encEmail,
		PasswordHash:   string(hash),
		Role:           domain.RoleWriter, // Default role
		Status:         domain.UserStatusActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.store.Users().CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	// Generate access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, accessToken, nil
}

// LoginInput is the login payload.
type LoginInput struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResult holds the auth tokens and user.
type LoginResult struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
	Requires2FA  bool
}

// Login authenticates a user and returns tokens.
func (s *Service) Login(ctx context.Context, input LoginInput, ip, userAgent string) (*LoginResult, error) {
	user, err := s.store.Users().GetUserByUsername(ctx, strings.ToLower(input.Username))
	if err != nil {
		// Constant-time comparison to prevent timing attacks
		bcrypt.CompareHashAndPassword([]byte("$2a$12$dummy"), []byte(input.Password))
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.IsBanned() {
		return nil, ErrUserBanned
	}

	if user.IsSilenced() {
		return nil, ErrUserSilenced
	}

	// Check 2FA
	if user.TwoFAEnabled {
		// Return partial result — frontend must submit 2FA code
		return &LoginResult{User: user, Requires2FA: true}, nil
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
		Requires2FA:  false,
	}, nil
}

// RefreshTokens exchanges a valid refresh token for new access + refresh tokens.
func (s *Service) RefreshTokens(ctx context.Context, refreshToken, ip, userAgent string) (*LoginResult, error) {
	hash := crypto.HashToken(refreshToken)

	session, err := s.store.Sessions().GetSessionByRefreshHash(ctx, hash)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.store.Sessions().DeleteSession(ctx, session.ID)
		return nil, ErrTokenInvalid
	}

	user, err := s.store.Users().GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Rotate refresh token (old one deleted)
	_ = s.store.Sessions().DeleteSession(ctx, session.ID)
	newRefreshToken, err := s.createSession(ctx, user, ip, userAgent)
	if err != nil {
		return nil, err
	}

	newAccessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:         user,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout invalidates the session.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	hash := crypto.HashToken(refreshToken)
	session, err := s.store.Sessions().GetSessionByRefreshHash(ctx, hash)
	if err != nil {
		return nil // Already gone — idempotent
	}
	return s.store.Sessions().DeleteSession(ctx, session.ID)
}

// LogoutAll revokes all sessions for a user (learned from Ghost's "sign out all devices").
func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	return s.store.Sessions().DeleteAllUserSessions(ctx, userID)
}

// generateAccessToken creates a short-lived JWT (15min default).
func (s *Service) generateAccessToken(user *domain.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"role":     string(user.Role),
		"iat":      now.Unix(),
		"exp":      now.Add(s.cfg.JWTAccessTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// createSession creates a new refresh token session and returns the raw token.
func (s *Service) createSession(ctx context.Context, user *domain.User, ip, userAgent string) (string, error) {
	raw, err := crypto.GenerateRefreshToken()
	if err != nil {
		return "", err
	}

	session := &domain.Session{
		ID:          uuid.New().String(),
		UserID:      user.ID,
		RefreshHash: crypto.HashToken(raw),
		IPAddress:   ip,
		UserAgent:   userAgent,
		DeviceInfo:  parseDevice(userAgent),
		LastSeenAt:  time.Now(),
		ExpiresAt:   time.Now().Add(s.cfg.JWTRefreshTTL),
		CreatedAt:   time.Now(),
	}

	if err := s.store.Sessions().CreateSession(ctx, session); err != nil {
		return "", err
	}

	return raw, nil
}

// validateAndConsumeInvite validates an invite token and marks it used.
func (s *Service) validateAndConsumeInvite(ctx context.Context, code string) error {
	if code == "" {
		return errors.New("invite code required")
	}
	hash := crypto.HashToken(code)
	token, err := s.store.Tokens().GetTokenByHash(ctx, hash, domain.TokenTypeInvite)
	if err != nil || !token.IsValid() {
		return errors.New("invalid or expired invite code")
	}
	return s.store.Tokens().MarkTokenUsed(ctx, token.ID)
}

// parseDevice extracts a human-readable device string from User-Agent.
func parseDevice(ua string) string {
	switch {
	case strings.Contains(ua, "Mobile"):
		return "Mobile"
	case strings.Contains(ua, "Tablet"):
		return "Tablet"
	case strings.Contains(ua, "curl"):
		return "CLI"
	default:
		return "Desktop"
	}
}
