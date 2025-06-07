package google

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// CreateCredentialsFile creates a Google credentials JSON file from environment variable content
func CreateCredentialsFile(l *zap.Logger) error {
	jsonContent := os.Getenv("GOOGLE_CALENDAR_SA_JSON")
	if jsonContent == "" {
		l.Debug("google_calendar_sa_json environment variable not set, skipping credentials file creation")
		return nil
	}

	var temp interface{}
	if err := json.Unmarshal([]byte(jsonContent), &temp); err != nil {
		return fmt.Errorf("invalid json content in google_calendar_sa_json: %w", err)
	}

	credentialsPath := "/app/secrets/google-credentials.json"

	dir := filepath.Dir(credentialsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(credentialsPath, []byte(jsonContent), 0600); err != nil {
		return fmt.Errorf("failed to write google credentials file %s: %w", credentialsPath, err)
	}

	return nil
}
