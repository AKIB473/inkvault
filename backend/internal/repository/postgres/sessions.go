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

type sessionRepo struct{ pool *pgxpool.Pool }

func (r *sessionRepo) CreateSession(ctx context.Context, s *domain.Session) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions
			(id, user_id, refresh_hash, device_info, ip_address, user_agent, last_seen_at, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		s.ID, s.UserID, s.RefreshHash, s.DeviceInfo,
		s.IPAddress, s.UserAgent, s.LastSeenAt, s.ExpiresAt, s.CreatedAt,
	)
	return err
}

func (r *sessionRepo) GetSessionByRefreshHash(ctx context.Context, hash string) (*domain.Session, error) {
	s := &domain.Session{}
	err := r.pool.QueryRow(ctx, `
		SELECT id,user_id,refresh_hash,device_info,ip_address,user_agent,last_seen_at,expires_at,created_at
		FROM sessions WHERE refresh_hash=$1`, hash).
		Scan(&s.ID, &s.UserID, &s.RefreshHash, &s.DeviceInfo,
			&s.IPAddress, &s.UserAgent, &s.LastSeenAt, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return s, nil
}

func (r *sessionRepo) UpdateSessionLastSeen(ctx context.Context, sessionID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sessions SET last_seen_at=$2 WHERE id=$1`,
		sessionID, time.Now(),
	)
	return err
}

func (r *sessionRepo) DeleteSession(ctx context.Context, sessionID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE id=$1`, sessionID)
	return err
}

func (r *sessionRepo) DeleteAllUserSessions(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id=$1`, userID)
	return err
}

func (r *sessionRepo) ListUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,user_id,refresh_hash,device_info,ip_address,user_agent,last_seen_at,expires_at,created_at
		FROM sessions WHERE user_id=$1 ORDER BY last_seen_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		s := &domain.Session{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.RefreshHash, &s.DeviceInfo,
			&s.IPAddress, &s.UserAgent, &s.LastSeenAt, &s.ExpiresAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}
