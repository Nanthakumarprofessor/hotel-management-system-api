package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Port        string
	DatabaseURL string
	ProjectRoot string
	BaseUrl     string
	LogLevel    string
	LogDir      string
	LogFileName string
}

// LoadConfig loads configuration from environment files
func LoadConfig(serviceName, env string) (*Config, error) {
	// Construct env file path

	envFile := fmt.Sprintf("envs/.env.%s", env)

	// Load environment variables from file
	err := godotenv.Load(envFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Warning: env file %s not present, continuing without it\n", envFile)
		} else {
			return nil, fmt.Errorf("failed to load env file: %w", err)
		}
	}

	// Read PORT with default 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Read DATABASE_URL (required)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	// Read PROJECT_ROOT with default current directory
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		projectRoot = "."
	}

	// Read BASE_URL
	baseUrl := os.Getenv("BASE_URL")

	// Read log settings with defaults
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "./logs"
	}

	logFileName := os.Getenv("LOG_FILE_NAME")
	if logFileName == "" {
		logFileName = "hotel-management-system.log"
	}

	return &Config{
		Port:        port,
		DatabaseURL: databaseURL,
		ProjectRoot: projectRoot,
		BaseUrl:     baseUrl,
		LogLevel:    logLevel,
		LogDir:      logDir,
		LogFileName: logFileName,
	}, nil
}
