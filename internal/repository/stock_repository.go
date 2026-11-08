package repository

import (
	"errors"

	"gorm.io/gorm"

	"stockwise-ai/internal/model"
)

type StockRepository interface {
	GetAll() ([]model.Stock, error)
	FindBySymbol(symbol string) (*model.Stock, error)
	CreateOrUpdate(stock *model.Stock) error
}

type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) StockRepository {
	return &stockRepository{db: db}
}

func (r *stockRepository) GetAll() ([]model.Stock, error) {
	var stocks []model.Stock
	err := r.db.Order("symbol ASC").Find(&stocks).Error
	return stocks, err
}

func (r *stockRepository) FindBySymbol(symbol string) (*model.Stock, error) {
	var stock model.Stock
	err := r.db.Where("UPPER(symbol) = UPPER(?)", symbol).First(&stock).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &stock, nil
}

func (r *stockRepository) CreateOrUpdate(stock *model.Stock) error {
	existing, err := r.FindBySymbol(stock.Symbol)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.db.Create(stock).Error
	}
	stock.ID = existing.ID
	return r.db.Save(stock).Error
}
