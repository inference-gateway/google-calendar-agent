package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	server "github.com/inference-gateway/adk/server"
	google "github.com/inference-gateway/google-calendar-agent/internal/google"
	zap "go.uber.org/zap"
)

// CheckConflictsSkill struct holds the skill with dependencies
type CheckConflictsSkill struct {
	logger *zap.Logger
	google google.CalendarService
}

// NewCheckConflictsSkill creates a new check_conflicts skill
func NewCheckConflictsSkill(logger *zap.Logger, google google.CalendarService) server.Tool {
	skill := &CheckConflictsSkill{
		logger: logger,
		google: google,
	}
	return server.NewBasicTool(
		"check_conflicts",
		"Check for scheduling conflicts in the specified time range",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"endTime": map[string]any{
					"description": "End time to check (RFC3339 format, required)",
					"type":        "string",
				},
				"startTime": map[string]any{
					"description": "Start time to check (RFC3339 format, required)",
					"type":        "string",
				},
			},
			"required": []string{"startTime", "endTime"},
		},
		skill.CheckConflictsHandler,
	)
}

// CheckConflictsHandler handles the check_conflicts skill execution
func (s *CheckConflictsSkill) CheckConflictsHandler(ctx context.Context, args map[string]any) (string, error) {
	s.logger.Debug("checking for conflicts", zap.Any("args", args))

	startTimeStr, ok := args["startTime"].(string)
	if !ok || startTimeStr == "" {
		return "", fmt.Errorf("startTime is required")
	}

	endTimeStr, ok := args["endTime"].(string)
	if !ok || endTimeStr == "" {
		return "", fmt.Errorf("endTime is required")
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return "", fmt.Errorf("invalid startTime format: %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return "", fmt.Errorf("invalid endTime format: %w", err)
	}

	calendarID := s.google.GetCalendarID()
	conflicts, err := s.google.CheckConflicts(calendarID, startTime, endTime)
	if err != nil {
		s.logger.Error("failed to check conflicts", zap.Error(err))
		return "", fmt.Errorf("failed to check conflicts: %w", err)
	}

	s.logger.Info("conflicts check completed", zap.Int("conflictCount", len(conflicts)))

	hasConflicts := len(conflicts) > 0
	var conflictList []map[string]any
	for _, conflict := range conflicts {
		conflictData := map[string]any{
			"eventId": conflict.Id,
			"summary": conflict.Summary,
		}

		if conflict.Start != nil {
			conflictData["startTime"] = conflict.Start.DateTime
		}
		if conflict.End != nil {
			conflictData["endTime"] = conflict.End.DateTime
		}
		if conflict.Location != "" {
			conflictData["location"] = conflict.Location
		}

		conflictList = append(conflictList, conflictData)
	}

	result := map[string]any{
		"success":       true,
		"hasConflicts":  hasConflicts,
		"conflicts":     conflictList,
		"conflictCount": len(conflicts),
		"timeRange": map[string]string{
			"startTime": startTimeStr,
			"endTime":   endTimeStr,
		},
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}
