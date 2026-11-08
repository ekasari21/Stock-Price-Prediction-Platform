package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"stockwise-ai/config"
	"stockwise-ai/internal/handler"
	"stockwise-ai/internal/middleware"
	"stockwise-ai/internal/repository"
	"stockwise-ai/internal/service"
)

func main() {
	log.Println("==================================================")
	log.Println("           Starting Stockwise AI Server           ")
	log.Println("==================================================")

	// 1. Load Application Configuration
	cfg := config.LoadConfig()

	// 2. Initialize Database (GORM + PostgreSQL / SQLite fallback)
	db, err := repository.InitDB(cfg)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}

	// 3. Initialize Repositories (Data Access Layer)
	userRepo := repository.NewUserRepository(db)
	stockRepo := repository.NewStockRepository(db)
	predictionRepo := repository.NewPredictionRepository(db)

	// 4. Initialize Services (Business Logic & ML Integration Layer)
	authService := service.NewAuthService(userRepo, cfg)
	stockService := service.NewStockService(stockRepo)
	predictionService := service.NewPredictionService(predictionRepo, stockRepo, cfg)

	// 5. Initialize Handlers (Controller / Presentation Layer)
	authHandler := handler.NewAuthHandler(authService)
	stockHandler := handler.NewStockHandler(stockService)
	predictionHandler := handler.NewPredictionHandler(predictionService)

	// 6. Setup Gin Router
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	// Static files and HTML templates
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	// Web Dashboard Route
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Stockwise AI",
		})
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0",
		})
	})

	// RESTful API v1 Routes
	v1 := router.Group("/api/v1")
	{
		// Auth Routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)
			authGroup.GET("/me", middleware.JWTAuthMiddleware(authService), authHandler.GetCurrentUser)
		}

		// Stock Routes
		stockGroup := v1.Group("/stocks")
		{
			stockGroup.GET("", stockHandler.GetAllStocks)
			stockGroup.GET("/:symbol", stockHandler.GetStockBySymbol)
		}

		// Prediction Routes
		predGroup := v1.Group("/predictions")
		{
			// Forecast accepts optional auth: if authenticated, attaches to user profile
			predGroup.POST("/forecast", func(c *gin.Context) {
				authHeader := c.GetHeader("Authorization")
				if authHeader != "" {
					middleware.JWTAuthMiddleware(authService)(c)
					if c.IsAborted() {
						return
					}
				}
				predictionHandler.Forecast(c)
			})

			// Protected prediction history
			protectedPred := predGroup.Group("")
			protectedPred.Use(middleware.JWTAuthMiddleware(authService))
			{
				protectedPred.GET("/history", predictionHandler.GetHistory)
				protectedPred.GET("/:id", predictionHandler.GetByID)
			}
		}
	}

	// 7. Start HTTP Server using standard net/http and Gin
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Stockwise AI listening on http://localhost%s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Stockwise AI server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting successfully.")
}
