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

type postRepo struct{ pool *pgxpool.Pool }

func (r *postRepo) CreatePost(ctx context.Context, p *domain.Post) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO posts
			(id,blog_id,author_id,title,slug,excerpt,content,content_type,
			 status,visibility,is_pinned,tags,cover_image,
			 seo_title,seo_description,og_image,canonical_url,
			 reading_time_mins,published_at,scheduled_at,created_at,updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`,
		p.ID, p.BlogID, p.AuthorID, p.Title, p.Slug, p.Excerpt,
		p.Content, p.ContentType, string(p.Status), visibilityString(p.Visibility),
		p.IsPinned, p.Tags, nilIfEmpty(p.CoverImage),
		nilIfEmpty(p.SEOTitle), nilIfEmpty(p.SEODescription),
		nilIfEmpty(p.OGImage), nilIfEmpty(p.CanonicalURL),
		p.ReadingTimeMins, p.PublishedAt, p.ScheduledAt,
		p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *postRepo) GetPostByID(ctx context.Context, id string) (*domain.Post, error) {
	return r.scanPost(r.pool.QueryRow(ctx, `
		SELECT `+postCols+` FROM posts WHERE id=$1`, id))
}

func (r *postRepo) GetPostBySlug(ctx context.Context, blogID, slug string) (*domain.Post, error) {
	return r.scanPost(r.pool.QueryRow(ctx, `
		SELECT `+postCols+` FROM posts WHERE blog_id=$1 AND slug=$2`, blogID, slug))
}

func (r *postRepo) GetPostsByBlog(ctx context.Context, blogID string, status domain.PostStatus, limit, offset int) ([]*domain.Post, error) {
	query := `SELECT ` + postCols + ` FROM posts WHERE blog_id=$1`
	args := []interface{}{blogID}
	if status != "" {
		query += ` AND status=$2 ORDER BY COALESCE(published_at,created_at) DESC LIMIT $3 OFFSET $4`
		args = append(args, string(status), limit, offset)
	} else {
		query += ` ORDER BY COALESCE(published_at,created_at) DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	}
	return r.queryPosts(ctx, query, args...)
}

func (r *postRepo) GetPostsByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*domain.Post, error) {
	return r.queryPosts(ctx, `
		SELECT `+postCols+` FROM posts WHERE author_id=$1
		ORDER BY COALESCE(published_at,created_at) DESC LIMIT $2 OFFSET $3`,
		authorID, limit, offset)
}

func (r *postRepo) UpdatePost(ctx context.Context, p *domain.Post) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE posts SET
			title=$2,slug=$3,excerpt=$4,content=$5,content_type=$6,
			status=$7,visibility=$8,is_pinned=$9,tags=$10,cover_image=$11,
			seo_title=$12,seo_description=$13,og_image=$14,canonical_url=$15,
			reading_time_mins=$16,published_at=$17,scheduled_at=$18,updated_at=$19
		WHERE id=$1`,
		p.ID, p.Title, p.Slug, p.Excerpt, p.Content, p.ContentType,
		string(p.Status), visibilityString(p.Visibility),
		p.IsPinned, p.Tags, nilIfEmpty(p.CoverImage),
		nilIfEmpty(p.SEOTitle), nilIfEmpty(p.SEODescription),
		nilIfEmpty(p.OGImage), nilIfEmpty(p.CanonicalURL),
		p.ReadingTimeMins, p.PublishedAt, p.ScheduledAt, time.Now(),
	)
	return err
}

func (r *postRepo) DeletePost(ctx context.Context, id, authorID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM posts WHERE id=$1 AND author_id=$2`, id, authorID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *postRepo) IncrementViews(ctx context.Context, postID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE posts SET views=views+1 WHERE id=$1`, postID)
	return err
}

func (r *postRepo) SearchPosts(ctx context.Context, query string, limit, offset int) ([]*domain.Post, error) {
	// PostgreSQL full-text search (learned from GoBlog FTS5 + Ghost pattern)
	return r.queryPosts(ctx, `
		SELECT `+postCols+`
		FROM posts
		WHERE status='published'
		  AND visibility='public'
		  AND to_tsvector('english', coalesce(title,'') || ' ' || coalesce(excerpt,'') || ' ' || coalesce(content,''))
		      @@ plainto_tsquery('english', $1)
		ORDER BY ts_rank(
			to_tsvector('english', coalesce(title,'') || ' ' || coalesce(excerpt,'') || ' ' || coalesce(content,'')),
			plainto_tsquery('english', $1)
		) DESC
		LIMIT $2 OFFSET $3`,
		query, limit, offset,
	)
}

func (r *postRepo) CreateRevision(ctx context.Context, rev *domain.PostRevision) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO post_revisions (id,post_id,content,edited_by,created_at)
		VALUES ($1,$2,$3,$4,$5)`,
		rev.ID, rev.PostID, rev.Content, rev.EditedBy, rev.CreatedAt,
	)
	return err
}

