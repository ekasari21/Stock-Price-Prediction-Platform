package service

import (
	"stockwise-ai/internal/model"
	"stockwise-ai/internal/repository"
)

type StockService interface {
	GetAllStocks() ([]model.Stock, error)
	GetStockBySymbol(symbol string) (*model.Stock, error)
	SaveOrUpdateStock(stock *model.Stock) error
}

type stockService struct {
	stockRepo repository.StockRepository
}

func NewStockService(stockRepo repository.StockRepository) StockService {
	return &stockService{stockRepo: stockRepo}
}

func (s *stockService) GetAllStocks() ([]model.Stock, error) {
	return s.stockRepo.GetAll()
}

func (s *stockService) GetStockBySymbol(symbol string) (*model.Stock, error) {
	return s.stockRepo.FindBySymbol(symbol)
}

func (s *stockService) SaveOrUpdateStock(stock *model.Stock) error {
	return s.stockRepo.CreateOrUpdate(stock)
}
