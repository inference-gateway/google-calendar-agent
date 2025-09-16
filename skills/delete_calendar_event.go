package skills

import (
	"context"
	"encoding/json"
	"fmt"

	server "github.com/inference-gateway/adk/server"
	google "github.com/inference-gateway/google-calendar-agent/internal/google"
	zap "go.uber.org/zap"
)

// DeleteCalendarEventSkill struct holds the skill with dependencies
type DeleteCalendarEventSkill struct {
	logger *zap.Logger
	google google.CalendarService
}

// NewDeleteCalendarEventSkill creates a new delete_calendar_event skill
func NewDeleteCalendarEventSkill(logger *zap.Logger, google google.CalendarService) server.Tool {
	skill := &DeleteCalendarEventSkill{
		logger: logger,
		google: google,
	}
	return server.NewBasicTool(
		"delete_calendar_event",
		"Delete an event from Google Calendar",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"eventId": map[string]any{
					"description": "Event ID to delete (required)",
					"type":        "string",
				},
			},
			"required": []string{"eventId"},
		},
		skill.DeleteCalendarEventHandler,
	)
}

// DeleteCalendarEventHandler handles the delete_calendar_event skill execution
func (s *DeleteCalendarEventSkill) DeleteCalendarEventHandler(ctx context.Context, args map[string]any) (string, error) {
	s.logger.Debug("deleting calendar event", zap.Any("args", args))

	eventID, ok := args["eventId"].(string)
	if !ok || eventID == "" {
		return "", fmt.Errorf("eventId is required")
	}

	calendarID := s.google.GetCalendarID()
	err := s.google.DeleteEvent(calendarID, eventID)
	if err != nil {
		s.logger.Error("failed to delete calendar event", zap.Error(err), zap.String("eventId", eventID))
		return "", fmt.Errorf("failed to delete calendar event: %w", err)
	}

	s.logger.Info("calendar event deleted successfully", zap.String("eventId", eventID))

	result := map[string]any{
		"success": true,
		"eventId": eventID,
		"message": "Event deleted successfully",
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}
