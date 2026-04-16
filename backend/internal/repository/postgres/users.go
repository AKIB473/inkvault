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

type userRepo struct{ pool *pgxpool.Pool }

func (r *userRepo) CreateUser(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users
			(id, username, display_name, email_encrypted, password_hash, role, status, two_fa_enabled, avatar_url, bio, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		u.ID, u.Username, u.DisplayName, u.EmailEncrypted,
		u.PasswordHash, string(u.Role), int(u.Status),
		u.TwoFAEnabled, u.AvatarURL, u.Bio,
		u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (r *userRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT id,username,display_name,email_encrypted,password_hash,role,status,
		       two_fa_enabled,avatar_url,bio,created_at,updated_at
		FROM users WHERE id=$1`, id))
}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT id,username,display_name,email_encrypted,password_hash,role,status,
		       two_fa_enabled,avatar_url,bio,created_at,updated_at
		FROM users WHERE username=$1`, username))
}

func (r *userRepo) GetUserByEmailHash(ctx context.Context, emailHash string) (*domain.User, error) {
	// Note: we store the full encrypted email, not a hash.
	// This method is a placeholder for a future HMAC-indexed lookup.
	return nil, errors.New("not implemented: use username lookup")
}

func (r *userRepo) UpdateUser(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET
			display_name=$2, avatar_url=$3, bio=$4, updated_at=$5
		WHERE id=$1`,
		u.ID, u.DisplayName, u.AvatarURL, u.Bio, time.Now(),
	)
	return err
}

func (r *userRepo) UpdateEncryptedEmail(ctx context.Context, userID, encryptedEmail string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET email_encrypted=$2, updated_at=$3 WHERE id=$1`,
		userID, encryptedEmail, time.Now(),
	)
	return err
}

func (r *userRepo) UpdatePasswordHash(ctx context.Context, userID, hash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash=$2, updated_at=$3 WHERE id=$1`,
		userID, hash, time.Now(),
	)
	return err
}

func (r *userRepo) UpdateRole(ctx context.Context, userID string, role domain.Role) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET role=$2, updated_at=$3 WHERE id=$1`,
		userID, string(role), time.Now(),
	)
	return err
}

func (r *userRepo) UpdateStatus(ctx context.Context, userID string, status domain.UserStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET status=$2, updated_at=$3 WHERE id=$1`,
		userID, int(status), time.Now(),
	)
	return err
}

func (r *userRepo) Enable2FA(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET two_fa_enabled=true, updated_at=$2 WHERE id=$1`,
		userID, time.Now(),
	)
	return err
}

func (r *userRepo) Disable2FA(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET two_fa_enabled=false, updated_at=$2 WHERE id=$1`,
		userID, time.Now(),
	)
	return err
}

func (r *userRepo) DeleteAccount(ctx context.Context, userID string) error {
	// Hard delete — FK CASCADE cleans up sessions, tokens, posts, blogs, media
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, userID)
	return err
}

func (r *userRepo) ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,username,display_name,email_encrypted,password_hash,role,status,
		       two_fa_enabled,avatar_url,bio,created_at,updated_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u, err := r.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// scanUser works with both pgx.Row and pgx.Rows via the RowScanner interface.
func (r *userRepo) scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	var roleStr string
	var statusInt int
	err := row.Scan(
		&u.ID, &u.Username, &u.DisplayName, &u.EmailEncrypted,
		&u.PasswordHash, &roleStr, &statusInt,
		&u.TwoFAEnabled, &u.AvatarURL, &u.Bio,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	u.Role = domain.Role(roleStr)
	u.Status = domain.UserStatus(statusInt)
	return u, nil
}
