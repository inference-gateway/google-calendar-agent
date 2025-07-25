package config

import (
	"context"
	"fmt"

	"github.com/inference-gateway/a2a/adk/server/config"
	"github.com/sethvargo/go-envconfig"
)

// Config represents the application configuration
type Config struct {
	// Environment specifies the deployment environment (dev, staging, prod)
	Environment string `env:"ENVIRONMENT, default=dev"`

	// DemoMode enables demo mode with mock services
	DemoMode bool `env:"DEMO_MODE, default=false"`

	// Google Calendar Configuration
	Google GoogleConfig `env:", prefix=GOOGLE_"`

	// Logging Configuration
	Logging LoggingConfig `env:", prefix=LOG_"`

	// A2A Configuration
	A2A config.Config `env:", prefix=A2A_"`
}

// GoogleConfig holds Google Calendar API related configuration
type GoogleConfig struct {
	// CalendarID is the target Google Calendar ID to use
	CalendarID string `env:"CALENDAR_ID, default=primary"`

	// ServiceAccountJSON contains the Google Service Account credentials in JSON format
	ServiceAccountJSON string `env:"CALENDAR_SA_JSON"`

	// CredentialsPath is the path to Google credentials file (alternative to ServiceAccountJSON)
	CredentialsPath string `env:"APPLICATION_CREDENTIALS"`

	// ReadOnly determines if the calendar should be accessed in read-only mode
	ReadOnly bool `env:"CALENDAR_READ_ONLY, default=false"`

	// TimeZone is the default timezone for interpreting user time inputs (e.g., "Europe/Berlin", "America/New_York")
	TimeZone string `env:"CALENDAR_TIMEZONE, default=UTC"`
}

// LoggingConfig holds logging related configuration
type LoggingConfig struct {
	// Level sets the log level (debug, info, warn, error)
	Level string `env:"LEVEL, default=info"`

	// Format sets the log format (json, console)
	Format string `env:"FORMAT, default=json"`

	// Output sets the log output destination (stdout, stderr, file path)
	Output string `env:"OUTPUT, default=stdout"`

	// EnableCaller adds caller information to log entries
	EnableCaller bool `env:"ENABLE_CALLER, default=true"`

	// EnableStacktrace adds stacktrace to error level logs
	EnableStacktrace bool `env:"ENABLE_STACKTRACE, default=true"`
}

// Load loads configuration from environment variables
func Load(ctx context.Context) (*Config, error) {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, fmt.Errorf("failed to process configuration: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// LoadWithLookuper loads configuration using a custom lookuper (useful for testing)
func LoadWithLookuper(ctx context.Context, lookuper envconfig.Lookuper) (*Config, error) {
	var cfg Config
	if err := envconfig.ProcessWith(ctx, &envconfig.Config{
		Target:   &cfg,
		Lookuper: lookuper,
	}); err != nil {
		return nil, fmt.Errorf("failed to process configuration: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	if !c.DemoMode {
		if c.Google.ServiceAccountJSON == "" && c.Google.CredentialsPath == "" {
			return fmt.Errorf("either GOOGLE_CALENDAR_SA_JSON or GOOGLE_APPLICATION_CREDENTIALS must be provided when not in demo mode")
		}
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level '%s', must be one of: debug, info, warn, error", c.Logging.Level)
	}

	return nil
}

// GetServerAddress returns the formatted server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf(":%s", c.A2A.ServerConfig.Port)
}

// IsProduction returns true if the application is running in production
func (c *Config) IsProduction() bool {
	return c.Environment == "prod" || c.Environment == "production"
}

// IsDevelopment returns true if the application is running in development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "dev" || c.Environment == "development"
}

// IsDebugEnabled returns true if debug mode is enabled
func (c *Config) IsDebugEnabled() bool {
	return c.Logging.Level == "debug" || c.IsDevelopment()
}

// ShouldUseMockService returns true if mock services should be used
func (c *Config) ShouldUseMockService() bool {
	return c.DemoMode
}
