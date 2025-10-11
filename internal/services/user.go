package services

import (
	"context"

	"memology-backend/internal/models"
	"memology-backend/internal/repository"
	"memology-backend/pkg/auth"

	"github.com/google/uuid"
)

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if req.Username != "" {
		existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != userID {
			return nil, ErrUserExists
		}
		user.Username = req.Username
	}

	if req.Email != "" {
		existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != userID {
			return nil, ErrUserExists
		}
		user.Email = req.Email
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if !auth.VerifyPassword(req.CurrentPassword, user.PasswordHash) {
		return ErrInvalidCredentials
	}

	newPasswordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = newPasswordHash
	return s.userRepo.Update(ctx, user)
}

func (s *userService) UploadAvatar(ctx context.Context, userID uuid.UUID, fileData []byte, filename string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	avatarURL := "/avatars/" + userID.String() + ".jpg"
	user.AvatarURL = avatarURL

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.Delete(ctx, userID)
}

func (s *userService) GetUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return s.userRepo.List(ctx, limit, offset)
}
