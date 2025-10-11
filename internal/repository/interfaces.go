package repository

import (
	"context"

	"memology-backend/internal/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *models.UserSession) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.UserSession, error)
	Update(ctx context.Context, session *models.UserSession) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type MemeRepository interface {
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
