package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"stockwise-ai/config"
	"stockwise-ai/internal/model"
	"stockwise-ai/internal/repository"
)

type PredictionService interface {
	RunPrediction(userID uint, req *model.PredictionRequest) (*model.PredictionResultDTO, error)
	GetUserHistory(userID uint, limit int) ([]model.Prediction, error)
	GetPredictionByID(id uint) (*model.PredictionResultDTO, error)
}

type predictionService struct {
	predictionRepo repository.PredictionRepository
	stockRepo      repository.StockRepository
	cfg            *config.Config
}

func NewPredictionService(
	predRepo repository.PredictionRepository,
	stockRepo repository.StockRepository,
	cfg *config.Config,
) PredictionService {
	return &predictionService{
		predictionRepo: predRepo,
		stockRepo:      stockRepo,
		cfg:            cfg,
	}
}

func (s *predictionService) RunPrediction(userID uint, req *model.PredictionRequest) (*model.PredictionResultDTO, error) {
	if req.Days <= 0 {
		req.Days = 7
	}
	if req.Days > 30 {
		req.Days = 30
	}
	if req.Period == "" {
		req.Period = "1y"
	}
	req.Symbol = strings.ToUpper(strings.TrimSpace(req.Symbol))
	req.ModelType = strings.ToLower(strings.TrimSpace(req.ModelType))

	// Prepare Python CLI execution
	pythonBin := s.cfg.PythonBin
	if pythonBin == "" {
		pythonBin = "python"
	}

	args := []string{
		"-m", "ml.engine",
		"--symbol", req.Symbol,
		"--model", req.ModelType,
		"--days", strconv.Itoa(req.Days),
		"--period", req.Period,
	}

	log.Printf("Executing ML engine: %s %s", pythonBin, strings.Join(args, " "))

	cmd := exec.Command(pythonBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = stdout.String()
		}
		log.Printf("Python execution error: %v, detail: %s", err, errMsg)
		return nil, fmt.Errorf("ML model inference failed: %s", errMsg)
	}

	// Parse JSON from stdout
	var result model.PredictionResultDTO
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		log.Printf("Error unmarshaling Python output: %v, raw: %s", err, stdout.String())
		return nil, fmt.Errorf("invalid response format from ML engine: %w", err)
	}

	// Update stock price in DB
	stock, _ := s.stockRepo.FindBySymbol(result.Symbol)
	if stock == nil {
		stock = &model.Stock{
			Symbol:      result.Symbol,
			Name:        result.Symbol,
			Currency:    result.Currency,
			LastPrice:   result.LastPrice,
			PriceChange: result.PriceChange,
			PctChange:   result.PctChange,
		}
	} else {
		stock.LastPrice = result.LastPrice
		stock.PriceChange = result.PriceChange
		stock.PctChange = result.PctChange
		stock.Currency = result.Currency
	}
	_ = s.stockRepo.CreateOrUpdate(stock)

	// Serialize JSON payloads for storage
	predJson, _ := json.Marshal(result.Predictions)
	histJson, _ := json.Marshal(result.History)
	featJson, _ := json.Marshal(result.FeatureImportance)

	dbRecord := &model.Prediction{
		UserID:          userID,
		Symbol:          result.Symbol,
		ModelType:       result.ModelType,
		ModelName:       result.ModelName,
		TargetDays:      req.Days,
		LastPrice:       result.LastPrice,
		Currency:        result.Currency,
		MAE:             result.Metrics.MAE,
		RMSE:            result.Metrics.RMSE,
		MAPE:            result.Metrics.MAPE,
		R2Score:         result.Metrics.R2Score,
		PredictionsJSON: string(predJson),
		HistoryJSON:     string(histJson),
		FeaturesJSON:    string(featJson),
	}

	if err := s.predictionRepo.Create(dbRecord); err != nil {
		log.Printf("Warning: Failed to save prediction to database: %v", err)
	} else {
		result.ID = dbRecord.ID
	}

	return &result, nil
}

func (s *predictionService) GetUserHistory(userID uint, limit int) ([]model.Prediction, error) {
	return s.predictionRepo.GetByUserID(userID, limit)
}

func (s *predictionService) GetPredictionByID(id uint) (*model.PredictionResultDTO, error) {
	pred, err := s.predictionRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if pred == nil {
		return nil, errors.New("prediction not found")
	}

	var preds []model.PredictionPoint
	var hist []model.HistoryPoint
	var feats []model.FeatureImportance

	_ = json.Unmarshal([]byte(pred.PredictionsJSON), &preds)
	_ = json.Unmarshal([]byte(pred.HistoryJSON), &hist)
	_ = json.Unmarshal([]byte(pred.FeaturesJSON), &feats)

	return &model.PredictionResultDTO{
		ID:          pred.ID,
		Symbol:      pred.Symbol,
		Currency:    pred.Currency,
		LastPrice:   pred.LastPrice,
		ModelName:   pred.ModelName,
		ModelType:   pred.ModelType,
		Metrics: model.EvaluationMetrics{
			MAE:     pred.MAE,
			RMSE:    pred.RMSE,
			MAPE:    pred.MAPE,
			R2Score: pred.R2Score,
		},
		Predictions:       preds,
		History:           hist,
		FeatureImportance: feats,
		GeneratedAt:       pred.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
