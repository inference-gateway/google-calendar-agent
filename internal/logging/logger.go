package logging

import (
	"fmt"

	config "github.com/inference-gateway/google-calendar-agent/config"
	zap "go.uber.org/zap"
)

// NewLogger creates a new logger based on the logging configuration
func NewLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	var zapConfig zap.Config

	// Set up config based on environment
	if cfg.Format == "console" {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// Set log level
	level, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level '%s': %w", cfg.Level, err)
	}
	zapConfig.Level = level

	// Configure caller information
	zapConfig.DisableCaller = !cfg.EnableCaller
	zapConfig.DisableStacktrace = !cfg.EnableStacktrace

	// Set output paths based on configuration
	switch cfg.Output {
	case "stdout":
		zapConfig.OutputPaths = []string{"stdout"}
	case "stderr":
		zapConfig.OutputPaths = []string{"stderr"}
	case "":
		zapConfig.OutputPaths = []string{"stdout"}
	default:
		// Assume it's a file path
		zapConfig.OutputPaths = []string{cfg.Output}
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}
