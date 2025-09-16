package skills

import (
	"context"
	"encoding/json"
	"fmt"

	server "github.com/inference-gateway/adk/server"
	google "github.com/inference-gateway/google-calendar-agent/internal/google"
	zap "go.uber.org/zap"
)

// GetCalendarEventSkill struct holds the skill with dependencies
type GetCalendarEventSkill struct {
	logger *zap.Logger
	google google.CalendarService
}

// NewGetCalendarEventSkill creates a new get_calendar_event skill
func NewGetCalendarEventSkill(logger *zap.Logger, google google.CalendarService) server.Tool {
	skill := &GetCalendarEventSkill{
		logger: logger,
		google: google,
	}
	return server.NewBasicTool(
		"get_calendar_event",
		"Get details of a specific event from Google Calendar",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"eventId": map[string]any{
					"description": "Event ID to retrieve (required)",
					"type":        "string",
				},
			},
			"required": []string{"eventId"},
		},
		skill.GetCalendarEventHandler,
	)
}

// GetCalendarEventHandler handles the get_calendar_event skill execution
func (s *GetCalendarEventSkill) GetCalendarEventHandler(ctx context.Context, args map[string]any) (string, error) {
	s.logger.Debug("getting calendar event", zap.Any("args", args))

	eventID, ok := args["eventId"].(string)
	if !ok || eventID == "" {
		return "", fmt.Errorf("eventId is required")
	}

	calendarID := s.google.GetCalendarID()
	event, err := s.google.GetEvent(calendarID, eventID)
	if err != nil {
		s.logger.Error("failed to get calendar event", zap.Error(err), zap.String("eventId", eventID))
		return "", fmt.Errorf("failed to get calendar event: %w", err)
	}

	s.logger.Info("calendar event retrieved successfully",
		zap.String("eventId", event.Id),
		zap.String("summary", event.Summary))

	result := map[string]any{
		"success": true,
		"eventId": event.Id,
		"summary": event.Summary,
		"status":  event.Status,
	}

	if event.Start != nil {
		result["startTime"] = event.Start.DateTime
	}
	if event.End != nil {
		result["endTime"] = event.End.DateTime
	}
	if event.Description != "" {
		result["description"] = event.Description
	}
	if event.Location != "" {
		result["location"] = event.Location
	}
	if event.HtmlLink != "" {
		result["htmlLink"] = event.HtmlLink
	}
	if len(event.Attendees) > 0 {
		var attendees []string
		for _, attendee := range event.Attendees {
			attendees = append(attendees, attendee.Email)
		}
		result["attendees"] = attendees
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}
