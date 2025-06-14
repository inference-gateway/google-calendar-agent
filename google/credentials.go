package google

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	config "github.com/inference-gateway/google-calendar-agent/config"
	zap "go.uber.org/zap"
)

// CreateCredentialsFile creates a Google credentials JSON file from configuration
func CreateCredentialsFile(l *zap.Logger, cfg *config.Config) error {
	credentialsType, credentials, err := cfg.GetGoogleCredentialsOption()
	if err != nil {
		return fmt.Errorf("failed to get google credentials: %w", err)
	}

	if credentials == "" {
		l.Debug("google credentials not set, skipping credentials file creation")
		return nil
	}

	switch credentialsType {
	case "file":
		// If they only provided GOOGLE_APPLICATION_CREDENTIALS - the user is most likely mounting a credentials file directly
		l.Debug("using existing credentials file", zap.String("path", credentials))
		return nil
	case "json":
		credentialsPath := cfg.Google.CredentialsPath
		if credentialsPath == "" {
			credentialsPath = "/tmp/google-credentials.json"
		}

		var temp interface{}
		if err := json.Unmarshal([]byte(credentials), &temp); err != nil {
			return fmt.Errorf("invalid json content in google service account credentials: %w", err)
		}

		dir := filepath.Dir(credentialsPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		if err := os.WriteFile(credentialsPath, []byte(credentials), 0600); err != nil {
			return fmt.Errorf("failed to write google credentials file %s: %w", credentialsPath, err)
		}

		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsPath); err != nil {
			return fmt.Errorf("failed to set GOOGLE_APPLICATION_CREDENTIALS environment variable: %w", err)
		}

		l.Debug("google credentials file created from JSON content", zap.String("path", credentialsPath))
		return nil
	default:
		return fmt.Errorf("unknown credentials type: %s", credentialsType)
	}
}
