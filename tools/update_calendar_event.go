package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"

	server "github.com/inference-gateway/adk/server"

	google "github.com/inference-gateway/google-calendar-agent/internal/google"
)

// UpdateCalendarEventTool struct holds the tool with dependencies
type UpdateCalendarEventTool struct {
	logger *zap.Logger
	google google.CalendarService
}

// NewUpdateCalendarEventTool creates a new update_calendar_event tool
func NewUpdateCalendarEventTool(logger *zap.Logger, google google.CalendarService) server.Tool {
	tool := &UpdateCalendarEventTool{
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
					"description": "End time in RFC3339 format, or YYYY-MM-DD (exclusive) for an all-day event. Optional.",
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
					"description": "Start time in RFC3339 format, or YYYY-MM-DD for an all-day event. Optional.",
					"type":        "string",
				},
				"summary": map[string]any{
					"description": "Event title/summary. Optional.",
					"type":        "string",
				},
			},
			"required": []string{"eventId"},
		},
		tool.UpdateCalendarEventHandler,
	)
}

// UpdateCalendarEventHandler handles the update_calendar_event tool execution
func (s *UpdateCalendarEventTool) UpdateCalendarEventHandler(ctx context.Context, args map[string]any) (string, error) {
	span := startToolSpan(ctx, "update_calendar_event")
	defer span.End()
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

	if v, exists := args["summary"]; exists && v != nil {
		s, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("summary must be a string, got %T", v)
		}
		existingEvent.Summary = s
	}

	if v, exists := args["description"]; exists && v != nil {
		s, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("description must be a string, got %T", v)
		}
		existingEvent.Description = s
	}

	if v, exists := args["location"]; exists && v != nil {
		s, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("location must be a string, got %T", v)
		}
		existingEvent.Location = s
	}

	recurring := len(existingEvent.Recurrence) > 0

	if v, exists := args["startTime"]; exists && v != nil {
		s, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("startTime must be a string, got %T", v)
		}
		start, err := rescheduleEventDateTime(existingEvent.Start, s, recurring)
		if err != nil {
			return "", fmt.Errorf("invalid startTime: %w", err)
		}
		existingEvent.Start = start
	}

	if v, exists := args["endTime"]; exists && v != nil {
		s, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("endTime must be a string, got %T", v)
		}
		end, err := rescheduleEventDateTime(existingEvent.End, s, recurring)
		if err != nil {
			return "", fmt.Errorf("invalid endTime: %w", err)
		}
		existingEvent.End = end
	}

	if mixesAllDayAndTimed(existingEvent) {
		return "", fmt.Errorf("startTime and endTime must both be dates (YYYY-MM-DD) for an all-day event or both be RFC3339 date-times for a timed event")
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
		"startTime": google.EventDateTimeValue(updatedEvent.Start),
		"endTime":   google.EventDateTimeValue(updatedEvent.End),
		"allDay":    google.IsAllDay(updatedEvent),
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

// rescheduleEventDateTime applies a new start or end value while keeping the existing time zone.
// A YYYY-MM-DD value keeps or makes the boundary all-day, an RFC3339 value makes it timed.
// Recurring events require a time zone, so one is filled from the user's zone when missing.
func rescheduleEventDateTime(current *calendar.EventDateTime, value string, recurring bool) (*calendar.EventDateTime, error) {
	updated := &calendar.EventDateTime{}
	if current != nil {
		*updated = *current
	}

	switch {
	case isRFC3339(value):
		updated.DateTime = value
		updated.Date = ""
	case isDate(value):
		updated.Date = value
		updated.DateTime = ""
	default:
		return nil, fmt.Errorf("%q is neither an RFC3339 date-time nor a YYYY-MM-DD date", value)
	}

	if recurring && updated.TimeZone == "" {
		_, tzName, _ := resolveTimezone()
		updated.TimeZone = tzName
	}
	return updated, nil
}

// mixesAllDayAndTimed reports whether one boundary is a date and the other a date-time,
// which the Google Calendar API rejects.
func mixesAllDayAndTimed(event *calendar.Event) bool {
	if event.Start == nil || event.End == nil {
		return false
	}
	return (event.Start.Date != "") != (event.End.Date != "")
}

func isRFC3339(value string) bool {
	_, err := time.Parse(time.RFC3339, value)
	return err == nil
}

func isDate(value string) bool {
	_, err := time.Parse(google.DateLayout, value)
	return err == nil
}
