package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"memology-backend/internal/models"
	"memology-backend/internal/repository"
	"memology-backend/pkg/auth"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
	ErrInvalidToken       = errors.New("invalid token")
)

type authService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	jwtManager  *auth.JWTManager
}

func NewAuthService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, jwtManager *auth.JWTManager) AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtManager:  jwtManager,
	}
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserExists
	}

	if req.Email != "" {
		existingUser, err = s.userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil {
			return nil, ErrUserExists
		}
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	var user *models.User
	var err error

	user, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = s.userRepo.GetByEmail(ctx, req.Username)
		if err != nil {
			return nil, err
		}
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if !auth.VerifyPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	tokenHash := s.hashToken(refreshToken)
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrInvalidToken
	}

	if session.ExpiresAt.Before(time.Now()) {
		s.sessionRepo.Delete(ctx, session.ID)
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, ErrUserInactive
	}

	s.sessionRepo.Delete(ctx, session.ID)
	return s.generateAuthResponse(ctx, user)
}

func (s *authService) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	tokenHash := s.hashToken(refreshToken)
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}
	if session != nil {
		return s.sessionRepo.Delete(ctx, session.ID)
	}
	return nil
}

func (s *authService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

func (s *authService) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, ErrUserInactive
	}

	return &TokenClaims{
		UserID:   claims.UserID,
		Username: claims.Username,
		IsActive: claims.IsActive,
	}, nil
}

func (s *authService) generateAuthResponse(ctx context.Context, user *models.User) (*AuthResponse, error) {
	accessToken, refreshToken, err := s.jwtManager.GenerateTokens(user.ID, user.Username, user.IsActive)
	if err != nil {
		return nil, err
	}

	tokenHash := s.hashToken(refreshToken)
	session := &models.UserSession{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
	}, nil
}

func (s *authService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
