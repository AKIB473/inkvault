package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/you/inkvault/internal/domain"
	"github.com/you/inkvault/internal/repository"
)

type tokenRepo struct{ pool *pgxpool.Pool }

func (r *tokenRepo) CreateToken(ctx context.Context, t *domain.Token) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO tokens
			(id, hash, user_id, type, expires_at, invited_email, created_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		t.ID, t.Hash, nilIfEmpty(t.UserID), string(t.Type),
		t.ExpiresAt, nilIfEmpty(t.InvitedEmail), nilIfEmpty(t.CreatedBy), t.CreatedAt,
	)
	return err
}

func (r *tokenRepo) GetTokenByHash(ctx context.Context, hash string, tokenType domain.TokenType) (*domain.Token, error) {
	t := &domain.Token{}
	var typeStr string
	var userID, invitedEmail, createdBy *string

	err := r.pool.QueryRow(ctx, `
		SELECT id,hash,user_id,type,used_at,expires_at,invited_email,created_by,created_at
		FROM tokens WHERE hash=$1 AND type=$2`,
		hash, string(tokenType)).
		Scan(&t.ID, &t.Hash, &userID, &typeStr,
			&t.UsedAt, &t.ExpiresAt, &invitedEmail, &createdBy, &t.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	t.Type = domain.TokenType(typeStr)
	if userID != nil {
		t.UserID = *userID
	}
	if invitedEmail != nil {
		t.InvitedEmail = *invitedEmail
	}
	if createdBy != nil {
		t.CreatedBy = *createdBy
	}
	return t, nil
}

func (r *tokenRepo) MarkTokenUsed(ctx context.Context, tokenID string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE tokens SET used_at=$2 WHERE id=$1`,
		tokenID, now,
	)
	return err
}

func (r *tokenRepo) DeleteExpiredTokens(ctx context.Context) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM tokens WHERE expires_at < $1`, time.Now(),
	)
	return err
}

func (r *tokenRepo) DeleteTokensByUser(ctx context.Context, userID string, tokenType domain.TokenType) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM tokens WHERE user_id=$1 AND type=$2`,
		userID, string(tokenType),
	)
	return err
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
