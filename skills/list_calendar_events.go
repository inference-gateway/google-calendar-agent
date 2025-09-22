package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	server "github.com/inference-gateway/adk/server"
	google "github.com/inference-gateway/google-calendar-agent/internal/google"
	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"
)

// ListCalendarEventsSkill struct holds the skill with dependencies
type ListCalendarEventsSkill struct {
	logger *zap.Logger
	google google.CalendarService
}

// NewListCalendarEventsSkill creates a new list_calendar_events skill
func NewListCalendarEventsSkill(logger *zap.Logger, google google.CalendarService) server.Tool {
	skill := &ListCalendarEventsSkill{
		logger: logger,
		google: google,
	}
	return server.NewBasicTool(
		"list_calendar_events",
		"List upcoming events from Google Calendar",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"maxResults": map[string]any{
					"description": "Maximum number of events to return (default: 10, max: 100)",
					"maximum":     100,
					"minimum":     1,
					"type":        "integer",
				},
				"query": map[string]any{
					"description": "Free text search terms to find events. Optional.",
					"type":        "string",
				},
				"timeMax": map[string]any{
					"description": "End time (RFC3339 format, e.g., 2024-01-01T23:59:59Z). Optional.",
					"type":        "string",
				},
				"timeMin": map[string]any{
					"description": "Start time (RFC3339 format, e.g., 2024-01-01T00:00:00Z). Defaults to now.",
					"type":        "string",
				},
			},
		},
		skill.ListCalendarEventsHandler,
	)
}

// ListCalendarEventsHandler handles the list_calendar_events skill execution
func (s *ListCalendarEventsSkill) ListCalendarEventsHandler(ctx context.Context, args map[string]any) (string, error) {
	s.logger.Debug("listing calendar events", zap.Any("args", args))

	maxResults := 10
	if mr, exists := args["maxResults"]; exists && mr != nil {
		if mrInt, ok := mr.(float64); ok {
			maxResults = int(mrInt)
		}
	}

	query := ""
	if q, exists := args["query"]; exists && q != nil {
		query = q.(string)
	}

	timeMin := time.Now()
	if tm, exists := args["timeMin"]; exists && tm != nil {
		if tmStr, ok := tm.(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, tmStr); err == nil {
				timeMin = parsedTime
			}
		}
	}

	timeMax := time.Time{}
	if tm, exists := args["timeMax"]; exists && tm != nil {
		if tmStr, ok := tm.(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, tmStr); err == nil {
				timeMax = parsedTime
			}
		}
	}

	calendarID := s.google.GetCalendarID()
	events, err := s.google.ListEvents(calendarID, timeMin, timeMax)
	if err != nil {
		s.logger.Error("failed to list calendar events", zap.Error(err))
		return "", fmt.Errorf("failed to list calendar events: %w", err)
	}

	filteredEvents := events
	if query != "" {
		filteredEvents = []*calendar.Event{}
		for _, event := range events {
			if strings.Contains(strings.ToLower(event.Summary), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(event.Description), strings.ToLower(query)) {
				filteredEvents = append(filteredEvents, event)
			}
		}
	}

	if len(filteredEvents) > maxResults {
		filteredEvents = filteredEvents[:maxResults]
	}

	s.logger.Info("calendar events retrieved successfully", zap.Int("count", len(filteredEvents)))

	var eventList []map[string]any
	for _, event := range filteredEvents {
		eventData := map[string]any{
			"eventId": event.Id,
			"summary": event.Summary,
			"status":  event.Status,
		}

		if event.Start != nil {
			eventData["startTime"] = event.Start.DateTime
		}
		if event.End != nil {
			eventData["endTime"] = event.End.DateTime
		}
		if event.Description != "" {
			eventData["description"] = event.Description
		}
		if event.Location != "" {
			eventData["location"] = event.Location
		}
		if event.HtmlLink != "" {
			eventData["htmlLink"] = event.HtmlLink
		}
		if len(event.Attendees) > 0 {
			var attendees []string
			for _, attendee := range event.Attendees {
				attendees = append(attendees, attendee.Email)
			}
			eventData["attendees"] = attendees
		}

		eventList = append(eventList, eventData)
	}

	result := map[string]any{
		"success": true,
		"events":  eventList,
		"count":   len(eventList),
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}
