package session

import (
	"context"

	"memology-backend/internal/models"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, session *models.UserSession) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.UserSession, error)
	Update(ctx context.Context, session *models.UserSession) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}
