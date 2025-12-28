package libs_test

import (
	"testing"

	"github.com/alonecandies/golwarc/libs"
)

// =====================
// Logger Unit Tests
// =====================

func TestInitDefaultLogger(t *testing.T) {
	err := libs.InitDefaultLogger()
	if err != nil {
		t.Fatalf("Failed to initialize default logger: %v", err)
	}
	defer libs.Sync()
}

func TestGetLogger(t *testing.T) {
	libs.InitDefaultLogger()
	defer libs.Sync()

	logger := libs.GetLogger()
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}
}

func TestLoggerLevels(t *testing.T) {
	libs.InitDefaultLogger()
	defer libs.Sync()

	logger := libs.GetLogger()

	// These should not panic
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
}

func TestInitLoggerWithConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  libs.LoggerConfig
		wantErr bool
	}{
		{
			name: "debug level",
			config: libs.LoggerConfig{
				Level:       "debug",
				Development: true,
				OutputPaths: []string{"stdout"},
			},
			wantErr: false,
		},
		{
			name: "info level",
			config: libs.LoggerConfig{
				Level:       "info",
				Development: true,
				OutputPaths: []string{"stdout"},
			},
			wantErr: false,
		},
		{
			name: "warn level",
			config: libs.LoggerConfig{
				Level:       "warn",
				Development: false,
				OutputPaths: []string{"stdout"},
			},
			wantErr: false,
		},
		{
			name: "error level",
			config: libs.LoggerConfig{
				Level:       "error",
				Development: false,
				OutputPaths: []string{"stdout"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := libs.InitLogger(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitLogger() error = %v, wantErr %v", err, tt.wantErr)
			}
			libs.Sync()
		})
	}
}

func TestSetLogLevel(t *testing.T) {
	libs.InitDefaultLogger()
	defer libs.Sync()

	// Test changing log levels - should not panic
	libs.SetLogLevel("debug")
	libs.SetLogLevel("info")
	libs.SetLogLevel("warn")
	libs.SetLogLevel("error")
}

func TestGetLoggerAutoInit(t *testing.T) {
	// Reset logger
	libs.Logger = nil

	// GetLogger should auto-initialize
	logger := libs.GetLogger()
	if logger == nil {
		t.Fatal("Logger should auto-initialize")
	}
	libs.Sync()
}
