package config_test

import (
	"os"
	"testing"

	"hotel-updated/internal/config"

	"github.com/stretchr/testify/assert"
)

func setEnv(t *testing.T, pairs map[string]string) {
	t.Helper()
	for k, v := range pairs {
		os.Setenv(k, v)
	}
	t.Cleanup(func() {
		for k := range pairs {
			os.Unsetenv(k)
		}
	})
}

func TestLoadConfig_AllEnvSet(t *testing.T) {
	setEnv(t, map[string]string{
		"DATABASE_URL":  "postgres://test:test@localhost/testdb",
		"PORT":          "9090",
		"PROJECT_ROOT":  "/test/root",
		"BASE_URL":      "http://localhost:9090",
		"LOG_LEVEL":     "debug",
		"LOG_DIR":       "/tmp/logs",
		"LOG_FILE_NAME": "test.log",
	})

	cfg, err := config.LoadConfig("hotel-service", "development")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "postgres://test:test@localhost/testdb", cfg.DatabaseURL)
	assert.Equal(t, "/test/root", cfg.ProjectRoot)
	assert.Equal(t, "http://localhost:9090", cfg.BaseUrl)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "/tmp/logs", cfg.LogDir)
	assert.Equal(t, "test.log", cfg.LogFileName)
}

func TestLoadConfig_Defaults(t *testing.T) {
	setEnv(t, map[string]string{"DATABASE_URL": "postgres://test:test@localhost/testdb"})
	os.Unsetenv("PORT")
	os.Unsetenv("PROJECT_ROOT")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LOG_DIR")
	os.Unsetenv("LOG_FILE_NAME")

	cfg, err := config.LoadConfig("hotel-service", "development")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, ".", cfg.ProjectRoot)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "./logs", cfg.LogDir)
	assert.Equal(t, "hotel-management-system.log", cfg.LogFileName)
}

func TestLoadConfig_MissingDatabaseURL(t *testing.T) {
	os.Unsetenv("DATABASE_URL")

	cfg, err := config.LoadConfig("hotel-service", "development")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}

func TestLoadConfig_EnvFileLoadError(t *testing.T) {
    // ✅ Create a directory (not a file)
    os.MkdirAll("envs", 0755)

    // ✅ Create a DIRECTORY instead of file
    os.Mkdir("envs/.env.invalid", 0755)

    setEnv(t, map[string]string{
        "DATABASE_URL": "postgres://test:test@localhost/testdb",
    })

    cfg, err := config.LoadConfig("hotel-service", "invalid")

    assert.Nil(t, cfg)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to load env file")
}
