package handlers

import (
	"net/http"
	"strconv"

	"memology-backend/internal/models"
	"memology-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type MemeHandler struct {
	memeService services.MemeService
	validator   *validator.Validate
}

func NewMemeHandler(memeService services.MemeService) *MemeHandler {
	return &MemeHandler{
		memeService: memeService,
		validator:   validator.New(),
	}
}

type MemeHistoryResponse struct {
	Memes []*models.Meme `json:"memes"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// @Summary Generate new meme
// @Description Generate meme from user input using neural network. Style and is_public are optional. Returns meme with pending status and task_id for checking progress. By default, memes are public.
// @Tags memes
// @Accept json
// @Produce json
// @Param request body services.CreateMemeRequest true "Meme generation request"
// @Success 201 {object} models.Meme
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /memes/generate [post]
func (h *MemeHandler) GenerateMeme(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var req services.CreateMemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if req.Prompt == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "prompt is required"})
		return
	}

	meme, err := h.memeService.CreateMeme(c.Request.Context(), userID.(uuid.UUID), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, meme)
}

// @Summary Get meme by ID
// @Description Get meme details by ID. Private memes can only be viewed by their owner.
// @Tags memes
// @Produce json
// @Param id path string true "Meme ID"
// @Success 200 {object} models.Meme
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /memes/{id} [get]
func (h *MemeHandler) GetMeme(c *gin.Context) {
	memeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid meme ID"})
		return
	}

	meme, err := h.memeService.GetMeme(c.Request.Context(), memeID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "meme not found"})
		return
	}

	if !meme.IsPublic {
		userID, exists := c.Get("user_id")
		if !exists || meme.UserID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "access denied to private meme"})
			return
		}
	}

	c.JSON(http.StatusOK, meme)
}

// @Summary Get user memes
// @Description Get list of memes created by current user with pagination
// @Tags memes
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} MemeHistoryResponse
// @Failure 401 {object} ErrorResponse
// @Router /memes/my [get]
func (h *MemeHandler) GetMyMemes(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	page := 1
	if p, exists := c.GetQuery("page"); exists {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	limit := 20
	if l, exists := c.GetQuery("limit"); exists {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	offset := (page - 1) * limit

	memes, total, err := h.memeService.GetUserMemes(c.Request.Context(), userID.(uuid.UUID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MemeHistoryResponse{
		Memes: memes,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// @Summary Get public memes
// @Description Get paginated list of public memes
// @Tags memes
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} MemeHistoryResponse
// @Router /memes/public [get]
func (h *MemeHandler) GetPublicMemes(c *gin.Context) {
	page := 1
	if p, exists := c.GetQuery("page"); exists {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	limit := 20
	if l, exists := c.GetQuery("limit"); exists {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	offset := (page - 1) * limit

	memes, total, err := h.memeService.GetPublicMemes(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MemeHistoryResponse{
		Memes: memes,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// @Summary Get all memes
// @Description Get paginated list of all memes (admin only)
// @Tags memes
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} MemeHistoryResponse
// @Router /memes [get]
func (h *MemeHandler) GetAllMemes(c *gin.Context) {
	page := 1
	if p, exists := c.GetQuery("page"); exists {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	limit := 20
	if l, exists := c.GetQuery("limit"); exists {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	offset := (page - 1) * limit

	memes, total, err := h.memeService.GetAllMemes(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MemeHistoryResponse{
		Memes: memes,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// @Summary Delete meme
// @Description Delete meme by ID (only owner can delete)
// @Tags memes
// @Produce json
// @Param id path string true "Meme ID"
// @Success 200 {object} MessageResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /memes/{id} [delete]
func (h *MemeHandler) DeleteMeme(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	memeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid meme ID"})
		return
	}

	err = h.memeService.DeleteMeme(c.Request.Context(), userID.(uuid.UUID), memeID)
	if err != nil {
		if err == services.ErrMemeNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "meme not found"})
			return
		}
		if err == services.ErrUnauthorized {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "unauthorized to delete this meme"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "meme deleted successfully"})
}

// @Summary Check meme generation status
// @Description Check if meme generation is completed and fetch result if ready
// @Tags memes
// @Produce json
// @Param id path string true "Meme ID"
// @Success 200 {object} models.Meme
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /memes/{id}/status [get]
func (h *MemeHandler) CheckMemeStatus(c *gin.Context) {
	memeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid meme ID"})
		return
	}

	meme, err := h.memeService.CheckTaskStatus(c.Request.Context(), memeID)
	if err != nil {
		if err == services.ErrMemeNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "meme not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, meme)
}

// @Summary Get available meme styles
// @Description Get list of available meme generation styles from neural network
// @Tags memes
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} ErrorResponse
// @Router /memes/styles [get]
func (h *MemeHandler) GetAvailableStyles(c *gin.Context) {
	styles, err := h.memeService.GetAvailableStyles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, styles)
}
