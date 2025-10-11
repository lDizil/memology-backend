package meme

import (
	"context"

	"memology-backend/internal/models"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, meme *models.Meme) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Meme, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Meme, error)
	Update(ctx context.Context, meme *models.Meme) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.Meme, error)
}

type MetricsRepository interface {
	Create(ctx context.Context, metrics *models.MemeMetrics) error
	GetByMemeID(ctx context.Context, memeID uuid.UUID) (*models.MemeMetrics, error)
	Update(ctx context.Context, metrics *models.MemeMetrics) error
	IncrementClick(ctx context.Context, memeID uuid.UUID) error
	IncrementDownload(ctx context.Context, memeID uuid.UUID) error
	UpdateRating(ctx context.Context, memeID uuid.UUID, delta int) error
}
