package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"stockwise-ai/internal/middleware"
	"stockwise-ai/internal/model"
	"stockwise-ai/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary User registration
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Registration Info"
// @Success 201 {object} model.ApiResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Validation error: "+err.Error()))
		return
	}

	res, err := h.authService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, model.SuccessResponse(res, "User registered successfully"))
}

// Login godoc
// @Summary User authentication
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login Credentials"
// @Success 200 {object} model.ApiResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Validation error: "+err.Error()))
		return
	}

	res, err := h.authService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(res, "Login successful"))
}

// RefreshToken godoc
// @Summary Refresh JWT access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} model.ApiResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Validation error: "+err.Error()))
		return
	}

	res, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(res, "Token refreshed successfully"))
}

// GetCurrentUser godoc
// @Summary Get authenticated user profile
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.ApiResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse("Unauthorized"))
		return
	}

	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse("User not found"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(user.ToResponse()))
}
