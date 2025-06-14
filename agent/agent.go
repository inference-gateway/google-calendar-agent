package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/inference-gateway/a2a/adk/server"
	"github.com/inference-gateway/google-calendar-agent/config"
	"github.com/inference-gateway/google-calendar-agent/google"
	"go.uber.org/zap"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// GoogleCalendarAgent wraps the Google Calendar service with A2A tools
type GoogleCalendarAgent struct {
	config     *config.Config
	logger     *zap.Logger
	calSvc     google.CalendarService
	isMockMode bool
}

// NewGoogleCalendarAgent creates a new Google Calendar agent
func NewGoogleCalendarAgent(cfg *config.Config, logger *zap.Logger) (*GoogleCalendarAgent, error) {
	agent := &GoogleCalendarAgent{
		config: cfg,
		logger: logger,
	}

	if cfg.ShouldUseMockService() {
		agent.isMockMode = true
		logger.Info("Google Calendar agent initialized in mock mode")
	} else {
		ctx := context.Background()

		var opts []option.ClientOption
		if cfg.Google.ServiceAccountJSON != "" {
			opts = append(opts, option.WithCredentialsJSON([]byte(cfg.Google.ServiceAccountJSON)))
		} else if cfg.Google.CredentialsPath != "" {
			opts = append(opts, option.WithCredentialsFile(cfg.Google.CredentialsPath))
		}

		calSvc, err := google.NewCalendarService(ctx, cfg, logger, opts...)
		if err != nil {
			if cfg.App.Environment == "dev" {
				// In dev mode, fall back to mock if credentials are missing
				logger.Warn("Failed to initialize Google Calendar service, falling back to mock mode", zap.Error(err))
				agent.isMockMode = true
			} else {
				return nil, fmt.Errorf("failed to create Google Calendar service: %w", err)
			}
		} else {
			agent.calSvc = calSvc
			logger.Info("✅ Google Calendar service initialized successfully")
		}
	}

	return agent, nil
}

// RegisterTools registers all Google Calendar tools with the tools handler
func (g *GoogleCalendarAgent) RegisterTools(toolBox *server.DefaultToolBox) {
	g.registerListEventsTool(toolBox)
	g.registerCreateEventTool(toolBox)
	g.registerUpdateEventTool(toolBox)
	g.registerDeleteEventTool(toolBox)
	g.registerGetEventTool(toolBox)
	g.registerFindAvailableTimeTool(toolBox)
	g.registerCheckConflictsTool(toolBox)
}

// Close performs cleanup
func (g *GoogleCalendarAgent) Close(ctx context.Context) error {
	g.logger.Info("Closing Google Calendar agent")
	return nil
}

// registerListEventsTool registers the list events tool
func (g *GoogleCalendarAgent) registerListEventsTool(toolBox *server.DefaultToolBox) {
	tool := server.NewBasicTool(
		"list_calendar_events",
		"List upcoming events from Google Calendar",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"timeMin": map[string]interface{}{
					"type":        "string",
					"description": "Start time (RFC3339 format, e.g., 2024-01-01T00:00:00Z). Defaults to now.",
				},
				"timeMax": map[string]interface{}{
					"type":        "string",
					"description": "End time (RFC3339 format, e.g., 2024-01-01T23:59:59Z). Optional.",
				},
				"maxResults": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of events to return (default: 10, max: 100)",
					"minimum":     1,
					"maximum":     100,
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Free text search terms to find events. Optional.",
				},
			},
		},
		g.handleListEvents,
	)
	toolBox.AddTool(tool)
}

// handleListEvents handles the list events tool call
func (g *GoogleCalendarAgent) handleListEvents(ctx context.Context, args map[string]interface{}) (string, error) {
	if g.isMockMode {
		return g.getMockEvents(), nil
	}

	timeMin := time.Now()
	if val, ok := args["timeMin"].(string); ok && val != "" {
		if parsedTime, err := time.Parse(time.RFC3339, val); err == nil {
			timeMin = parsedTime
		}
	}

	timeMax := timeMin.Add(24 * time.Hour)
	if val, ok := args["timeMax"].(string); ok && val != "" {
		if parsedTime, err := time.Parse(time.RFC3339, val); err == nil {
			timeMax = parsedTime
		}
	}

	events, err := g.calSvc.ListEvents(g.config.Google.CalendarID, timeMin, timeMax)
	if err != nil {
		return "", fmt.Errorf("failed to list events: %w", err)
	}

	result := map[string]interface{}{
		"events": events,
		"count":  len(events),
		"mock":   false,
	}

	response, _ := json.Marshal(result)
	return string(response), nil
}

