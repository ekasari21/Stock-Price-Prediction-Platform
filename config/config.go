package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort                   string
	DBDriver                     string
	DBHost                       string
	DBPort                       string
	DBUser                       string
	DBPassword                   string
	DBName                       string
	DBSSLMode                    string
	JWTSecret                    string
	JWTExpirationHours           int
	RefreshTokenExpirationHours int
	PythonBin                    string
}

func LoadConfig() *Config {
	// Load .env file if available
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading, reading from environment variables")
	}

	jwtExp, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		jwtExp = 24
	}

	refreshExp, err := strconv.Atoi(getEnv("REFRESH_TOKEN_EXPIRATION_HOURS", "168"))
	if err != nil {
		refreshExp = 168
	}

	return &Config{
		ServerPort:                   getEnv("SERVER_PORT", "8080"),
		DBDriver:                     getEnv("DB_DRIVER", "postgres"),
		DBHost:                       getEnv("DB_HOST", "localhost"),
		DBPort:                       getEnv("DB_PORT", "5432"),
		DBUser:                       getEnv("DB_USER", "postgres"),
		DBPassword:                   getEnv("DB_PASSWORD", "postgres"),
		DBName:                       getEnv("DB_NAME", "stockwise_db"),
		DBSSLMode:                    getEnv("DB_SSLMODE", "disable"),
		JWTSecret:                    getEnv("JWT_SECRET", "stockwise_ai_secure_jwt_secret_key_2026"),
		JWTExpirationHours:           jwtExp,
		RefreshTokenExpirationHours: refreshExp,
		PythonBin:                    getEnv("PYTHON_BIN", "python"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaultVal
}
