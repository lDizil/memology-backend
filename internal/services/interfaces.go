package services

import (
	"context"

	"memology-backend/internal/models"

	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
}

type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*models.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error
	UploadAvatar(ctx context.Context, userID uuid.UUID, fileData []byte, filename string) (*models.User, error)
	DeleteAccount(ctx context.Context, userID uuid.UUID) error
	GetUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50" example:"johndoe"`
	Email    string `json:"email" validate:"required,email" example:"john@example.com"`
	Password string `json:"password" validate:"required,min=6" example:"password123"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required" example:"johndoe или john@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}

type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string       `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn    int64        `json:"expires_in" example:"3600"`
}

type TokenClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	IsActive bool      `json:"is_active"`
}

type UpdateProfileRequest struct {
	Username string `json:"username,omitempty" validate:"omitempty,min=3,max=50" example:"johndoe_new"`
	Email    string `json:"email,omitempty" validate:"omitempty,email" example:"john.new@example.com"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required" example:"oldpassword"`
	NewPassword     string `json:"new_password" validate:"required,min=6" example:"newpassword123"`
}
