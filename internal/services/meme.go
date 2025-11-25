package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"

	"memology-backend/internal/models"
	"memology-backend/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrMemeNotFound = errors.New("meme not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidFile  = errors.New("invalid file")
	ErrTaskPending  = errors.New("task is still pending")
)

type memeService struct {
	memeRepo      repository.MemeRepository
	minioSvc      MinIOService
	aiSvc         AIService
	taskProcessor *TaskProcessor
}

func NewMemeService(memeRepo repository.MemeRepository, minioSvc MinIOService, aiSvc AIService) MemeService {
	return &memeService{
		memeRepo:      memeRepo,
		minioSvc:      minioSvc,
		aiSvc:         aiSvc,
		taskProcessor: nil,
	}
}

func NewMemeServiceWithProcessor(memeRepo repository.MemeRepository, minioSvc MinIOService, aiSvc AIService, taskProcessor *TaskProcessor) MemeService {
	return &memeService{
		memeRepo:      memeRepo,
		minioSvc:      minioSvc,
		aiSvc:         aiSvc,
		taskProcessor: taskProcessor,
	}
}

func (s *memeService) CreateMeme(ctx context.Context, userID uuid.UUID, req CreateMemeRequest) (*models.Meme, error) {
	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	meme := &models.Meme{
		UserID:   userID,
		Prompt:   req.Prompt,
		Style:    req.Style,
		Status:   "pending",
		IsPublic: isPublic,
	}

	taskID, err := s.aiSvc.GenerateMeme(ctx, req.Prompt, req.Style)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI task: %w", err)
	}

	meme.TaskID = taskID

	if err := s.memeRepo.Create(ctx, meme); err != nil {
		return nil, fmt.Errorf("failed to create meme: %w", err)
	}

	if s.taskProcessor != nil {
		if err := s.taskProcessor.AddTask(meme.ID); err != nil {
			log.Printf("Warning: failed to add task to processor: %v", err)
		}
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

func (s *memeService) GetPublicMemes(ctx context.Context, limit, offset int) ([]*models.Meme, error) {
	return s.memeRepo.GetPublicMemes(ctx, limit, offset)
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

func (s *memeService) CheckTaskStatus(ctx context.Context, memeID uuid.UUID) (*models.Meme, error) {
	meme, err := s.memeRepo.GetByID(ctx, memeID)
	if err != nil {
		return nil, ErrMemeNotFound
	}

	if meme.Status == "completed" {
		return meme, nil
	}

	if meme.TaskID == "" {
		return nil, fmt.Errorf("meme has no task ID")
	}

	taskStatus, err := s.aiSvc.GetTaskStatus(ctx, meme.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to check task status: %w", err)
	}

	if taskStatus.Status == "completed" || taskStatus.Status == "SUCCESS" {
		if err := s.ProcessCompletedTask(ctx, memeID); err != nil {
			return nil, fmt.Errorf("failed to process completed task: %w", err)
		}
		return s.memeRepo.GetByID(ctx, memeID)
	}

	meme.Status = taskStatus.Status
	if err := s.memeRepo.Update(ctx, meme); err != nil {
		return nil, fmt.Errorf("failed to update meme status: %w", err)
	}

	return meme, nil
}

func (s *memeService) ProcessCompletedTask(ctx context.Context, memeID uuid.UUID) error {
	meme, err := s.memeRepo.GetByID(ctx, memeID)
	if err != nil {
		return ErrMemeNotFound
	}

	if meme.TaskID == "" {
		return fmt.Errorf("meme has no task ID")
	}

	imageData, err := s.aiSvc.GetTaskResult(ctx, meme.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task result: %w", err)
	}

	objectName := fmt.Sprintf("memes/%s.jpg", meme.ID.String())

	if err := s.minioSvc.UploadBytes(ctx, objectName, imageData); err != nil {
		return fmt.Errorf("failed to upload image to MinIO: %w", err)
	}

	meme.ImageURL = s.minioSvc.GetMemeURL(objectName)
	meme.Status = "completed"

	if err := s.memeRepo.Update(ctx, meme); err != nil {
		s.minioSvc.DeleteMeme(ctx, objectName)
		return fmt.Errorf("failed to update meme: %w", err)
	}

	return nil
}

func (s *memeService) GetAvailableStyles(ctx context.Context) ([]string, error) {
	return s.aiSvc.GetAvailableStyles(ctx)
}
