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

// CreateCalendarEventSkill struct holds the skill with logger
type CreateCalendarEventSkill struct {
	logger     *zap.Logger
	config     *config.Config
	calSvc     google.CalendarService
	isMockMode bool
}

// NewCreateCalendarEventSkill creates a new create-calendar-event skill
func NewCreateCalendarEventSkill(logger *zap.Logger) server.Tool {
	skill := &CreateCalendarEventSkill{
		logger: logger,
	}

	// Initialize configuration and calendar service
	skill.initializeConfig()
	skill.initializeCalendarService()

	return server.NewBasicTool(
		"create-calendar-event",
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
		skill.CreateCalendarEventHandler,
	)
}

// initializeConfig loads configuration from environment
func (s *CreateCalendarEventSkill) initializeConfig() {
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
func (s *CreateCalendarEventSkill) initializeCalendarService() {
	if s.config.ShouldUseMockService() {
		s.isMockMode = true
		s.logger.Info("Create calendar event skill initialized in mock mode")
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
		s.logger.Info("✅ Google Calendar service initialized successfully for create event")
	}
}

// CreateCalendarEventHandler handles the create-calendar-event skill execution
func (s *CreateCalendarEventSkill) CreateCalendarEventHandler(ctx context.Context, args map[string]any) (string, error) {
	s.logger.Info("Processing create-calendar-event request", zap.Any("args", args))

	// Parse and validate required arguments
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

	// Parse times
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return "", fmt.Errorf("invalid startTime format: %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return "", fmt.Errorf("invalid endTime format: %w", err)
	}

	if s.isMockMode {
		s.logger.Debug("creating mock event")
		return s.createMockEvent(summary, startTime, endTime), nil
	}

	s.logger.Debug("processing create event request in non-mock mode")

	// Build event request
	event := &calendar.Event{
		Summary: summary,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
		},
	}

	// Add optional fields
	if description, ok := args["description"].(string); ok && description != "" {
		event.Description = description
	}

	if location, ok := args["location"].(string); ok && location != "" {
		event.Location = location
	}

	if attendeesList, ok := args["attendees"].([]interface{}); ok {
		for _, attendeeInterface := range attendeesList {
			if attendee, ok := attendeeInterface.(string); ok {
				event.Attendees = append(event.Attendees, &calendar.EventAttendee{
					Email: attendee,
				})
			}
		}
	}

	createdEvent, err := s.calSvc.CreateEvent(s.config.Google.CalendarID, event)
	if err != nil {
		return "", fmt.Errorf("failed to create event: %w", err)
	}

	response := a2a.CalendarEventResponse{
		Events:  []*calendar.Event{createdEvent},
		Message: fmt.Sprintf("Successfully created event '%s' at %s", summary, startTime.Format("2006-01-02 15:04")),
		Success: true,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	s.logger.Info("✅ Successfully created event", zap.String("summary", summary), zap.String("eventId", createdEvent.Id))
	return string(jsonBytes), nil
}

// createMockEvent returns a mock created event for testing
func (s *CreateCalendarEventSkill) createMockEvent(summary string, startTime, endTime time.Time) string {
	mockEvent := &calendar.Event{
		Id:      fmt.Sprintf("mock-event-%d", time.Now().Unix()),
		Summary: summary,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
		},
		Status: "confirmed",
	}

	response := a2a.CalendarEventResponse{
		Events:  []*calendar.Event{mockEvent},
		Message: fmt.Sprintf("Successfully created mock event '%s'", summary),
		Success: true,
	}

	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes)
}
