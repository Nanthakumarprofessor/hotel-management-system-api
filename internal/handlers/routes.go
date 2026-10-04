package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"

	"hotel-updated/internal/config"
	"hotel-updated/internal/loggers"
	"hotel-updated/internal/repository"
	"hotel-updated/internal/services"
	"hotel-updated/pkg/database"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *mux.Router, db *database.Db, logger *loggers.Logger, cfg *config.Config) {
	// Initialize repositories
	roomRepo := repository.NewRoomRepository(db, logger)
	bookingRepo := repository.NewBookingRepository(db, logger)

	// Initialize services
	roomService := services.NewRoomService(roomRepo, logger)
	bookingService := services.NewBookingService(bookingRepo, roomRepo, logger)

	// Initialize handlers
	roomHandler := NewRoomHandler(roomService, logger)
	bookingHandler := NewBookingHandler(bookingService, logger)

	// router.HandleFunc("/docs",func(w http.ResponseWriter,r *http.Request){
	// 	http.ServeFile(w,r,"./docs/index.html")
	// }).Methods("GET")

	router.HandleFunc("/hotel-output/docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		openAPIPath := filepath.Join(cfg.ProjectRoot, "docs", "openapi.yaml")
		// Read the openAPI.yaml file
		yamlContent, err := os.ReadFile(openAPIPath)
		if err != nil {
			fmt.Println("Error reading openAPI.yaml:", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// Replace the placeholder with the actual base URL
		updatedContent := strings.ReplaceAll(string(yamlContent), "{BASE_URL}", cfg.BaseUrl)
		// Serve the updated OpenAPI documentation
		w.Header().Set("Content-Type", "application/x-yaml")

		if _, writeErr := w.Write([]byte(updatedContent)); writeErr != nil {
			logger.Info("Error writing response")
		}
	}).Methods(http.MethodGet)

	// Serve the Swagger UI files on the /docs path
	swaggerUIPath := http.Dir(filepath.Join(cfg.ProjectRoot, "docs", "swaggerui"))
	fs := http.FileServer(swaggerUIPath)
	router.PathPrefix("/hotel-output/docs/").Handler(http.StripPrefix("/hotel-output/docs/", fs))

	// Create API subrouter
	api := router.PathPrefix("/api/v1").Subrouter()

	// Room Management routes
	api.HandleFunc("/rooms/categories", roomHandler.GetRoomCategories).Methods("GET")
	api.HandleFunc("/rooms/available", roomHandler.GetAvailableRooms).Methods("GET")

	// Booking Management routes
	api.HandleFunc("/bookings", bookingHandler.CreateBooking).Methods("POST")
	api.HandleFunc("/bookings/extend", bookingHandler.ExtendBooking).Methods("PUT")
	api.HandleFunc("/bookings/cancel", bookingHandler.CancelBooking).Methods("PATCH")
	api.HandleFunc("/bookings/restore", bookingHandler.RestoreBooking).Methods("PATCH")

	logger.Info("Routes configured successfully")
}
