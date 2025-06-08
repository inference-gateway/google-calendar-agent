package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdk "github.com/inference-gateway/sdk"
	"go.uber.org/zap"

	"github.com/inference-gateway/google-calendar-agent/config"
)

// InferenceGatewayService implements the LLM Service interface using the Inference Gateway
type InferenceGatewayService struct {
	client   sdk.Client
	config   *config.Config
	logger   *zap.Logger
	provider sdk.Provider
	model    string
	enabled  bool
}

// NewInferenceGatewayService creates a new Inference Gateway LLM service
func NewInferenceGatewayService(cfg *config.Config, logger *zap.Logger) (*InferenceGatewayService, error) {
	if !cfg.LLM.Enabled {
		logger.Info("LLM service is disabled")
		return &InferenceGatewayService{
			config:  cfg,
			logger:  logger,
			enabled: false,
		}, nil
	}

	clientOptions := &sdk.ClientOptions{
		BaseURL: cfg.LLM.GatewayURL,
		Timeout: cfg.LLM.Timeout,
		Tools:   buildCalendarTools(),
	}

	client := sdk.NewClient(clientOptions)

	provider := sdk.Provider(cfg.LLM.Provider)

	logger.Info("initialized LLM service",
		zap.String("provider", cfg.LLM.Provider),
		zap.String("model", cfg.LLM.Model),
		zap.String("gatewayURL", cfg.LLM.GatewayURL))

	return &InferenceGatewayService{
		client:   client,
		config:   cfg,
		logger:   logger,
		provider: provider,
		model:    cfg.LLM.Model,
		enabled:  true,
	}, nil
}

// ProcessNaturalLanguage processes natural language input using the LLM with tools
func (s *InferenceGatewayService) ProcessNaturalLanguage(ctx context.Context, input string) (*ProcessingResult, error) {
	if !s.enabled {
		return nil, fmt.Errorf("LLM service is disabled")
	}

	startTime := time.Now()

	systemPrompt := s.buildSystemPrompt()

	messages := []sdk.Message{
		{
			Role:    sdk.System,
			Content: systemPrompt,
		},
		{
			Role:    sdk.User,
			Content: input,
		},
	}

	tools := buildCalendarTools()

	s.logger.Debug("sending request to LLM with tools",
		zap.String("provider", string(s.provider)),
		zap.String("model", s.model),
		zap.String("input", input),
		zap.Int("tools_count", len(*tools)))

	response, err := s.client.WithTools(tools).GenerateContent(ctx, s.provider, s.model, messages)
	if err != nil {
		s.logger.Error("failed to generate content", zap.Error(err))
		return nil, fmt.Errorf("failed to process natural language: %w", err)
	}

	processingTime := time.Since(startTime)

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned from LLM")
	}

	s.logger.Debug("received LLM response",
		zap.Duration("processingTime", processingTime))

	result, err := s.parseToolResponse(response, input)
	if err != nil {
		s.logger.Error("failed to parse LLM response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	result.RawResponse = response.Choices[0].Message.Content
	result.ProcessingTime = processingTime

	if response.Usage != nil {
		result.TokensUsed = &TokenUsage{
			PromptTokens:     int(response.Usage.PromptTokens),
			CompletionTokens: int(response.Usage.CompletionTokens),
			TotalTokens:      int(response.Usage.TotalTokens),
		}
	}

	s.logger.Info("successfully processed natural language",
		zap.String("intent", result.Intent),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("processingTime", processingTime))

	return result, nil
}

// IsEnabled returns whether the LLM service is enabled
func (s *InferenceGatewayService) IsEnabled() bool {
	return s.enabled
}

// GetProvider returns the configured provider
func (s *InferenceGatewayService) GetProvider() string {
	if !s.enabled {
		return ""
	}
	return string(s.provider)
}

// GetModel returns the configured model
func (s *InferenceGatewayService) GetModel() string {
	if !s.enabled {
		return ""
	}
	return s.model
}

// buildSystemPrompt creates the system prompt for calendar operations
func (s *InferenceGatewayService) buildSystemPrompt() string {
	timezone := s.config.Google.TimeZone
	if timezone == "" {
		timezone = "UTC"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		s.logger.Warn("failed to load timezone, using UTC",
			zap.String("timezone", timezone),
			zap.Error(err))
		loc = time.UTC
	}

	now := time.Now().In(loc)
	currentDate := now.Format("2006-01-02")
	currentTime := now.Format("15:04:05")
	currentWeekday := now.Weekday().String()

	return fmt.Sprintf(`You are a helpful calendar assistant that can manage calendar events. You have access to calendar tools to help users with their requests.

Current date and time information:
- Current date: %s (%s)
- Current time: %s
- Timezone: %s

When users ask about calendar operations, use the appropriate tool to help them:
- create_event: Create new calendar events
- list_events: List events in a time range
- update_event: Modify existing events
- delete_event: Remove events
- search_events: Find events by criteria
- get_availability: Check free/busy times

Guidelines for time handling:
- Use the current date/time above as reference for relative time calculations
- For relative times (like "tomorrow", "next week"), calculate absolute dates based on the current date
- All times should be in ISO 8601 format in the specified timezone (%s)
- If no specific time is mentioned, use reasonable defaults (e.g., 1-hour meetings starting at next available hour)

Always be helpful and use the tools to assist with calendar requests. If a request is ambiguous, ask for clarification rather than making assumptions.`,
		currentDate, currentWeekday, currentTime, timezone, timezone)
}

// parseToolResponse parses the tool call response from the LLM
func (s *InferenceGatewayService) parseToolResponse(response *sdk.CreateChatCompletionResponse, originalInput string) (*ProcessingResult, error) {
	choice := response.Choices[0]

	if choice.Message.ToolCalls != nil && len(*choice.Message.ToolCalls) > 0 {
		toolCall := (*choice.Message.ToolCalls)[0]

		var result ProcessingResult

		switch toolCall.Function.Name {
		case "create_event":
			result.Intent = "create_event"
		case "list_events":
			result.Intent = "list_events"
		case "update_event":
			result.Intent = "update_event"
		case "delete_event":
			result.Intent = "delete_event"
		case "search_events":
			result.Intent = "search_events"
		case "get_availability":
			result.Intent = "get_availability"
		default:
			return nil, fmt.Errorf("unknown tool call: %s", toolCall.Function.Name)
		}

		var parameters map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &parameters); err != nil {
			return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
		}

		result.Parameters = parameters
		result.Confidence = 0.95
		result.Response = fmt.Sprintf("I'll help you %s with the provided parameters.", result.Intent)

		return &result, nil
	}

	responseContent := choice.Message.Content

	intent := "question"
	confidence := 0.8

	lowerContent := strings.ToLower(responseContent)
	if strings.Contains(lowerContent, "?") ||
		strings.Contains(lowerContent, "need more") ||
		strings.Contains(lowerContent, "clarify") ||
		strings.Contains(lowerContent, "specify") ||
		strings.Contains(lowerContent, "which") ||
		strings.Contains(lowerContent, "when") ||
		strings.Contains(lowerContent, "what time") ||
		strings.Contains(lowerContent, "could you") ||
		strings.Contains(lowerContent, "please provide") {
		intent = "clarification"
		confidence = 0.9
	}

	return &ProcessingResult{
		Intent:     intent,
		Confidence: confidence,
		Parameters: make(map[string]interface{}),
		Response:   responseContent,
	}, nil
}

