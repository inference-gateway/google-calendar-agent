package tools

import (
	"context"
	"encoding/json"
	"fmt"

	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"

	server "github.com/inference-gateway/adk/server"

	google "github.com/inference-gateway/google-calendar-agent/internal/google"
)

// CreateCalendarEventTool struct holds the tool with dependencies
type CreateCalendarEventTool struct {
	logger *zap.Logger
	google google.CalendarService
}

// NewCreateCalendarEventTool creates a new create_calendar_event tool
func NewCreateCalendarEventTool(logger *zap.Logger, google google.CalendarService) server.Tool {
	tool := &CreateCalendarEventTool{
		logger: logger,
		google: google,
	}
	return server.NewBasicTool(
		"create_calendar_event",
		"Create a new event in Google Calendar",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"attendees": map[string]any{
					"description": "List of attendee email addresses. Optional.",
					"items":       map[string]any{"type": "string"},
					"type":        "array",
				},
				"description": map[string]any{
					"description": "Event description. Optional.",
					"type":        "string",
				},
				"endTime": map[string]any{
					"description": "End time in RFC3339 format (required, e.g., 2024-01-01T11:00:00Z)",
					"type":        "string",
				},
				"location": map[string]any{
					"description": "Event location. Optional.",
					"type":        "string",
				},
				"startTime": map[string]any{
					"description": "Start time in RFC3339 format (required, e.g., 2024-01-01T10:00:00Z)",
					"type":        "string",
				},
				"summary": map[string]any{
					"description": "Event title/summary (required)",
					"type":        "string",
				},
			},
			"required": []string{"summary", "startTime", "endTime"},
		},
		tool.CreateCalendarEventHandler,
	)
}

// CreateCalendarEventHandler handles the create_calendar_event tool execution
func (s *CreateCalendarEventTool) CreateCalendarEventHandler(ctx context.Context, args map[string]any) (string, error) {
	span := startToolSpan(ctx, "create_calendar_event")
	defer span.End()
	s.logger.Debug("creating calendar event", zap.Any("args", args))

	summary, ok := args["summary"].(string)
	if !ok || summary == "" {
		return "", fmt.Errorf("summary is required")
	}

	startTime, ok := args["startTime"].(string)
	if !ok || startTime == "" {
		return "", fmt.Errorf("startTime is required")
	}

	endTime, ok := args["endTime"].(string)
	if !ok || endTime == "" {
		return "", fmt.Errorf("endTime is required")
	}

	description := ""
	if desc, exists := args["description"]; exists && desc != nil {
		s, ok := desc.(string)
		if !ok {
			return "", fmt.Errorf("description must be a string, got %T", desc)
		}
		description = s
	}

	location := ""
	if loc, exists := args["location"]; exists && loc != nil {
		s, ok := loc.(string)
		if !ok {
			return "", fmt.Errorf("location must be a string, got %T", loc)
		}
		location = s
	}

	var attendeeEmails []string
	if attendees, exists := args["attendees"]; exists && attendees != nil {
		if attendeeList, ok := attendees.([]any); ok {
			for _, attendee := range attendeeList {
				if email, ok := attendee.(string); ok {
					attendeeEmails = append(attendeeEmails, email)
				}
			}
		}
	}

	event := &calendar.Event{
		Summary:     summary,
		Description: description,
		Location:    location,
		Start: &calendar.EventDateTime{
			DateTime: startTime,
		},
		End: &calendar.EventDateTime{
			DateTime: endTime,
		},
	}

	if len(attendeeEmails) > 0 {
		var attendees []*calendar.EventAttendee
		for _, email := range attendeeEmails {
			attendees = append(attendees, &calendar.EventAttendee{
				Email: email,
			})
		}
		event.Attendees = attendees
	}

	calendarID := s.google.GetCalendarID()
	createdEvent, err := s.google.CreateEvent(calendarID, event)
	if err != nil {
		s.logger.Error("failed to create calendar event", zap.Error(err))
		return "", fmt.Errorf("failed to create calendar event: %w", err)
	}

	s.logger.Info("calendar event created successfully",
		zap.String("eventId", createdEvent.Id),
		zap.String("summary", createdEvent.Summary))

	result := map[string]any{
		"success":   true,
		"eventId":   createdEvent.Id,
		"summary":   createdEvent.Summary,
		"startTime": createdEvent.Start.DateTime,
		"endTime":   createdEvent.End.DateTime,
		"htmlLink":  createdEvent.HtmlLink,
	}

	if createdEvent.Description != "" {
		result["description"] = createdEvent.Description
	}
	if createdEvent.Location != "" {
		result["location"] = createdEvent.Location
	}
	if len(createdEvent.Attendees) > 0 {
		var attendees []string
		for _, attendee := range createdEvent.Attendees {
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
