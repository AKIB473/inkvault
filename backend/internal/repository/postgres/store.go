// Package postgres implements the repository.Store interface using pgx + PostgreSQL.
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/you/inkvault/internal/repository"
)

// Store is the Postgres implementation of repository.Store.
type Store struct {
	pool     *pgxpool.Pool
	users    *userRepo
	tokens   *tokenRepo
	sessions *sessionRepo
	blogs    *blogRepo
	posts    *postRepo
	audit    *auditRepo
	media    *mediaRepo
}

// New connects to Postgres and returns a Store.
func New(ctx context.Context, dsn string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 25
	cfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	s := &Store{pool: pool}
	s.users = &userRepo{pool: pool}
	s.tokens = &tokenRepo{pool: pool}
	s.sessions = &sessionRepo{pool: pool}
	s.blogs = &blogRepo{pool: pool}
	s.posts = &postRepo{pool: pool}
	s.audit = &auditRepo{pool: pool}
	s.media = &mediaRepo{pool: pool}

	return s, nil
}

func (s *Store) Close() { s.pool.Close() }

func (s *Store) Users() repository.UserRepository    { return s.users }
func (s *Store) Tokens() repository.TokenRepository   { return s.tokens }
func (s *Store) Sessions() repository.SessionRepository { return s.sessions }
func (s *Store) Blogs() repository.BlogRepository     { return s.blogs }
func (s *Store) Posts() repository.PostRepository     { return s.posts }
func (s *Store) Audit() repository.AuditRepository    { return s.audit }
func (s *Store) Media() repository.MediaRepository    { return s.media }