func (r *postRepo) GetRevisionsByPost(ctx context.Context, postID string) ([]*domain.PostRevision, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,post_id,content,edited_by,created_at
		FROM post_revisions WHERE post_id=$1
		ORDER BY created_at DESC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var revs []*domain.PostRevision
	for rows.Next() {
		rev := &domain.PostRevision{}
		if err := rows.Scan(&rev.ID, &rev.PostID, &rev.Content, &rev.EditedBy, &rev.CreatedAt); err != nil {
			return nil, err
		}
		revs = append(revs, rev)
	}
	return revs, rows.Err()
}

func (r *postRepo) GetRevisionByID(ctx context.Context, id string) (*domain.PostRevision, error) {
	rev := &domain.PostRevision{}
	err := r.pool.QueryRow(ctx, `
		SELECT id,post_id,content,edited_by,created_at
		FROM post_revisions WHERE id=$1`, id).
		Scan(&rev.ID, &rev.PostID, &rev.Content, &rev.EditedBy, &rev.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return rev, nil
}

// --- helpers ---

const postCols = `id,blog_id,author_id,title,slug,excerpt,content,content_type,
	status,visibility,is_pinned,tags,cover_image,
	seo_title,seo_description,og_image,canonical_url,
	views,reading_time_mins,published_at,scheduled_at,
	last_newsletter_sent_at,created_at,updated_at`

func (r *postRepo) queryPosts(ctx context.Context, query string, args ...interface{}) ([]*domain.Post, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		p, err := r.scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepo) scanPost(row pgx.Row) (*domain.Post, error) {
	p := &domain.Post{}
	var statusStr, visStr string
	var coverImage, seoTitle, seoDesc, ogImage, canonicalURL *string

	err := row.Scan(
		&p.ID, &p.BlogID, &p.AuthorID, &p.Title, &p.Slug, &p.Excerpt,
		&p.Content, &p.ContentType,
		&statusStr, &visStr, &p.IsPinned, &p.Tags, &coverImage,
		&seoTitle, &seoDesc, &ogImage, &canonicalURL,
		&p.Views, &p.ReadingTimeMins, &p.PublishedAt, &p.ScheduledAt,
		&p.LastNewsletterSentAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	p.Status = domain.PostStatus(statusStr)
	p.Visibility = domain.PostVisibility(visibilityToInt(visStr))
	if coverImage != nil {
		p.CoverImage = *coverImage
	}
	if seoTitle != nil {
		p.SEOTitle = *seoTitle
	}
	if seoDesc != nil {
		p.SEODescription = *seoDesc
	}
	if ogImage != nil {
		p.OGImage = *ogImage
	}
	if canonicalURL != nil {
		p.CanonicalURL = *canonicalURL
	}
	return p, nil
}

func visibilityString(v domain.PostVisibility) string {
	switch v {
	case domain.VisibilityPublic:
		return "public"
	case domain.VisibilityUnlisted:
		return "unlisted"
	case domain.VisibilityMembers:
		return "members_only"
	case domain.VisibilityPrivate:
		return "private"
	case domain.VisibilityPassword:
		return "password_protected"
	default:
		return "public"
	}
}
