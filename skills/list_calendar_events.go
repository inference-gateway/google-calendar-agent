package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	server "github.com/inference-gateway/adk/server"
	a2a "github.com/inference-gateway/adk/types"
	config "github.com/inference-gateway/google-calendar-agent/config"
	google "github.com/inference-gateway/google-calendar-agent/internal/google"
	envconfig "github.com/sethvargo/go-envconfig"
	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"
	option "google.golang.org/api/option"
)

// ListCalendarEventsSkill struct holds the skill with logger
type ListCalendarEventsSkill struct {
	logger     *zap.Logger
	config     *config.Config
	calSvc     google.CalendarService
	isMockMode bool
}

// NewListCalendarEventsSkill creates a new list-calendar-events skill
func NewListCalendarEventsSkill(logger *zap.Logger) server.Tool {
	skill := &ListCalendarEventsSkill{
		logger: logger,
	}

	// Initialize configuration and calendar service
	skill.initializeConfig()
	skill.initializeCalendarService()

	return server.NewBasicTool(
		"list-calendar-events",
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

// initializeConfig loads configuration from environment
func (s *ListCalendarEventsSkill) initializeConfig() {
	s.config = &config.Config{}
	ctx := context.Background()
	if err := envconfig.Process(ctx, s.config); err != nil {
		s.logger.Warn("Failed to load config, using defaults", zap.Error(err))
		s.config = &config.Config{
			Environment: "dev",
		}
	}
}

// initializeCalendarService sets up the calendar service
func (s *ListCalendarEventsSkill) initializeCalendarService() {
	if s.config.ShouldUseMockService() {
		s.isMockMode = true
		s.logger.Info("List calendar events skill initialized in mock mode")
		return
	}

	ctx := context.Background()
	var opts []option.ClientOption
	if s.config.Google.ServiceAccountJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(s.config.Google.ServiceAccountJSON)))
	} else if s.config.Google.CredentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(s.config.Google.CredentialsPath))
	}

	calSvc, err := google.NewCalendarService(ctx, s.config, s.logger, opts...)
	if err != nil {
		if s.config.Environment == "dev" {
			s.logger.Warn("Failed to initialize Google Calendar service, falling back to mock mode", zap.Error(err))
			s.isMockMode = true
		} else {
			s.logger.Error("Failed to create Google Calendar service", zap.Error(err))
		}
	} else {
		s.calSvc = calSvc
		s.logger.Info("✅ Google Calendar service initialized successfully for list events")
	}
}

// ListCalendarEventsHandler handles the list-calendar-events skill execution
func (s *ListCalendarEventsSkill) ListCalendarEventsHandler(ctx context.Context, args map[string]any) (string, error) {
	s.logger.Info("Processing list-calendar-events request", zap.Any("args", args))

	if s.isMockMode {
		s.logger.Debug("returning mock events")
		return s.getMockEvents(), nil
	}

	s.logger.Debug("processing list events request in non-mock mode")

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

	events, err := s.calSvc.ListEvents(s.config.Google.CalendarID, timeMin, timeMax)
	if err != nil {
		return "", fmt.Errorf("failed to list events: %w", err)
	}

	response := a2a.CalendarEventResponse{
		Events:  events,
		Message: fmt.Sprintf("Found %d events between %s and %s", len(events), timeMin.Format("2006-01-02 15:04"), timeMax.Format("2006-01-02 15:04")),
		Success: true,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	s.logger.Info("✅ Successfully listed events", zap.Int("count", len(events)))
	return string(jsonBytes), nil
}

// getMockEvents returns mock events for testing
func (s *ListCalendarEventsSkill) getMockEvents() string {
	mockEvents := []*calendar.Event{
		{
			Id:          "mock-event-1",
			Summary:     "Team Meeting",
			Description: "Weekly team standup meeting",
			Start: &calendar.EventDateTime{
				DateTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
			},
			Location: "Conference Room A",
		},
		{
			Id:          "mock-event-2",
			Summary:     "Client Call",
			Description: "Quarterly review with client",
			Start: &calendar.EventDateTime{
				DateTime: time.Now().Add(3 * time.Hour).Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: time.Now().Add(4 * time.Hour).Format(time.RFC3339),
			},
			Location: "Virtual",
		},
	}

	response := a2a.CalendarEventResponse{
		Events:  mockEvents,
		Message: fmt.Sprintf("Found %d mock events for demonstration", len(mockEvents)),
		Success: true,
	}

	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes)
}