// registerCreateEventTool registers the create event tool
func (g *GoogleCalendarAgent) registerCreateEventTool(toolBox *server.DefaultToolBox) {
	tool := server.NewBasicTool(
		"create_calendar_event",
		"Create a new event in Google Calendar",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"summary": map[string]interface{}{
					"type":        "string",
					"description": "Event title/summary (required)",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Event description. Optional.",
				},
				"startTime": map[string]interface{}{
					"type":        "string",
					"description": "Start time in RFC3339 format (required, e.g., 2024-01-01T10:00:00Z)",
				},
				"endTime": map[string]interface{}{
					"type":        "string",
					"description": "End time in RFC3339 format (required, e.g., 2024-01-01T11:00:00Z)",
				},
				"attendees": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "List of attendee email addresses. Optional.",
				},
				"location": map[string]interface{}{
					"type":        "string",
					"description": "Event location. Optional.",
				},
			},
			"required": []string{"summary", "startTime", "endTime"},
		},
		g.handleCreateEvent,
	)
	toolBox.AddTool(tool)
}

// handleCreateEvent handles the create event tool call
func (g *GoogleCalendarAgent) handleCreateEvent(ctx context.Context, args map[string]interface{}) (string, error) {
	if g.isMockMode {
		return g.getMockCreateEvent(args), nil
	}

	summary, ok := args["summary"].(string)
	if !ok || summary == "" {
		return "", fmt.Errorf("summary is required")
	}

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

	if endTime.Before(startTime) {
		return "", fmt.Errorf("endTime must be after startTime")
	}

	event := &calendar.Event{
		Summary: summary,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
		},
	}

	if desc, ok := args["description"].(string); ok && desc != "" {
		event.Description = desc
	}

	if loc, ok := args["location"].(string); ok && loc != "" {
		event.Location = loc
	}

	if attendeesRaw, ok := args["attendees"]; ok {
		if attendeesList, ok := attendeesRaw.([]interface{}); ok {
			var attendees []*calendar.EventAttendee
			for _, attendeeRaw := range attendeesList {
				if email, ok := attendeeRaw.(string); ok {
					attendees = append(attendees, &calendar.EventAttendee{Email: email})
				}
			}
			event.Attendees = attendees
		}
	}

	createdEvent, err := g.calSvc.CreateEvent(g.config.Google.CalendarID, event)
	if err != nil {
		return "", fmt.Errorf("failed to create event: %w", err)
	}

	result := map[string]interface{}{
		"success": true,
		"eventId": createdEvent.Id,
		"message": "Event created successfully",
		"event":   createdEvent,
		"mock":    false,
	}

	response, _ := json.Marshal(result)
	return string(response), nil
}

// registerUpdateEventTool registers the update event tool
func (g *GoogleCalendarAgent) registerUpdateEventTool(toolBox *server.DefaultToolBox) {
	tool := server.NewBasicTool(
		"update_calendar_event",
		"Update an existing event in Google Calendar",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"eventId": map[string]interface{}{
					"type":        "string",
					"description": "Event ID to update (required)",
				},
				"summary": map[string]interface{}{
					"type":        "string",
					"description": "Event title/summary. Optional.",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Event description. Optional.",
				},
				"startTime": map[string]interface{}{
					"type":        "string",
					"description": "Start time in RFC3339 format. Optional.",
				},
				"endTime": map[string]interface{}{
					"type":        "string",
					"description": "End time in RFC3339 format. Optional.",
				},
				"location": map[string]interface{}{
					"type":        "string",
					"description": "Event location. Optional.",
				},
			},
			"required": []string{"eventId"},
		},
		g.handleUpdateEvent,
	)
	toolBox.AddTool(tool)
}

// handleUpdateEvent handles the update event tool call
func (g *GoogleCalendarAgent) handleUpdateEvent(ctx context.Context, args map[string]interface{}) (string, error) {
	if g.isMockMode {
		return g.getMockUpdateEvent(args), nil
	}

	eventId, ok := args["eventId"].(string)
	if !ok || eventId == "" {
		return "", fmt.Errorf("eventId is required")
	}

	existingEvent, err := g.calSvc.GetEvent(g.config.Google.CalendarID, eventId)
	if err != nil {
		return "", fmt.Errorf("failed to get existing event: %w", err)
	}

	if summary, ok := args["summary"].(string); ok && summary != "" {
		existingEvent.Summary = summary
	}

	if desc, ok := args["description"].(string); ok && desc != "" {
		existingEvent.Description = desc
	}

	if loc, ok := args["location"].(string); ok && loc != "" {
		existingEvent.Location = loc
	}

	if startTimeStr, ok := args["startTime"].(string); ok && startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			existingEvent.Start.DateTime = startTime.Format(time.RFC3339)
		}
	}

	if endTimeStr, ok := args["endTime"].(string); ok && endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			existingEvent.End.DateTime = endTime.Format(time.RFC3339)
		}
	}

	updatedEvent, err := g.calSvc.UpdateEvent(g.config.Google.CalendarID, eventId, existingEvent)
	if err != nil {
		return "", fmt.Errorf("failed to update event: %w", err)
	}

	result := map[string]interface{}{
		"success": true,
		"eventId": updatedEvent.Id,
		"message": "Event updated successfully",
		"event":   updatedEvent,
		"mock":    false,
	}

	response, _ := json.Marshal(result)
	return string(response), nil
}

