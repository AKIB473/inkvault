package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/you/inkvault/internal/domain"
	"github.com/you/inkvault/internal/repository"
)

type mediaRepo struct{ pool *pgxpool.Pool }

func (r *mediaRepo) CreateMedia(ctx context.Context, m *domain.Media) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO media (id,uploader_id,blog_id,filename,mime_type,size_bytes,storage_key,public_url,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		m.ID, m.UploaderID, nilIfEmpty(m.BlogID),
		m.Filename, m.MimeType, m.SizeBytes,
		m.StorageKey, m.PublicURL, m.CreatedAt,
	)
	return err
}

func (r *mediaRepo) GetMediaByKey(ctx context.Context, key string) (*domain.Media, error) {
	m := &domain.Media{}
	var blogID *string
	err := r.pool.QueryRow(ctx, `
		SELECT id,uploader_id,blog_id,filename,mime_type,size_bytes,storage_key,public_url,created_at
		FROM media WHERE storage_key=$1`, key).
		Scan(&m.ID, &m.UploaderID, &blogID, &m.Filename, &m.MimeType,
			&m.SizeBytes, &m.StorageKey, &m.PublicURL, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	if blogID != nil {
		m.BlogID = *blogID
	}
	return m, nil
}

func (r *mediaRepo) ListMediaByUploader(ctx context.Context, uploaderID string, limit, offset int) ([]*domain.Media, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,uploader_id,blog_id,filename,mime_type,size_bytes,storage_key,public_url,created_at
		FROM media WHERE uploader_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		uploaderID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []*domain.Media
	for rows.Next() {
		m := &domain.Media{}
		var blogID *string
		if err := rows.Scan(&m.ID, &m.UploaderID, &blogID, &m.Filename, &m.MimeType,
			&m.SizeBytes, &m.StorageKey, &m.PublicURL, &m.CreatedAt); err != nil {
			return nil, err
		}
		if blogID != nil {
			m.BlogID = *blogID
		}
		media = append(media, m)
	}
	return media, rows.Err()
}

func (r *mediaRepo) DeleteMedia(ctx context.Context, key string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM media WHERE storage_key=$1`, key)
	return err
}
