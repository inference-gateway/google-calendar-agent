package toolbox

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/inference-gateway/adk/types"
	server "github.com/inference-gateway/adk/server"
	zap "go.uber.org/zap"
)

// DemoTaskHandler implements TaskHandler interface for demo mode
type DemoTaskHandler struct {
	toolBox *server.DefaultToolBox
	logger  *zap.Logger
	agent   server.OpenAICompatibleAgent
}

// NewDemoTaskHandler creates a new demo task handler
func NewDemoTaskHandler(toolBox *server.DefaultToolBox, logger *zap.Logger) *DemoTaskHandler {
	return &DemoTaskHandler{
		toolBox: toolBox,
		logger:  logger,
	}
}

// HandleTask processes tasks in demo mode by pattern matching and calling appropriate tools
func (d *DemoTaskHandler) HandleTask(ctx context.Context, task *types.Task, message *types.Message) (*types.Task, error) {
	d.logger.Info("Demo task handler processing task", zap.String("task_id", task.ID))

	var userMessage string
	if message != nil && message.Role == "user" {
		for _, part := range message.Parts {
			if partMap, ok := part.(map[string]interface{}); ok {
				if partMap["kind"] == "text" {
					if text, exists := partMap["text"]; exists {
						if textStr, ok := text.(string); ok {
							userMessage = strings.ToLower(textStr)
							break
						}
					}
				}
			}
		}
	}

	d.logger.Debug("Processing user message", zap.String("message", userMessage))

	var toolName string
	var toolArgs map[string]interface{}

	if strings.Contains(userMessage, "list") || strings.Contains(userMessage, "show") || strings.Contains(userMessage, "events") {
		toolName = "list_calendar_events"
		toolArgs = map[string]interface{}{
			"maxResults": 10,
		}
	} else if strings.Contains(userMessage, "create") || strings.Contains(userMessage, "schedule") || strings.Contains(userMessage, "book") {
		toolName = "create_calendar_event"
		toolArgs = map[string]interface{}{
			"summary":   "Demo Event",
			"startTime": time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			"endTime":   time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		}
	} else if strings.Contains(userMessage, "find") && strings.Contains(userMessage, "time") {
		toolName = "find_available_time"
		toolArgs = map[string]interface{}{
			"startDate": time.Now().Format(time.RFC3339),
			"endDate":   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			"duration":  60,
		}
	} else {
		toolName = "list_calendar_events"
		toolArgs = map[string]interface{}{
			"maxResults": 10,
		}
	}

	if !d.toolBox.HasTool(toolName) {
		d.logger.Error("Tool not found", zap.String("tool_name", toolName))
		return task, fmt.Errorf("tool not found: %s", toolName)
	}

	d.logger.Info("Calling tool", zap.String("tool_name", toolName), zap.Any("args", toolArgs))
	result, err := d.toolBox.ExecuteTool(ctx, toolName, toolArgs)
	if err != nil {
		d.logger.Error("Tool call failed", zap.Error(err))
		return task, fmt.Errorf("tool call failed: %w", err)
	}

	responseMsg := &types.Message{
		Role: "assistant",
		Parts: []types.Part{
			map[string]interface{}{
				"kind": "text",
				"text": fmt.Sprintf("I've processed your request using the %s tool. Here's the result:\n\n%s", toolName, result),
			},
		},
	}

	if message != nil {
		task.History = append(task.History, *message)
	}
	task.History = append(task.History, *responseMsg)

	task.Status.State = types.TaskStateCompleted
	task.Status.Message = responseMsg
	now := time.Now().Format(time.RFC3339)
	task.Status.Timestamp = &now

	d.logger.Info("Demo task completed successfully", zap.String("task_id", task.ID))
	return task, nil
}

// SetAgent sets the OpenAI-compatible agent for the task handler
func (d *DemoTaskHandler) SetAgent(agent server.OpenAICompatibleAgent) {
	d.agent = agent
}

// GetAgent returns the configured OpenAI-compatible agent
func (d *DemoTaskHandler) GetAgent() server.OpenAICompatibleAgent {
	return d.agent
}

// Mock Response Helpers
// This file contains all mock implementations for Google Calendar operations.
// These methods provide realistic mock responses for testing and development environments.

// getMockEvents returns mock calendar events for testing
func (g *GoogleCalendarTools) getMockEvents() string {
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

// getMockCreateEvent returns a mock response for event creation
func (g *GoogleCalendarTools) getMockCreateEvent(args map[string]interface{}) string {
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

// getMockUpdateEvent returns a mock response for event updates
func (g *GoogleCalendarTools) getMockUpdateEvent(args map[string]interface{}) string {
	result := map[string]interface{}{
		"success": true,
		"eventId": args["eventId"],
		"message": "Event would be updated (mock mode)",
		"mock":    true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

// getMockDeleteEvent returns a mock response for event deletion
func (g *GoogleCalendarTools) getMockDeleteEvent(args map[string]interface{}) string {
	result := map[string]interface{}{
		"success": true,
		"eventId": args["eventId"],
		"message": "Event would be deleted (mock mode)",
		"mock":    true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

// getMockGetEvent returns a mock response for getting event details
func (g *GoogleCalendarTools) getMockGetEvent(args map[string]interface{}) string {
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

// getMockAvailableTime returns mock available time slots
func (g *GoogleCalendarTools) getMockAvailableTime(args map[string]interface{}) string {
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

// getMockConflicts returns mock conflict checking results
func (g *GoogleCalendarTools) getMockConflicts(args map[string]interface{}) string {
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
