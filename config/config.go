package config

import (
	"context"
	"fmt"
	"time"

	"github.com/sethvargo/go-envconfig"
)

// Config represents the application configuration
type Config struct {
	// Google Calendar Configuration
	Google GoogleConfig `env:", prefix=GOOGLE_"`

	// Server Configuration
	Server ServerConfig `env:", prefix=SERVER_"`

	// Logging Configuration
	Logging LoggingConfig `env:", prefix=LOG_"`

	// TLS Configuration
	TLS TLSConfig `env:", prefix=TLS_"`

	// Application Configuration
	App AppConfig `env:", prefix=APP_"`

	// LLM Configuration
	LLM LLMConfig `env:", prefix=LLM_"`
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

// ServerConfig holds HTTP server related configuration
type ServerConfig struct {
	// Port is the port the server will listen on
	Port string `env:"PORT, default=8080"`

	// Host is the host the server will bind to
	Host string `env:"HOST, default=0.0.0.0"`

	// Mode sets the Gin server mode (debug, release, test)
	Mode string `env:"GIN_MODE, default=release"`

	// EnableTLS determines if HTTPS should be enabled
	EnableTLS bool `env:"ENABLE_TLS, default=false"`

	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration `env:"READ_TIMEOUT, default=10s"`

	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT, default=10s"`

	// IdleTimeout is the maximum amount of time to wait for the next request
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT, default=60s"`
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

// TLSConfig holds TLS/HTTPS related configuration
type TLSConfig struct {
	// CertPath is the path to the TLS certificate file
	CertPath string `env:"CERT_PATH"`

	// KeyPath is the path to the TLS private key file
	KeyPath string `env:"KEY_PATH"`

	// MinVersion sets the minimum TLS version (1.2, 1.3)
	MinVersion string `env:"MIN_VERSION, default=1.2"`

	// CipherSuites is a comma-separated list of cipher suites
	CipherSuites string `env:"CIPHER_SUITES"`
}

// AppConfig holds general application configuration
type AppConfig struct {
	// Environment specifies the deployment environment (dev, staging, prod)
	Environment string `env:"ENVIRONMENT, default=dev"`

	// DemoMode enables demo mode with mock services
	DemoMode bool `env:"DEMO_MODE, default=false"`

	// MaxRequestSize sets the maximum request body size in bytes
	MaxRequestSize int64 `env:"MAX_REQUEST_SIZE, default=1048576"` // 1MB

	// RequestTimeout sets the maximum duration for handling requests
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT, default=30s"`
}

// LLMConfig holds LLM provider configuration for natural language processing
// Supports both Inference Gateway and OpenAI-compatible API endpoints
type LLMConfig struct {
	// GatewayURL is the URL of the Inference Gateway or OpenAI-compatible API endpoint
	GatewayURL string `env:"GATEWAY_URL, default=http://localhost:8080/v1"`

	// Provider is the LLM provider to use through the Inference Gateway
	// Supported providers: openai, anthropic, groq, ollama, deepseek, cohere, cloudflare
	Provider string `env:"PROVIDER, default=groq"`

	// Model is the specific model to use (e.g., gpt-4o, claude-3-opus, deepseek-r1-distill-llama-70b)
	Model string `env:"MODEL, default=deepseek-r1-distill-llama-70b"`

	// Timeout is the timeout for LLM requests
	Timeout time.Duration `env:"TIMEOUT, default=30s"`

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int `env:"MAX_TOKENS, default=2048"`

	// Temperature controls randomness in generation (0.0 to 2.0)
	Temperature float64 `env:"TEMPERATURE, default=0.7"`

	// Enabled determines if LLM functionality is enabled
	Enabled bool `env:"ENABLED, default=true"`
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
	if !c.App.DemoMode {
		if c.Google.ServiceAccountJSON == "" && c.Google.CredentialsPath == "" {
			return fmt.Errorf("either GOOGLE_CALENDAR_SA_JSON or GOOGLE_APPLICATION_CREDENTIALS must be provided when not in demo mode")
		}
	}

	if c.Server.EnableTLS {
		if c.TLS.CertPath == "" {
			return fmt.Errorf("TLS_CERT_PATH is required when TLS is enabled")
		}
		if c.TLS.KeyPath == "" {
			return fmt.Errorf("TLS_KEY_PATH is required when TLS is enabled")
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

	validModes := map[string]bool{
		"debug":   true,
		"release": true,
		"test":    true,
	}
	if !validModes[c.Server.Mode] {
		return fmt.Errorf("invalid server mode '%s', must be one of: debug, release, test", c.Server.Mode)
	}

	if c.Server.EnableTLS {
		validTLSVersions := map[string]bool{
			"1.2": true,
			"1.3": true,
		}
		if !validTLSVersions[c.TLS.MinVersion] {
			return fmt.Errorf("invalid TLS version '%s', must be one of: 1.2, 1.3", c.TLS.MinVersion)
		}
	}

	if c.LLM.Enabled {
		if c.LLM.GatewayURL == "" {
			return fmt.Errorf("LLM_GATEWAY_URL is required when LLM is enabled")
		}
		if c.LLM.Provider == "" {
			return fmt.Errorf("LLM_PROVIDER is required when LLM is enabled")
		}

		validProviders := map[string]bool{
			"openai":     true,
			"anthropic":  true,
			"groq":       true,
			"ollama":     true,
			"deepseek":   true,
			"cohere":     true,
			"cloudflare": true,
		}
		if !validProviders[c.LLM.Provider] {
			return fmt.Errorf("invalid LLM provider '%s', must be one of: openai, anthropic, groq, ollama, deepseek, cohere, cloudflare", c.LLM.Provider)
		}

		if c.LLM.Model == "" {
			return fmt.Errorf("LLM_MODEL is required when LLM is enabled")
		}
		if c.LLM.Temperature < 0.0 || c.LLM.Temperature > 2.0 {
			return fmt.Errorf("LLM_TEMPERATURE must be between 0.0 and 2.0, got %f", c.LLM.Temperature)
		}
		if c.LLM.MaxTokens <= 0 {
			return fmt.Errorf("LLM_MAX_TOKENS must be greater than 0, got %d", c.LLM.MaxTokens)
		}
	}

	return nil
}

// GetServerAddress returns the formatted server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// IsProduction returns true if the application is running in production
func (c *Config) IsProduction() bool {
	return c.App.Environment == "prod" || c.App.Environment == "production"
}

// IsDevelopment returns true if the application is running in development
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "dev" || c.App.Environment == "development"
}

// IsDebugEnabled returns true if debug mode is enabled
func (c *Config) IsDebugEnabled() bool {
	return c.Logging.Level == "debug" || c.IsDevelopment()
}

// ShouldUseMockService returns true if mock services should be used
func (c *Config) ShouldUseMockService() bool {
	return c.App.DemoMode
}
