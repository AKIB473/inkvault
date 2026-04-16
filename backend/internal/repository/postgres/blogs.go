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

type blogRepo struct{ pool *pgxpool.Pool }

func (r *blogRepo) CreateBlog(ctx context.Context, b *domain.Blog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO blogs
			(id,owner_id,slug,title,description,domain,visibility,password_hash,
			 style_sheet,language,rtl,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		b.ID, b.OwnerID, b.Slug, b.Title, b.Description,
		nilIfEmpty(b.Domain), visibilityString(b.Visibility), nilIfEmpty(b.PasswordHash),
		nilIfEmpty(b.StyleSheet), b.Language, b.RTL,
		b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (r *blogRepo) GetBlogByID(ctx context.Context, id string) (*domain.Blog, error) {
	return r.scanBlog(r.pool.QueryRow(ctx, `
		SELECT id,owner_id,slug,title,description,domain,visibility,password_hash,
		       style_sheet,language,rtl,created_at,updated_at
		FROM blogs WHERE id=$1`, id))
}

func (r *blogRepo) GetBlogBySlug(ctx context.Context, slug string) (*domain.Blog, error) {
	return r.scanBlog(r.pool.QueryRow(ctx, `
		SELECT id,owner_id,slug,title,description,domain,visibility,password_hash,
		       style_sheet,language,rtl,created_at,updated_at
		FROM blogs WHERE slug=$1`, slug))
}

func (r *blogRepo) GetBlogByDomain(ctx context.Context, d string) (*domain.Blog, error) {
	return r.scanBlog(r.pool.QueryRow(ctx, `
		SELECT id,owner_id,slug,title,description,domain,visibility,password_hash,
		       style_sheet,language,rtl,created_at,updated_at
		FROM blogs WHERE domain=$1`, d))
}

func (r *blogRepo) GetBlogsByOwner(ctx context.Context, ownerID string) ([]*domain.Blog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,owner_id,slug,title,description,domain,visibility,password_hash,
		       style_sheet,language,rtl,created_at,updated_at
		FROM blogs WHERE owner_id=$1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var blogs []*domain.Blog
	for rows.Next() {
		b, err := r.scanBlog(rows)
		if err != nil {
			return nil, err
		}
		blogs = append(blogs, b)
	}
	return blogs, rows.Err()
}

func (r *blogRepo) UpdateBlog(ctx context.Context, b *domain.Blog) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE blogs SET
			title=$2,description=$3,visibility=$4,style_sheet=$5,
			language=$6,rtl=$7,updated_at=$8
		WHERE id=$1`,
		b.ID, b.Title, b.Description, visibilityString(b.Visibility),
		nilIfEmpty(b.StyleSheet), b.Language, b.RTL, time.Now(),
	)
	return err
}

func (r *blogRepo) DeleteBlog(ctx context.Context, id, ownerID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM blogs WHERE id=$1 AND owner_id=$2`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *blogRepo) scanBlog(row pgx.Row) (*domain.Blog, error) {
	b := &domain.Blog{}
	var vis, dom, pwHash, styleSheet *string
	err := row.Scan(
		&b.ID, &b.OwnerID, &b.Slug, &b.Title, &b.Description,
		&dom, &vis, &pwHash, &styleSheet,
		&b.Language, &b.RTL, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	if vis != nil {
		b.Visibility = domain.PostVisibility(visibilityToInt(*vis))
	}
	if dom != nil {
		b.Domain = *dom
	}
	if pwHash != nil {
		b.PasswordHash = *pwHash
	}
	if styleSheet != nil {
		b.StyleSheet = *styleSheet
	}
	return b, nil
}

func visibilityToInt(s string) int {
	switch s {
	case "unlisted":
		return 1
	case "members_only":
		return 2
	case "private":
		return 3
	case "password_protected":
		return 4
	default:
		return 0
	}
}
