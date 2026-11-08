package model

import (
	"time"
)

type Prediction struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint      `gorm:"index" json:"user_id"`
	Symbol          string    `gorm:"type:varchar(20);index;not null" json:"symbol"`
	ModelType       string    `gorm:"type:varchar(50);not null" json:"model_type"`
	ModelName       string    `gorm:"type:varchar(100);not null" json:"model_name"`
	TargetDays      int       `json:"target_days"`
	LastPrice       float64   `gorm:"type:decimal(12,2)" json:"last_price"`
	Currency        string    `gorm:"type:varchar(10)" json:"currency"`
	MAE             float64   `gorm:"type:decimal(10,4)" json:"mae"`
	RMSE            float64   `gorm:"type:decimal(10,4)" json:"rmse"`
	MAPE            float64   `gorm:"type:decimal(10,4)" json:"mape"`
	R2Score         float64   `gorm:"type:decimal(10,4)" json:"r2_score"`
	PredictionsJSON string    `gorm:"type:text" json:"predictions_json"`
	HistoryJSON     string    `gorm:"type:text" json:"history_json"`
	FeaturesJSON    string    `gorm:"type:text" json:"features_json"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type PredictionRequest struct {
	Symbol    string `json:"symbol" binding:"required"`
	ModelType string `json:"model_type" binding:"required"` // xgboost, random_forest, linear_regression, lstm, arima
	Days      int    `json:"days"`                          // default 7
	Period    string `json:"period"`                        // default 1y
}

type PredictionPoint struct {
	Date       string  `json:"date"`
	Predicted  float64 `json:"predicted"`
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
}

type HistoryPoint struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
	SMA10  float64 `json:"sma_10"`
	SMA20  float64 `json:"sma_20"`
}

type FeatureImportance struct {
	Feature    string  `json:"feature"`
	Importance float64 `json:"importance"`
}

type EvaluationMetrics struct {
	MAE     float64 `json:"mae"`
	RMSE    float64 `json:"rmse"`
	MAPE    float64 `json:"mape"`
	R2Score float64 `json:"r2_score"`
}

type PredictionResultDTO struct {
	ID                uint                `json:"id,omitempty"`
	Symbol            string              `json:"symbol"`
	Currency          string              `json:"currency"`
	LastPrice         float64             `json:"last_price"`
	PriceChange       float64             `json:"price_change"`
	PctChange         float64             `json:"pct_change"`
	ModelName         string              `json:"model_name"`
	ModelType         string              `json:"model_type"`
	Metrics           EvaluationMetrics   `json:"metrics"`
	Predictions       []PredictionPoint   `json:"predictions"`
	History           []HistoryPoint      `json:"history"`
	FeatureImportance []FeatureImportance `json:"feature_importances"`
	GeneratedAt       string              `json:"generated_at"`
}
