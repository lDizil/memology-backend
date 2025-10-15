package handlers

import (
	"net/http"
	"strconv"

	"memology-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService services.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

// @Summary Register new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body services.RegisterRequest true "Registration data"
// @Success 201 {object} services.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	response, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == services.ErrUserExists {
			status = http.StatusConflict
		}
		c.JSON(status, ErrorResponse{Error: err.Error()})
		return
	}

	setAuthCookies(c, response.AccessToken, response.RefreshToken, response.ExpiresIn)

	c.JSON(http.StatusCreated, response)
}

// @Summary User login
// @Description Login with username or email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body services.LoginRequest true "Login credentials"
// @Success 200 {object} services.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	response, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == services.ErrInvalidCredentials || err == services.ErrUserInactive {
			status = http.StatusUnauthorized
		}
		c.JSON(status, ErrorResponse{Error: err.Error()})
		return
	}

	setAuthCookies(c, response.AccessToken, response.RefreshToken, response.ExpiresIn)

	c.JSON(http.StatusOK, response)
}

// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} services.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	setAuthCookies(c, response.AccessToken, response.RefreshToken, response.ExpiresIn)

	c.JSON(http.StatusOK, response)
}

// @Summary Logout user
// @Description Logout user and invalidate refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := h.authService.Logout(c.Request.Context(), userID.(uuid.UUID), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	clearAuthCookies(c)

	c.JSON(http.StatusOK, MessageResponse{Message: "logged out successfully"})
}

// @Summary Logout from all devices
// @Description Logout user from all devices
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	err := h.authService.LogoutAll(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Clear cookies
	clearAuthCookies(c)

	c.JSON(http.StatusOK, MessageResponse{Message: "logged out from all devices"})
}

type UserHandler struct {
	userService services.UserService
	validator   *validator.Validate
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator.New(),
	}
}

// @Summary Get user profile
// @Description Get current user profile
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		status := http.StatusInternalServerError
		if err == services.ErrUserNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary Update user profile
// @Description Update current user profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.UpdateProfileRequest true "Profile update data"
// @Success 200 {object} models.User
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /users/profile/update [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID.(uuid.UUID), req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == services.ErrUserExists {
			status = http.StatusConflict
		}
		c.JSON(status, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary Change password
// @Description Change user password
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.ChangePasswordRequest true "Password change data"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /users/change-password [post]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var req services.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID.(uuid.UUID), req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == services.ErrInvalidCredentials {
			status = http.StatusUnauthorized
		}
		c.JSON(status, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "password changed successfully"})
}

// @Summary Get users list
// @Description Get paginated users list
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} models.User
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /users/list [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, err := h.userService.GetUsers(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// setAuthCookies sets access and refresh tokens in HTTP-only cookies
func setAuthCookies(c *gin.Context, accessToken, refreshToken string, expiresIn int64) {
	// Access token cookie (short-lived)
	c.SetCookie(
		"access_token", // name
		accessToken,    // value
		int(expiresIn), // maxAge in seconds
		"/",            // path
		"",             // domain (empty = current domain)
		false,          // secure (set true in production with HTTPS)
		true,           // httpOnly (prevents JS access)
	)

	// Refresh token cookie (long-lived, 7 days = 604800 seconds)
	c.SetCookie(
		"refresh_token",
		refreshToken,
		604800, // 7 days
		"/",
		"",
		false, // secure
		true,  // httpOnly
	)
}

// clearAuthCookies removes auth cookies by setting them to expired
func clearAuthCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
}
