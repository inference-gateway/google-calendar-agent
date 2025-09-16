package skills

import (
	"context"
	"encoding/json"
	"fmt"

	server "github.com/inference-gateway/adk/server"
	a2a "github.com/inference-gateway/adk/types"
	config "github.com/inference-gateway/google-calendar-agent/config"
	google "github.com/inference-gateway/google-calendar-agent/internal/google"
	envconfig "github.com/sethvargo/go-envconfig"
	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"
	option "google.golang.org/api/option"
)

// GetCalendarEventSkill struct holds the skill with logger
type GetCalendarEventSkill struct {
	logger     *zap.Logger
	config     *config.Config
	calSvc     google.CalendarService
	isMockMode bool
}

// NewGetCalendarEventSkill creates a new get-calendar-event skill
func NewGetCalendarEventSkill(logger *zap.Logger) server.Tool {
	skill := &GetCalendarEventSkill{
		logger: logger,
	}

	return server.NewBasicTool(
		"get-calendar-event",
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

// initializeConfig loads configuration from environment
func (s *GetCalendarEventSkill) initializeConfig() {
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
func (s *GetCalendarEventSkill) initializeCalendarService() {
	if s.config.ShouldUseMockService() {
		s.isMockMode = true
		s.logger.Info("Get calendar event skill initialized in mock mode")
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
		s.logger.Info("✅ Google Calendar service initialized successfully for get event")
	}
}

// GetCalendarEventHandler handles the get-calendar-event skill execution
func (s *GetCalendarEventSkill) GetCalendarEventHandler(ctx context.Context, args map[string]any) (string, error) {
	s.logger.Info("Processing get-calendar-event request", zap.Any("args", args))

	eventId, ok := args["eventId"].(string)
	if !ok || eventId == "" {
		return "", fmt.Errorf("eventId is required")
	}

	if s.isMockMode {
		s.logger.Debug("returning mock event")
		return s.getMockEvent(eventId), nil
	}

	event, err := s.calSvc.GetEvent(s.config.Google.CalendarID, eventId)
	if err != nil {
		return "", fmt.Errorf("failed to get event: %w", err)
	}

	response := a2a.CalendarEventResponse{
		Event:   event,
		Message: fmt.Sprintf("Retrieved event: %s", event.Summary),
		Success: true,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	s.logger.Info("✅ Successfully retrieved event", zap.String("eventId", eventId))
	return string(jsonBytes), nil
}

func (s *GetCalendarEventSkill) getMockEvent(eventId string) string {
	mockEvent := &calendar.Event{
		Id:          eventId,
		Summary:     "Mock Event",
		Description: "This is a mock event for demonstration purposes",
		Start: &calendar.EventDateTime{
			DateTime: "2024-01-15T10:00:00Z",
		},
		End: &calendar.EventDateTime{
			DateTime: "2024-01-15T11:00:00Z",
		},
		Location: "Mock Location",
		Status:   "confirmed",
	}

	response := a2a.CalendarEventResponse{
		Event:   mockEvent,
		Message: fmt.Sprintf("Retrieved mock event: %s", mockEvent.Summary),
		Success: true,
	}

	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes)
}
