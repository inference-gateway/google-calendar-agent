package skills

import (
	"context"
	"encoding/json"
	"fmt"

	server "github.com/inference-gateway/adk/server"
	google "github.com/inference-gateway/google-calendar-agent/internal/google"
	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"
)

// UpdateCalendarEventSkill struct holds the skill with dependencies
type UpdateCalendarEventSkill struct {
	logger *zap.Logger
	google google.CalendarService
}

// NewUpdateCalendarEventSkill creates a new update_calendar_event skill
func NewUpdateCalendarEventSkill(logger *zap.Logger, google google.CalendarService) server.Tool {
	skill := &UpdateCalendarEventSkill{
		logger: logger,
		google: google,
	}
	return server.NewBasicTool(
		"update_calendar_event",
		"Update an existing event in Google Calendar",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"description": map[string]any{
					"description": "Event description. Optional.",
					"type":        "string",
				},
				"endTime": map[string]any{
					"description": "End time in RFC3339 format. Optional.",
					"type":        "string",
				},
				"eventId": map[string]any{
					"description": "Event ID to update (required)",
					"type":        "string",
				},
				"location": map[string]any{
					"description": "Event location. Optional.",
					"type":        "string",
				},
				"startTime": map[string]any{
					"description": "Start time in RFC3339 format. Optional.",
					"type":        "string",
				},
				"summary": map[string]any{
					"description": "Event title/summary. Optional.",
					"type":        "string",
				},
			},
			"required": []string{"eventId"},
		},
		skill.UpdateCalendarEventHandler,
	)
}

// UpdateCalendarEventHandler handles the update_calendar_event skill execution
func (s *UpdateCalendarEventSkill) UpdateCalendarEventHandler(ctx context.Context, args map[string]any) (string, error) {
	s.logger.Debug("updating calendar event", zap.Any("args", args))

	eventID, ok := args["eventId"].(string)
	if !ok || eventID == "" {
		return "", fmt.Errorf("eventId is required")
	}

	calendarID := s.google.GetCalendarID()
	existingEvent, err := s.google.GetEvent(calendarID, eventID)
	if err != nil {
		s.logger.Error("failed to get existing calendar event", zap.Error(err), zap.String("eventId", eventID))
		return "", fmt.Errorf("failed to get existing calendar event: %w", err)
	}

	if summary, exists := args["summary"]; exists && summary != nil {
		existingEvent.Summary = summary.(string)
	}

	if description, exists := args["description"]; exists && description != nil {
		existingEvent.Description = description.(string)
	}

	if location, exists := args["location"]; exists && location != nil {
		existingEvent.Location = location.(string)
	}

	if startTime, exists := args["startTime"]; exists && startTime != nil {
		existingEvent.Start = &calendar.EventDateTime{
			DateTime: startTime.(string),
		}
	}

	if endTime, exists := args["endTime"]; exists && endTime != nil {
		existingEvent.End = &calendar.EventDateTime{
			DateTime: endTime.(string),
		}
	}

	updatedEvent, err := s.google.UpdateEvent(calendarID, eventID, existingEvent)
	if err != nil {
		s.logger.Error("failed to update calendar event", zap.Error(err), zap.String("eventId", eventID))
		return "", fmt.Errorf("failed to update calendar event: %w", err)
	}

	s.logger.Info("calendar event updated successfully",
		zap.String("eventId", updatedEvent.Id),
		zap.String("summary", updatedEvent.Summary))

	result := map[string]any{
		"success":   true,
		"eventId":   updatedEvent.Id,
		"summary":   updatedEvent.Summary,
		"startTime": updatedEvent.Start.DateTime,
		"endTime":   updatedEvent.End.DateTime,
		"htmlLink":  updatedEvent.HtmlLink,
	}

	if updatedEvent.Description != "" {
		result["description"] = updatedEvent.Description
	}
	if updatedEvent.Location != "" {
		result["location"] = updatedEvent.Location
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}
