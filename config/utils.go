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
