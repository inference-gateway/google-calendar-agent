package toolbox

import (
	"context"
	"fmt"

	server "github.com/inference-gateway/adk/server"
	config "github.com/inference-gateway/google-calendar-agent/config"
	google "github.com/inference-gateway/google-calendar-agent/google"
	zap "go.uber.org/zap"
	option "google.golang.org/api/option"
)

// GoogleCalendarTools provides Google Calendar functionality as A2A tools
type GoogleCalendarTools struct {
	config     *config.Config
	logger     *zap.Logger
	calSvc     google.CalendarService
	isMockMode bool
}

// NewGoogleCalendarTools creates a new Google Calendar tools instance
func NewGoogleCalendarTools(cfg *config.Config, logger *zap.Logger) (*GoogleCalendarTools, error) {
	tools := &GoogleCalendarTools{
		config: cfg,
		logger: logger,
	}

	if cfg.ShouldUseMockService() {
		tools.isMockMode = true
		logger.Info("Google Calendar tools initialized in mock mode")
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
			if cfg.Environment == "dev" {
				logger.Warn("Failed to initialize Google Calendar service, falling back to mock mode", zap.Error(err))
				tools.isMockMode = true
			} else {
				return nil, fmt.Errorf("failed to create Google Calendar service: %w", err)
			}
		} else {
			tools.calSvc = calSvc
			logger.Info("✅ Google Calendar service initialized successfully")
		}
	}

	return tools, nil
}

// RegisterTools registers all Google Calendar tools with the tools handler
func (g *GoogleCalendarTools) RegisterTools(toolBox *server.DefaultToolBox) {
	g.logger.Debug("Registering Google Calendar tools")
	g.registerListEventsTool(toolBox)
	g.registerCreateEventTool(toolBox)
	g.registerUpdateEventTool(toolBox)
	g.registerDeleteEventTool(toolBox)
	g.registerGetEventTool(toolBox)
	g.registerFindAvailableTimeTool(toolBox)
	g.registerCheckConflictsTool(toolBox)
	g.logger.Debug("Google Calendar tools registered successfully")
}

// registerListEventsTool registers the list events tool
func (g *GoogleCalendarTools) registerListEventsTool(toolBox *server.DefaultToolBox) {
	g.logger.Debug("Registering list_calendar_events tool")
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
	g.logger.Debug("list_calendar_events tool registered successfully")
}

// registerCreateEventTool registers the create event tool
func (g *GoogleCalendarTools) registerCreateEventTool(toolBox *server.DefaultToolBox) {
	g.logger.Debug("Registering create_calendar_event tool")
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
	g.logger.Debug("create_calendar_event tool registered successfully")
}

// registerUpdateEventTool registers the update event tool
func (g *GoogleCalendarTools) registerUpdateEventTool(toolBox *server.DefaultToolBox) {
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

// registerDeleteEventTool registers the delete event tool
func (g *GoogleCalendarTools) registerDeleteEventTool(toolBox *server.DefaultToolBox) {
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

// registerGetEventTool registers the get event tool
func (g *GoogleCalendarTools) registerGetEventTool(toolBox *server.DefaultToolBox) {
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

// registerFindAvailableTimeTool registers the find available time tool
func (g *GoogleCalendarTools) registerFindAvailableTimeTool(toolBox *server.DefaultToolBox) {
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

// registerCheckConflictsTool registers the check conflicts tool
func (g *GoogleCalendarTools) registerCheckConflictsTool(toolBox *server.DefaultToolBox) {
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
