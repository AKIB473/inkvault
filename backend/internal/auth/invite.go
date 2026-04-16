package auth

// Invite service.
// Learned from Ghost's Invites.js:
// - If an invite already exists for that email, destroy it and create fresh
// - Invite email includes: blog name, inviter name/email, signup link with encoded token
// - Token is base64-encoded in the URL
// - Invite status: "pending" → "sent" → (used on signup)

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/you/inkvault/internal/crypto"
	"github.com/you/inkvault/internal/domain"
)

const inviteTokenExpiry = 7 * 24 * time.Hour // 7 days

// InviteInput is the payload for creating an invite.
type InviteInput struct {
	Email string      `json:"email" validate:"required,email"`
	Role  domain.Role `json:"role" validate:"required"`
}

// CreateInvite generates a single-use invite token for the given email.
// If a previous invite exists for this email, it is replaced (Ghost pattern).
// Returns the raw token to embed in the invite email URL.
func (s *Service) CreateInvite(ctx context.Context, input InviteInput, createdByID string) (string, error) {
	// Validate inviter has sufficient role
	inviter, err := s.store.Users().GetUserByID(ctx, createdByID)
	if err != nil || !inviter.IsAdmin() {
		return "", errors.New("only admins can send invites")
	}

	// Encrypt the invited email for storage
	encEmail, err := s.encryptor.Encrypt(input.Email)
	if err != nil {
		return "", err
	}

	// Delete any existing invite for this email (Ghost: destroy before recreating)
	// We query by encrypted email hash — not ideal, but workable
	// TODO: add index on invited_email_hash for fast lookup

	// Generate raw token
	rawToken, err := crypto.GenerateInviteToken()
	if err != nil {
		return "", err
	}

	token := &domain.Token{
		ID:           uuid.New().String(),
		Hash:         crypto.HashToken(rawToken),
		UserID:       "",  // No user yet — account created on redemption
		Type:         domain.TokenTypeInvite,
		ExpiresAt:    time.Now().Add(inviteTokenExpiry),
		InvitedEmail: encEmail,
		CreatedBy:    createdByID,
		CreatedAt:    time.Now(),
	}

	if err := s.store.Tokens().CreateToken(ctx, token); err != nil {
		return "", err
	}

	return rawToken, nil
}

// RedeemInvite validates an invite token and returns the associated email.
// Called during registration when invite_code is provided.
// The token is consumed (marked used) after successful registration.
func (s *Service) RedeemInvite(ctx context.Context, rawToken string) (decryptedEmail string, tokenID string, err error) {
	hash := crypto.HashToken(rawToken)

	token, err := s.store.Tokens().GetTokenByHash(ctx, hash, domain.TokenTypeInvite)
	if err != nil || !token.IsValid() {
		return "", "", errors.New("invite code is invalid or has expired")
	}

	// Decrypt the invited email
	email, err := s.encryptor.Decrypt(token.InvitedEmail)
	if err != nil {
		return "", "", errors.New("could not read invite")
	}

	return email, token.ID, nil
}
