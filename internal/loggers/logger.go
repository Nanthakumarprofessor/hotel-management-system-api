package loggers

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zap *zap.Logger
}

type LogConfig struct {
	Level       string
	LogDir      string
	FileName    string
	ServiceName string
}

func NewLogger(cfg LogConfig) *Logger {
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create log directory %s: %v", cfg.LogDir, err))
	}

	filePath := filepath.Join(cfg.LogDir, cfg.FileName)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to open log file %s: %v", filePath, err))
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:      "ts",
		LevelKey:     "level",
		NameKey:      "service",
		MessageKey:   "msg",
		CallerKey:    "caller",
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		EncodeTime:   zapcore.EpochTimeEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	level := parseLevel(cfg.Level)
	jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), level),
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(file), level),
	)

	return &Logger{zap: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Named(cfg.ServiceName)}
}

// Sync flushes buffered log entries
func (l *Logger) Sync() {
	_ = l.zap.Sync()
}

// Info logs a structured info message
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Error logs a structured error message with stacktrace
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Warn logs a structured warning message
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Fatal logs a structured fatal message and exits the program
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
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
