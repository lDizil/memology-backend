package repository

import (
	"context"
	"time"

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
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, search string) ([]*models.Meme, error)
	GetPublicMemes(ctx context.Context, limit, offset int, search string) ([]*models.Meme, error)
	Update(ctx context.Context, meme *models.Meme) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.Meme, error)
	CountByUserID(ctx context.Context, userID uuid.UUID, search string) (int64, error)
	CountPublicMemes(ctx context.Context, search string) (int64, error)
	Count(ctx context.Context) (int64, error)
	FindStuckMemes(ctx context.Context, olderThan time.Duration) ([]*models.Meme, error)
}

type MetricsRepository interface {
	Create(ctx context.Context, metrics *models.MemeMetrics) error
	GetByMemeID(ctx context.Context, memeID uuid.UUID) (*models.MemeMetrics, error)
	Update(ctx context.Context, metrics *models.MemeMetrics) error
	IncrementClick(ctx context.Context, memeID uuid.UUID) error
	IncrementDownload(ctx context.Context, memeID uuid.UUID) error
	UpdateRating(ctx context.Context, memeID uuid.UUID, delta int) error
}
