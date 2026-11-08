package repository

import (
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"stockwise-ai/config"
	"stockwise-ai/internal/model"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	if cfg.DBDriver == "postgres" {
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
		)
		log.Printf("Connecting to PostgreSQL database at %s:%s/%s...", cfg.DBHost, cfg.DBPort, cfg.DBName)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			log.Printf("Warning: Failed to connect to PostgreSQL (%v). Falling back to local SQLite database (stockwise.db)", err)
			db, err = gorm.Open(sqlite.Open("stockwise.db"), gormConfig)
		} else {
			log.Println("Connected successfully to PostgreSQL database!")
		}
	} else {
		log.Println("Using SQLite database (stockwise.db)...")
		db, err = gorm.Open(sqlite.Open("stockwise.db"), gormConfig)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Automated database migrations using GORM
	log.Println("Running automated GORM database migrations...")
	err = db.AutoMigrate(
		&model.User{},
		&model.Stock{},
		&model.Prediction{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database tables: %w", err)
	}

	// Seed default stocks
	seedDefaultStocks(db)

	return db, nil
}

func seedDefaultStocks(db *gorm.DB) {
	var count int64
	db.Model(&model.Stock{}).Count(&count)
	if count == 0 {
		stocks := []model.Stock{
			{Symbol: "AAPL", Name: "Apple Inc.", Sector: "Technology", Exchange: "NASDAQ", Currency: "USD", LastPrice: 194.20, PriceChange: 1.15, PctChange: 0.60},
			{Symbol: "MSFT", Name: "Microsoft Corporation", Sector: "Technology", Exchange: "NASDAQ", Currency: "USD", LastPrice: 415.50, PriceChange: 2.80, PctChange: 0.68},
			{Symbol: "GOOGL", Name: "Alphabet Inc. (Google)", Sector: "Technology", Exchange: "NASDAQ", Currency: "USD", LastPrice: 176.30, PriceChange: -0.45, PctChange: -0.25},
			{Symbol: "NVDA", Name: "NVIDIA Corporation", Sector: "Semiconductors", Exchange: "NASDAQ", Currency: "USD", LastPrice: 128.90, PriceChange: 3.40, PctChange: 2.71},
			{Symbol: "TSLA", Name: "Tesla, Inc.", Sector: "Automotive", Exchange: "NASDAQ", Currency: "USD", LastPrice: 218.40, PriceChange: -1.20, PctChange: -0.55},
			{Symbol: "BBCA.JK", Name: "Bank Central Asia Tbk", Sector: "Financials", Exchange: "IDX", Currency: "IDR", LastPrice: 10250.00, PriceChange: 75.00, PctChange: 0.74},
			{Symbol: "BBRI.JK", Name: "Bank Rakyat Indonesia Tbk", Sector: "Financials", Exchange: "IDX", Currency: "IDR", LastPrice: 4820.00, PriceChange: 20.00, PctChange: 0.42},
		}
		for _, s := range stocks {
			db.Create(&s)
		}
		log.Printf("Seeded %d default stock symbols", len(stocks))
	}
}
