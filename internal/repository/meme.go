package repository

import (
	"context"
	"memology-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type memeRepository struct {
	db *gorm.DB
}

func NewMemeRepository(db *gorm.DB) MemeRepository {
	return &memeRepository{db: db}
}

func (r *memeRepository) Create(ctx context.Context, meme *models.Meme) error {
	return r.db.WithContext(ctx).Create(meme).Error
}

func (r *memeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Meme, error) {
	var meme models.Meme
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Metrics").
		First(&meme, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &meme, nil
}

func (r *memeRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Meme, error) {
	var memes []*models.Meme
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Metrics").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&memes).Error
	return memes, err
}

func (r *memeRepository) Update(ctx context.Context, meme *models.Meme) error {
	return r.db.WithContext(ctx).Save(meme).Error
}

func (r *memeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Meme{}, "id = ?", id).Error
}

func (r *memeRepository) GetPublicMemes(ctx context.Context, limit, offset int) ([]*models.Meme, error) {
	var memes []*models.Meme
	err := r.db.WithContext(ctx).
		Where("is_public = ?", true).
		Preload("User").
		Preload("Metrics").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&memes).Error
	return memes, err
}

func (r *memeRepository) List(ctx context.Context, limit, offset int) ([]*models.Meme, error) {
	var memes []*models.Meme
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Metrics").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&memes).Error
	return memes, err
}

func (r *memeRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Meme{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *memeRepository) CountPublicMemes(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Meme{}).
		Where("is_public = ?", true).
		Count(&count).Error
	return count, err
}

func (r *memeRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Meme{}).
		Count(&count).Error
	return count, err
}
