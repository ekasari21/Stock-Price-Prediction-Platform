package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"stockwise-ai/internal/model"
	"stockwise-ai/internal/service"
)

type StockHandler struct {
	stockService service.StockService
}

func NewStockHandler(stockService service.StockService) *StockHandler {
	return &StockHandler{stockService: stockService}
}

// GetAllStocks godoc
// @Summary Get list of tracked stocks
// @Tags Stocks
// @Produce json
// @Success 200 {object} model.ApiResponse
// @Router /api/v1/stocks [get]
func (h *StockHandler) GetAllStocks(c *gin.Context) {
	stocks, err := h.stockService.GetAllStocks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Failed to fetch stocks: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(stocks))
}

// GetStockBySymbol godoc
// @Summary Get stock detail by symbol
// @Tags Stocks
// @Produce json
// @Param symbol path string true "Stock Symbol"
// @Success 200 {object} model.ApiResponse
// @Router /api/v1/stocks/:symbol [get]
func (h *StockHandler) GetStockBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")
	stock, err := h.stockService.GetStockBySymbol(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Database error: "+err.Error()))
		return
	}
	if stock == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse("Stock not found"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(stock))
}
