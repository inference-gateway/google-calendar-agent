package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	zap "go.uber.org/zap"

	server "github.com/inference-gateway/adk/server"
)

// GetCurrentDatetimeTool struct holds the tool with services
type GetCurrentDatetimeTool struct {
	logger *zap.Logger
}

// NewGetCurrentDatetimeTool creates a new get_current_datetime tool
func NewGetCurrentDatetimeTool(logger *zap.Logger) server.Tool {
	tool := &GetCurrentDatetimeTool{
		logger: logger,
	}
	return server.NewBasicTool(
		"get_current_datetime",
		"Return the current date/time and the user's IANA timezone. Call this FIRST for any time-relative request (today, tomorrow, next Friday) before emitting RFC3339 timestamps to other calendar tools, so events land in the user's local timezone instead of an LLM-assumed default.",
		map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		tool.GetCurrentDatetimeHandler,
	)
}

// resolveTimezone picks the user's timezone in this order:
//  1. GOOGLE_CALENDAR_TIMEZONE (agent-specific override)
//  2. TZ (standard POSIX system timezone)
//  3. UTC fallback
//
// Returns the loaded *time.Location, the IANA name reported back to the
// LLM, and the source label for logging.
func resolveTimezone() (*time.Location, string, string) {
	candidates := []struct {
		name  string
		value string
	}{
		{"GOOGLE_CALENDAR_TIMEZONE", os.Getenv("GOOGLE_CALENDAR_TIMEZONE")},
		{"TZ", os.Getenv("TZ")},
	}
	for _, c := range candidates {
		if c.value == "" {
			continue
		}
		loc, err := time.LoadLocation(c.value)
		if err != nil {
			continue
		}
		return loc, c.value, c.name
	}
	return time.UTC, "UTC", "default"
}

// GetCurrentDatetimeHandler handles the get_current_datetime tool execution
func (t *GetCurrentDatetimeTool) GetCurrentDatetimeHandler(ctx context.Context, args map[string]any) (string, error) {
	loc, tzName, source := resolveTimezone()
	now := time.Now().In(loc)

	t.logger.Debug("resolved current datetime",
		zap.String("timezone", tzName),
		zap.String("source", source),
		zap.String("now", now.Format(time.RFC3339)))

	result := map[string]any{
		"now":             now.Format(time.RFC3339),
		"timezone":        tzName,
		"timezone_source": source,
		"weekday":         now.Weekday().String(),
		"date":            now.Format("2006-01-02"),
		"time":            now.Format("15:04:05"),
		"utc_offset":      now.Format("-07:00"),
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(resultJSON), nil
}
