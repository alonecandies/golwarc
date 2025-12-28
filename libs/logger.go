package libs

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	Level       string // debug, info, warn, error
	Development bool
	OutputPaths []string
}

// InitLogger initializes the global logger with the provided configuration
func InitLogger(config LoggerConfig) error {
	var zapConfig zap.Config

	if config.Development {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// Set log level
	level := zapcore.InfoLevel
	switch config.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Set output paths
	if len(config.OutputPaths) > 0 {
		zapConfig.OutputPaths = config.OutputPaths
	}

	// Build logger
	logger, err := zapConfig.Build()
	if err != nil {
		return err
	}

	Logger = logger
	return nil
}

// InitDefaultLogger initializes the logger with default settings
func InitDefaultLogger() error {
	return InitLogger(LoggerConfig{
		Level:       "info",
		Development: true,
		OutputPaths: []string{"stdout"},
	})
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if Logger == nil {
		// Initialize with default settings if not already initialized
		_ = InitDefaultLogger()
	}
	return Logger
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// SetLogLevel dynamically changes the log level
func SetLogLevel(level string) {
	if Logger == nil {
		return
	}

	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	Logger = Logger.WithOptions(zap.IncreaseLevel(zapLevel))
}

// Fatal logs a message at fatal level then exits
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
	os.Exit(1)
}