// registerDeleteEventTool registers the delete event tool
func (g *GoogleCalendarAgent) registerDeleteEventTool(toolBox *server.DefaultToolBox) {
	tool := server.NewBasicTool(
		"delete_calendar_event",
		"Delete an event from Google Calendar",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"eventId": map[string]interface{}{
					"type":        "string",
					"description": "Event ID to delete (required)",
				},
			},
			"required": []string{"eventId"},
		},
		g.handleDeleteEvent,
	)
	toolBox.AddTool(tool)
}

// handleDeleteEvent handles the delete event tool call
func (g *GoogleCalendarAgent) handleDeleteEvent(ctx context.Context, args map[string]interface{}) (string, error) {
	if g.isMockMode {
		return g.getMockDeleteEvent(args), nil
	}

	eventId, ok := args["eventId"].(string)
	if !ok || eventId == "" {
		return "", fmt.Errorf("eventId is required")
	}

	err := g.calSvc.DeleteEvent(g.config.Google.CalendarID, eventId)
	if err != nil {
		return "", fmt.Errorf("failed to delete event: %w", err)
	}

	result := map[string]interface{}{
		"success": true,
		"eventId": eventId,
		"message": "Event deleted successfully",
		"mock":    false,
	}

	response, _ := json.Marshal(result)
	return string(response), nil
}

// registerGetEventTool registers the get event tool
func (g *GoogleCalendarAgent) registerGetEventTool(toolBox *server.DefaultToolBox) {
	tool := server.NewBasicTool(
		"get_calendar_event",
		"Get details of a specific event from Google Calendar",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"eventId": map[string]interface{}{
					"type":        "string",
					"description": "Event ID to retrieve (required)",
				},
			},
			"required": []string{"eventId"},
		},
		g.handleGetEvent,
	)
	toolBox.AddTool(tool)
}

// handleGetEvent handles the get event tool call
func (g *GoogleCalendarAgent) handleGetEvent(ctx context.Context, args map[string]interface{}) (string, error) {
	if g.isMockMode {
		return g.getMockGetEvent(args), nil
	}

	eventId, ok := args["eventId"].(string)
	if !ok || eventId == "" {
		return "", fmt.Errorf("eventId is required")
	}

	event, err := g.calSvc.GetEvent(g.config.Google.CalendarID, eventId)
	if err != nil {
		return "", fmt.Errorf("failed to get event: %w", err)
	}

	result := map[string]interface{}{
		"success": true,
		"event":   event,
		"mock":    false,
	}

	response, _ := json.Marshal(result)
	return string(response), nil
}

// registerFindAvailableTimeTool registers the find available time tool
func (g *GoogleCalendarAgent) registerFindAvailableTimeTool(toolBox *server.DefaultToolBox) {
	tool := server.NewBasicTool(
		"find_available_time",
		"Find available time slots in the calendar",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"startDate": map[string]interface{}{
					"type":        "string",
					"description": "Start date for search (RFC3339 format, e.g., 2024-01-01T00:00:00Z)",
				},
				"endDate": map[string]interface{}{
					"type":        "string",
					"description": "End date for search (RFC3339 format, e.g., 2024-01-01T23:59:59Z)",
				},
				"duration": map[string]interface{}{
					"type":        "integer",
					"description": "Duration in minutes for the desired time slot (default: 60)",
					"minimum":     15,
					"maximum":     480,
				},
			},
			"required": []string{"startDate", "endDate"},
		},
		g.handleFindAvailableTime,
	)
	toolBox.AddTool(tool)
}

// handleFindAvailableTime handles the find available time tool call
func (g *GoogleCalendarAgent) handleFindAvailableTime(ctx context.Context, args map[string]interface{}) (string, error) {
	if g.isMockMode {
		return g.getMockAvailableTime(args), nil
	}

	// TODO: Implement real availability search
	// For now, return mock response
	return g.getMockAvailableTime(args), nil
}

