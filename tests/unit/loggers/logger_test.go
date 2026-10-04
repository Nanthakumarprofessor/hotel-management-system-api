package loggers

import (
	"hotel-updated/internal/loggers"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ✅ 1. Covers NewLogger + Info + Warn + Error + Sync
func TestNewLoggerAndBasicMethods(t *testing.T) {
	// ❗ Use current working directory instead of TempDir
	dir := "."

	cfg := loggers.LogConfig{
		Level:       "info",
		LogDir:      dir,
		FileName:    "test.log",
		ServiceName: "test-service",
	}

	logger := loggers.NewLogger(cfg)

	logger.Info("info message", zap.String("key", "value"))
	logger.Warn("warn message")
	logger.Error("error message")

	logger.Sync()

	logFile := filepath.Join(dir, "test.log")

	if _, err := os.Stat(logFile); err != nil {
		t.Fatalf("log file not created: %v", err)
	}

	// ✅ Manually delete file (safe after test)
	_ = os.Remove(logFile)
}

// ✅ 3. Covers Fatal (os.Exit) using subprocess
func TestFatal_ShouldExit(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		cfg := loggers.LogConfig{
			Level:       "info",
			LogDir:      os.TempDir(),
			FileName:    "fatal.log",
			ServiceName: "fatal-test",
		}

		logger := loggers.NewLogger(cfg)
		logger.Fatal("fatal test") // exits process
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFatal_ShouldExit")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected process to exit, but it didn't")
	}
}

func TestNewLogger_DebugLevel(t *testing.T) {
    cfg := loggers.LogConfig{
        Level:       "debug",  // ✅ THIS triggers debug branch
        LogDir:      ".",
        FileName:    "debug.log",
        ServiceName: "debug-test",
    }

    logger := loggers.NewLogger(cfg)

    logger.Info("debug level test")
    logger.Sync()

    _ = os.Remove("debug.log")
}

func TestNewLogger_WarnLevel(t *testing.T) {
    cfg := loggers.LogConfig{
        Level:       "warn", // ✅ triggers WARN case
        LogDir:      ".",
        FileName:    "warn.log",
        ServiceName: "test-service",
    }

    logger := loggers.NewLogger(cfg)
    logger.Warn("warn level test")
    logger.Sync()

    _ = os.Remove("warn.log")
}

func TestNewLogger_ErrorLevel(t *testing.T) {
    cfg := loggers.LogConfig{
        Level:       "error", // ✅ triggers ERROR case
        LogDir:      ".",
        FileName:    "error.log",
        ServiceName: "test-service",
    }

    logger := loggers.NewLogger(cfg)
    logger.Error("error level test")
    logger.Sync()

    _ = os.Remove("error.log")
}

func TestNewLogger_DefaultLevel(t *testing.T) {
    cfg := loggers.LogConfig{
        Level:       "invalid", // ✅ triggers DEFAULT case
        LogDir:      ".",
        FileName:    "default.log",
        ServiceName: "test-service",
    }

    logger := loggers.NewLogger(cfg)
    logger.Info("default level test")
    logger.Sync()

    _ = os.Remove("default.log")
}

// ✅ 4. Covers panic branch: MkdirAll failure
func TestNewLogger_MkdirFailure(t *testing.T) {
	cfg := loggers.LogConfig{
		Level:       "info",
		LogDir:      string([]byte{0}), // invalid dir
		FileName:    "test.log",
		ServiceName: "test-service",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for MkdirAll failure")
		}
	}()

	loggers.NewLogger(cfg)
}

// ✅ 5. Covers panic branch: OpenFile failure
func TestNewLogger_OpenFileFailure(t *testing.T) {
	dir := t.TempDir()

	cfg := loggers.LogConfig{
		Level:       "info",
		LogDir:      dir,
		FileName:    string([]byte{0}), // invalid filename
		ServiceName: "test-service",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for OpenFile failure")
		}
	}()

	loggers.NewLogger(cfg)
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
