package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"memology-backend/internal/models"
	"memology-backend/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrMemeNotFound = errors.New("meme not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidFile  = errors.New("invalid file")
)

type memeService struct {
	memeRepo repository.MemeRepository
	minioSvc MinIOService
}

func NewMemeService(memeRepo repository.MemeRepository, minioSvc MinIOService) MemeService {
	return &memeService{
		memeRepo: memeRepo,
		minioSvc: minioSvc,
	}
}

func (s *memeService) CreateMeme(ctx context.Context, userID uuid.UUID, req CreateMemeRequest) (*models.Meme, error) {
	meme := &models.Meme{
		UserID: userID,
		Prompt: req.Prompt,
		Status: "pending",
	}

	if err := s.memeRepo.Create(ctx, meme); err != nil {
		return nil, fmt.Errorf("failed to create meme: %w", err)
	}

	return meme, nil
}

func (s *memeService) UploadMemeImage(ctx context.Context, memeID uuid.UUID, file *multipart.FileHeader) error {
	objectName, err := s.minioSvc.UploadMeme(ctx, file)
	if err != nil {
		return fmt.Errorf("failed to upload image: %w", err)
	}

	meme, err := s.memeRepo.GetByID(ctx, memeID)
	if err != nil {
		s.minioSvc.DeleteMeme(ctx, objectName)
		return fmt.Errorf("failed to get meme: %w", err)
	}

	meme.ImageURL = s.minioSvc.GetMemeURL(objectName)
	meme.Status = "completed"

	if err := s.memeRepo.Update(ctx, meme); err != nil {
		s.minioSvc.DeleteMeme(ctx, objectName)
		return fmt.Errorf("failed to update meme: %w", err)
	}

	return nil
}

func (s *memeService) GetMeme(ctx context.Context, memeID uuid.UUID) (*models.Meme, error) {
	meme, err := s.memeRepo.GetByID(ctx, memeID)
	if err != nil {
		return nil, ErrMemeNotFound
	}
	return meme, nil
}

func (s *memeService) GetUserMemes(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Meme, error) {
	return s.memeRepo.GetByUserID(ctx, userID, limit, offset)
}

func (s *memeService) GetAllMemes(ctx context.Context, limit, offset int) ([]*models.Meme, error) {
	return s.memeRepo.List(ctx, limit, offset)
}

func (s *memeService) DeleteMeme(ctx context.Context, userID, memeID uuid.UUID) error {
	meme, err := s.memeRepo.GetByID(ctx, memeID)
	if err != nil {
		return ErrMemeNotFound
	}

	if meme.UserID != userID {
		return ErrUnauthorized
	}

	if err := s.memeRepo.Delete(ctx, memeID); err != nil {
		return fmt.Errorf("failed to delete meme: %w", err)
	}

	return nil
}