// registerCheckConflictsTool registers the check conflicts tool
func (g *GoogleCalendarAgent) registerCheckConflictsTool(toolBox *server.DefaultToolBox) {
	tool := server.NewBasicTool(
		"check_conflicts",
		"Check for scheduling conflicts in the specified time range",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"startTime": map[string]interface{}{
					"type":        "string",
					"description": "Start time to check (RFC3339 format, required)",
				},
				"endTime": map[string]interface{}{
					"type":        "string",
					"description": "End time to check (RFC3339 format, required)",
				},
			},
			"required": []string{"startTime", "endTime"},
		},
		g.handleCheckConflicts,
	)
	toolBox.AddTool(tool)
}

// handleCheckConflicts handles the check conflicts tool call
func (g *GoogleCalendarAgent) handleCheckConflicts(ctx context.Context, args map[string]interface{}) (string, error) {
	if g.isMockMode {
		return g.getMockConflicts(args), nil
	}

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

	conflicts, err := g.calSvc.CheckConflicts(g.config.Google.CalendarID, startTime, endTime)
	if err != nil {
		return "", fmt.Errorf("failed to check conflicts: %w", err)
	}

	result := map[string]interface{}{
		"hasConflicts":   len(conflicts) > 0,
		"conflictCount":  len(conflicts),
		"conflictEvents": conflicts,
		"timeRange": map[string]string{
			"start": startTimeStr,
			"end":   endTimeStr,
		},
		"mock": false,
	}

	response, _ := json.Marshal(result)
	return string(response), nil
}

// Mock response helpers
func (g *GoogleCalendarAgent) getMockEvents() string {
	mockEvents := []map[string]interface{}{
		{
			"id":      "mock-event-1",
			"summary": "Team Meeting",
			"start":   map[string]string{"dateTime": time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
			"end":     map[string]string{"dateTime": time.Now().Add(2 * time.Hour).Format(time.RFC3339)},
		},
		{
			"id":      "mock-event-2",
			"summary": "Lunch with Client",
			"start":   map[string]string{"dateTime": time.Now().Add(4 * time.Hour).Format(time.RFC3339)},
			"end":     map[string]string{"dateTime": time.Now().Add(5 * time.Hour).Format(time.RFC3339)},
		},
	}
	result := map[string]interface{}{
		"events": mockEvents,
		"count":  len(mockEvents),
		"mock":   true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

func (g *GoogleCalendarAgent) getMockCreateEvent(args map[string]interface{}) string {
	result := map[string]interface{}{
		"success":   true,
		"eventId":   fmt.Sprintf("mock-created-event-%d", time.Now().Unix()),
		"message":   "Event would be created (mock mode)",
		"summary":   args["summary"],
		"startTime": args["startTime"],
		"endTime":   args["endTime"],
		"mock":      true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

func (g *GoogleCalendarAgent) getMockUpdateEvent(args map[string]interface{}) string {
	result := map[string]interface{}{
		"success": true,
		"eventId": args["eventId"],
		"message": "Event would be updated (mock mode)",
		"mock":    true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

func (g *GoogleCalendarAgent) getMockDeleteEvent(args map[string]interface{}) string {
	result := map[string]interface{}{
		"success": true,
		"eventId": args["eventId"],
		"message": "Event would be deleted (mock mode)",
		"mock":    true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

func (g *GoogleCalendarAgent) getMockGetEvent(args map[string]interface{}) string {
	result := map[string]interface{}{
		"success": true,
		"event": map[string]interface{}{
			"id":      args["eventId"],
			"summary": "Mock Event",
			"start":   map[string]string{"dateTime": time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
			"end":     map[string]string{"dateTime": time.Now().Add(2 * time.Hour).Format(time.RFC3339)},
		},
		"mock": true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

func (g *GoogleCalendarAgent) getMockAvailableTime(args map[string]interface{}) string {
	duration := 60
	if val, ok := args["duration"].(float64); ok {
		duration = int(val)
	}

	start, _ := time.Parse(time.RFC3339, args["startDate"].(string))
	slots := []map[string]string{
		{
			"start": start.Add(2 * time.Hour).Format(time.RFC3339),
			"end":   start.Add(2*time.Hour + time.Duration(duration)*time.Minute).Format(time.RFC3339),
		},
		{
			"start": start.Add(4 * time.Hour).Format(time.RFC3339),
			"end":   start.Add(4*time.Hour + time.Duration(duration)*time.Minute).Format(time.RFC3339),
		},
	}
	result := map[string]interface{}{
		"availableSlots": slots,
		"count":          len(slots),
		"duration":       duration,
		"mock":           true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

func (g *GoogleCalendarAgent) getMockConflicts(args map[string]interface{}) string {
	result := map[string]interface{}{
		"hasConflicts":   false,
		"conflictCount":  0,
		"conflictEvents": []interface{}{},
		"timeRange": map[string]string{
			"start": args["startTime"].(string),
			"end":   args["endTime"].(string),
		},
		"mock": true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}
