package repository

import (
	"context"

	"github.com/you/inkvault/internal/domain"
)

// MediaRepository handles media file records.
type MediaRepository interface {
	CreateMedia(ctx context.Context, m *domain.Media) error
	GetMediaByKey(ctx context.Context, key string) (*domain.Media, error)
	ListMediaByUploader(ctx context.Context, uploaderID string, limit, offset int) ([]*domain.Media, error)
	DeleteMedia(ctx context.Context, key string) error
}
