package repository

import (
	"context"
	"errors"
	"time"

	"memology-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *models.UserSession) error {
	session.ID = uuid.New()
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *sessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error) {
	var session models.UserSession
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &session, err
}

func (r *sessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.UserSession, error) {
	var sessions []*models.UserSession
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&sessions).Error
	return sessions, err
}

func (r *sessionRepository) Update(ctx context.Context, session *models.UserSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *sessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.UserSession{}, id).Error
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserSession{}).Error
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.UserSession{}).Error
}
