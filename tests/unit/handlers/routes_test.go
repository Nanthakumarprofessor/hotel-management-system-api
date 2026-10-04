package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"hotel-updated/internal/config"
	"hotel-updated/internal/handlers"
	"hotel-updated/internal/loggers"
	"hotel-updated/pkg/database"
)

// setupMockPostgresDB creates a *database.Db backed by go-sqlmock (postgres dialect).
// No real Postgres connection is needed — all SQL is intercepted by the mock.
func setupMockPostgresDB(t *testing.T) (*database.Db, sqlmock.Sqlmock, func()) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm with sqlmock: %v", err)
	}
	return &database.Db{Gorm: gormDB}, mock, func() { sqlDB.Close() }
}

// newRoutesLogger returns a silent logger that writes to a temp directory,
// ensuring os.MkdirAll never receives an empty path (which panics on Windows).
func newRoutesLogger() *loggers.Logger {
	tmpDir, _ := os.MkdirTemp("", "routes-test-logs-*")
	return loggers.NewLogger(loggers.LogConfig{
		Level:       "error",
		LogDir:      tmpDir,
		FileName:    "routes-test.log",
		ServiceName: "routes-test",
	})
}

// newRoutesCfg builds a temporary project root with the minimal files that
// SetupRoutes needs (docs/openapi.yaml and docs/swaggerui/index.html).
func newRoutesCfg(t *testing.T) (*config.Config, string) {
	t.Helper()
	tmpDir := t.TempDir()
	swaggerDir := filepath.Join(tmpDir, "docs", "swaggerui")
	_ = os.MkdirAll(swaggerDir, os.ModePerm)

	_ = os.WriteFile(
		filepath.Join(tmpDir, "docs", "openapi.yaml"),
		[]byte("openapi: 3.0.0\nservers:\n  - url: {BASE_URL}"),
		0644,
	)
	_ = os.WriteFile(filepath.Join(swaggerDir, "index.html"), []byte("<html>swagger ui</html>"), 0644)

	cfg := &config.Config{
		BaseUrl:     "http://localhost:8080",
		ProjectRoot: tmpDir,
	}
	return cfg, tmpDir
}

// setupRouter builds a fully-wired mux.Router using sqlmock-backed postgres DB.
func setupRouter(t *testing.T) (*mux.Router, *config.Config) {
	t.Helper()
	cfg, _ := newRoutesCfg(t)
	db, _, cleanup := setupMockPostgresDB(t)
	t.Cleanup(cleanup)

	router := mux.NewRouter()
	handlers.SetupRoutes(router, db, newRoutesLogger(), cfg)
	return router, cfg
}

// newTestLogger is an alias kept so existing files in this package compile.
func newTestLogger() *loggers.Logger {
	return newRoutesLogger()
}

// ─── OpenAPI YAML endpoint ────────────────────────────────────────────────────

func TestSetupRoutes_OpenAPIYAML_Found(t *testing.T) {
	router, cfg := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/hotel-output/docs/openapi.yaml", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/x-yaml", rr.Header().Get("Content-Type"))
	assert.Contains(t, rr.Body.String(), cfg.BaseUrl)
}

func TestSetupRoutes_OpenAPIYAML_NotFound(t *testing.T) {
	db, _, cleanup := setupMockPostgresDB(t)
	t.Cleanup(cleanup)

	cfg := &config.Config{
		BaseUrl:     "http://localhost:8080",
		ProjectRoot: "/nonexistent-project-root-abc123",
	}
	router := mux.NewRouter()
	handlers.SetupRoutes(router, db, newRoutesLogger(), cfg)

	req := httptest.NewRequest(http.MethodGet, "/hotel-output/docs/openapi.yaml", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// failWriter lets us trigger the w.Write error branch inside the YAML handler.
type failWriter struct{ header http.Header }

func (f *failWriter) Header() http.Header      { return f.header }
func (f *failWriter) Write([]byte) (int, error) { return 0, os.ErrClosed }
func (f *failWriter) WriteHeader(int)            {}

func TestSetupRoutes_OpenAPIYAML_WriteError(t *testing.T) {
	router, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/hotel-output/docs/openapi.yaml", nil)
	router.ServeHTTP(&failWriter{header: http.Header{}}, req)
	// No panic and no assertions needed — we just exercise the writeErr branch.
}

// ─── Swagger UI static files ──────────────────────────────────────────────────

func TestSetupRoutes_SwaggerUI_ExistingFile(t *testing.T) {
	router, _ := setupRouter(t)

	// The Go file server serves the directory index at the trailing-slash URL.
	// Requesting /hotel-output/docs/ directly returns 200 with index.html content.
	req := httptest.NewRequest(http.MethodGet, "/hotel-output/docs/", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestSetupRoutes_SwaggerUI_MissingFile(t *testing.T) {
	router, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/hotel-output/docs/missing-file.js", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ─── Booking routes ───────────────────────────────────────────────────────────

func TestSetupRoutes_CreateBooking_InvalidJSON(t *testing.T) {
	router, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBufferString(`{bad}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSetupRoutes_ExtendBooking_InvalidJSON(t *testing.T) {
	router, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/bookings/extend", bytes.NewBufferString(`{bad}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSetupRoutes_CancelBooking_InvalidJSON(t *testing.T) {
	router, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/cancel", bytes.NewBufferString(`{bad}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSetupRoutes_RestoreBooking_InvalidJSON(t *testing.T) {
	router, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/restore", bytes.NewBufferString(`{bad}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ─── Room routes ──────────────────────────────────────────────────────────────

func TestSetupRoutes_GetRoomCategories_Registered(t *testing.T) {
	router, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/categories", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Route must be registered (not 404). The actual DB call will fail gracefully
	// because no SQL expectations are set — we only verify routing here.
	assert.NotEqual(t, http.StatusNotFound, rr.Code)
}

func TestSetupRoutes_GetAvailableRooms_InvalidParams(t *testing.T) {
	router, _ := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available?check_in=bad-date", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
