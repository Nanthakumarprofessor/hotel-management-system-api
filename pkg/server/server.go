package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.uber.org/zap"

	"hotel-updated/internal/config"
	"hotel-updated/internal/handlers"
	"hotel-updated/internal/loggers"
	"hotel-updated/pkg/database"
)

// Application holds the application dependencies
type Application struct {
	Logger    *loggers.Logger
	Router    *mux.Router
	Config    *config.Config
	DBService *database.DBService
}

// InitializeApp initializes the application
func InitializeApp(env string) (*Application, error) {
	// Load configuration first (needed for log config)
	cfg, err := config.LoadConfig("hotel-management-system", env)
	if err != nil {
		return nil, err
	}

	// Initialize logger using config from env
	logger := loggers.NewLogger(loggers.LogConfig{
		Level:       cfg.LogLevel,
		LogDir:      cfg.LogDir,
		FileName:    cfg.LogFileName,
		ServiceName: "hotel-management-system",
	})
	logger.Info("Logger initialized")

	// Initialize database
	dbService := &database.DBService{}
	logger.Info("Initializing database connection")

	db, err := dbService.EstablishConnection(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to establish database connection", zap.Error(err))
		return nil, err
	}
	logger.Info("Database connection established successfully")

	// Run migrations
	if err := database.AutoMigrate(db, logger); err != nil {
		logger.Error("Failed to run migrations", zap.Error(err))
		return nil, err
	}

	// Create router
	router := mux.NewRouter()

	// Setup routes
	handlers.SetupRoutes(router, db, logger, cfg)

	return &Application{
		Logger:    logger,
		Router:    router,
		Config:    cfg,
		DBService: dbService,
	}, nil
}

// RunServer starts the HTTP server
func RunServer(app *Application) error {
	// Configure CORS
	corsOpts := cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}

	handler := cors.New(corsOpts).Handler(app.Router)

	app.Logger.Info("Server starting", zap.String("port", app.Config.Port))
	return http.ListenAndServe(":"+app.Config.Port, handler)
}
