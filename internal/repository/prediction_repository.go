package repository

import (
	"errors"

	"gorm.io/gorm"

	"stockwise-ai/internal/model"
)

type PredictionRepository interface {
	Create(pred *model.Prediction) error
	GetByID(id uint) (*model.Prediction, error)
	GetByUserID(userID uint, limit int) ([]model.Prediction, error)
	GetRecentBySymbol(symbol string, limit int) ([]model.Prediction, error)
}

type predictionRepository struct {
	db *gorm.DB
}

func NewPredictionRepository(db *gorm.DB) PredictionRepository {
	return &predictionRepository{db: db}
}

func (r *predictionRepository) Create(pred *model.Prediction) error {
	return r.db.Create(pred).Error
}

func (r *predictionRepository) GetByID(id uint) (*model.Prediction, error) {
	var pred model.Prediction
	err := r.db.First(&pred, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pred, nil
}

func (r *predictionRepository) GetByUserID(userID uint, limit int) ([]model.Prediction, error) {
	var predictions []model.Prediction
	query := r.db.Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&predictions).Error
	return predictions, err
}

func (r *predictionRepository) GetRecentBySymbol(symbol string, limit int) ([]model.Prediction, error) {
	var predictions []model.Prediction
	query := r.db.Where("UPPER(symbol) = UPPER(?)", symbol).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&predictions).Error
	return predictions, err
}