// buildCalendarTools creates the tools definition for calendar operations
func buildCalendarTools() *[]sdk.ChatCompletionTool {
	tools := []sdk.ChatCompletionTool{
		{
			Type: sdk.Function,
			Function: sdk.FunctionObject{
				Name:        "create_event",
				Description: stringPtr("Create a new calendar event"),
				Parameters: &sdk.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]interface{}{
							"type":        "string",
							"description": "The title of the event",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "The description of the event",
						},
						"start_time": map[string]interface{}{
							"type":        "string",
							"description": "The start time of the event in ISO 8601 format",
						},
						"end_time": map[string]interface{}{
							"type":        "string",
							"description": "The end time of the event in ISO 8601 format",
						},
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The location of the event",
						},
						"attendees": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "string",
							},
							"description": "List of attendee email addresses",
						},
					},
					"required": []string{"title", "start_time", "end_time"},
				},
			},
		},
		{
			Type: sdk.Function,
			Function: sdk.FunctionObject{
				Name:        "list_events",
				Description: stringPtr("List calendar events in a time range"),
				Parameters: &sdk.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"start_date": map[string]interface{}{
							"type":        "string",
							"description": "The start date for listing events in ISO 8601 format",
						},
						"end_date": map[string]interface{}{
							"type":        "string",
							"description": "The end date for listing events in ISO 8601 format",
						},
					},
					"required": []string{"start_date", "end_date"},
				},
			},
		},
		{
			Type: sdk.Function,
			Function: sdk.FunctionObject{
				Name:        "update_event",
				Description: stringPtr("Update an existing calendar event"),
				Parameters: &sdk.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"event_id": map[string]interface{}{
							"type":        "string",
							"description": "The ID of the event to update",
						},
						"title": map[string]interface{}{
							"type":        "string",
							"description": "The new title of the event",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "The new description of the event",
						},
						"start_time": map[string]interface{}{
							"type":        "string",
							"description": "The new start time of the event in ISO 8601 format",
						},
						"end_time": map[string]interface{}{
							"type":        "string",
							"description": "The new end time of the event in ISO 8601 format",
						},
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The new location of the event",
						},
					},
					"required": []string{"event_id"},
				},
			},
		},
		{
			Type: sdk.Function,
			Function: sdk.FunctionObject{
				Name:        "delete_event",
				Description: stringPtr("Delete a calendar event"),
				Parameters: &sdk.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"event_id": map[string]interface{}{
							"type":        "string",
							"description": "The ID of the event to delete",
						},
					},
					"required": []string{"event_id"},
				},
			},
		},
		{
			Type: sdk.Function,
			Function: sdk.FunctionObject{
				Name:        "search_events",
				Description: stringPtr("Search for calendar events by criteria"),
				Parameters: &sdk.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "The search query to find events",
						},
						"start_date": map[string]interface{}{
							"type":        "string",
							"description": "The start date for searching events in ISO 8601 format",
						},
						"end_date": map[string]interface{}{
							"type":        "string",
							"description": "The end date for searching events in ISO 8601 format",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: sdk.Function,
			Function: sdk.FunctionObject{
				Name:        "get_availability",
				Description: stringPtr("Check availability (free/busy time) in a time range"),
				Parameters: &sdk.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"start_time": map[string]interface{}{
							"type":        "string",
							"description": "The start time for checking availability in ISO 8601 format",
						},
						"end_time": map[string]interface{}{
							"type":        "string",
							"description": "The end time for checking availability in ISO 8601 format",
						},
					},
					"required": []string{"start_time", "end_time"},
				},
			},
		},
	}
	return &tools
}

// stringPtr returns a pointer to a string (helper function)
func stringPtr(s string) *string {
	return &s
}
