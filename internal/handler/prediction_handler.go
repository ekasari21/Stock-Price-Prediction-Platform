package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"stockwise-ai/internal/middleware"
	"stockwise-ai/internal/model"
	"stockwise-ai/internal/service"
)

type PredictionHandler struct {
	predictionService service.PredictionService
}

func NewPredictionHandler(predictionService service.PredictionService) *PredictionHandler {
	return &PredictionHandler{predictionService: predictionService}
}

// Forecast godoc
// @Summary Run stock forecast with selected AI/ML model
// @Tags Predictions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body model.PredictionRequest true "Forecast Request"
// @Success 200 {object} model.ApiResponse
// @Router /api/v1/predictions/forecast [post]
func (h *PredictionHandler) Forecast(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.CtxUserID)
	var userID uint
	if exists {
		userID = userIDVal.(uint)
	}

	var req model.PredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	result, err := h.predictionService.RunPrediction(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Forecast failed: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(result, "Forecast completed successfully"))
}

// GetHistory godoc
// @Summary Get user's past predictions
// @Tags Predictions
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit results (default 10)"
// @Success 200 {object} model.ApiResponse
// @Router /api/v1/predictions/history [get]
func (h *PredictionHandler) GetHistory(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse("Unauthorized"))
		return
	}
	userID := userIDVal.(uint)

	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	history, err := h.predictionService.GetUserHistory(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Failed to fetch history: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(history))
}

// GetByID godoc
// @Summary Get detailed forecast by ID
// @Tags Predictions
// @Security BearerAuth
// @Produce json
// @Param id path int true "Prediction ID"
// @Success 200 {object} model.ApiResponse
// @Router /api/v1/predictions/:id [get]
func (h *PredictionHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid prediction ID"))
		return
	}

	result, err := h.predictionService.GetPredictionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse("Prediction not found: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(result))
}
