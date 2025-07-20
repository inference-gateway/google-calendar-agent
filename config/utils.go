package config

import (
	"fmt"
	"strings"
)

// GetGoogleCredentialsOption returns the appropriate Google API credential option
func (c *Config) GetGoogleCredentialsOption() (string, string, error) {
	if c.ShouldUseMockService() {
		return "", "", nil
	}

	if c.Google.ServiceAccountJSON != "" {
		return "json", c.Google.ServiceAccountJSON, nil
	}

	if c.Google.CredentialsPath != "" {
		return "file", c.Google.CredentialsPath, nil
	}

	return "", "", fmt.Errorf("no google credentials configured")
}

// GetLogLevel returns the zap log level equivalent
func (c *Config) GetLogLevel() string {
	// Convert our log level to match what the application expects
	switch strings.ToLower(c.Logging.Level) {
	case "debug":
		return "debug"
	case "info":
		return "info"
	case "warn", "warning":
		return "warn"
	case "error":
		return "error"
	default:
		return "info"
	}
}

// GetPort returns the port (TLS port adjustment is handled by A2A ADK)
func (c *Config) GetPort() string {
	return c.Server.Port
}

// GetProtocol returns the protocol scheme (TLS is handled by A2A ADK)
func (c *Config) GetProtocol() string {
	if c.Server.EnableTLS {
		return "https"
	}
	return "http"
}

// GetBaseURL returns the complete base URL for the server
func (c *Config) GetBaseURL() string {
	protocol := c.GetProtocol()
	port := c.GetPort()

	// For local development, include port
	return fmt.Sprintf("%s://localhost:%s", protocol, port)
}

// ToMap converts the config to a map for debugging/logging purposes
func (c *Config) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"google": map[string]interface{}{
			"calendar_id":     c.Google.CalendarID,
			"read_only":       c.Google.ReadOnly,
			"has_credentials": c.Google.ServiceAccountJSON != "" || c.Google.CredentialsPath != "",
		},
		"server": map[string]interface{}{
			"port":       c.Server.Port,
			"enable_tls": c.Server.EnableTLS,
		},
		"logging": map[string]interface{}{
			"level":             c.Logging.Level,
			"format":            c.Logging.Format,
			"output":            c.Logging.Output,
			"enable_caller":     c.Logging.EnableCaller,
			"enable_stacktrace": c.Logging.EnableStacktrace,
		},
		"app": map[string]interface{}{
			"environment":      c.App.Environment,
			"debug":            c.IsDebugEnabled(),
			"demo_mode":        c.App.DemoMode,
			"max_request_size": c.App.MaxRequestSize,
			"request_timeout":  c.App.RequestTimeout.String(),
		},
		"llm": map[string]interface{}{
			"enabled":     c.LLM.Enabled,
			"provider":    c.LLM.Provider,
			"model":       c.LLM.Model,
			"gateway_url": c.LLM.GatewayURL,
		},
	}
}
