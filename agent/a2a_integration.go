package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/inference-gateway/google-calendar-agent/a2a"
	"go.uber.org/zap"
)

// A2ACalendarTaskManager manages calendar tasks using A2A types
type A2ACalendarTaskManager struct {
	agent        *GoogleCalendarAgent
	errorHandler *A2AErrorHandler
	logger       *zap.Logger
}

// NewA2ACalendarTaskManager creates a new A2A calendar task manager
func NewA2ACalendarTaskManager(agent *GoogleCalendarAgent, logger *zap.Logger) *A2ACalendarTaskManager {
	return &A2ACalendarTaskManager{
		agent:        agent,
		errorHandler: NewA2AErrorHandler(),
		logger:       logger,
	}
}

// ExecuteListEventsTask executes a list events task using A2A types
func (tm *A2ACalendarTaskManager) ExecuteListEventsTask(ctx context.Context, taskID, contextID string, params map[string]interface{}) a2a.Task {
	workingMessage := CreateSuccessMessage(taskID, "📅 Retrieving calendar events...", nil)

	tm.logger.Info("Starting list events task",
		zap.String("taskId", taskID),
		zap.String("contextId", contextID))

	result, err := tm.agent.handleListEvents(ctx, params)
	if err != nil {
		errorTask := tm.errorHandler.CreateErrorTask(taskID, contextID, fmt.Sprintf("Failed to list events: %v", err))
		tm.logger.Error("List events task failed", zap.Error(err), zap.String("taskId", taskID))
		return errorTask
	}

	// Create success message and artifacts
	successMessage := CreateSuccessMessage(taskID, "✅ Successfully retrieved calendar events", map[string]interface{}{
		"result": result,
	})

	// For demonstration, let's assume we parsed the events and create an artifact
	// In a real implementation, you'd parse the JSON result
	artifacts := []a2a.Artifact{
		{
			ArtifactID:  "artifact_" + taskID + "_events",
			Name:        stringPtr("Calendar Events"),
			Description: stringPtr("List of calendar events"),
			Parts: []a2a.Part{
				CreateTextPart(result),
			},
		},
	}

	completedStatus := CreateTaskStatus(a2a.TaskStateCompleted, &successMessage)

	task := CreateTask(contextID, taskID, completedStatus, artifacts, []a2a.Message{workingMessage, successMessage})

	tm.logger.Info("List events task completed successfully", zap.String("taskId", taskID))
	return task
}

// ExecuteCreateEventTask executes a create event task using A2A types
func (tm *A2ACalendarTaskManager) ExecuteCreateEventTask(ctx context.Context, taskID, contextID string, params map[string]interface{}) a2a.Task {
	// Validate required parameters
	if err := tm.validateCreateEventParams(params); err != nil {
		errorTask := tm.errorHandler.CreateErrorTask(taskID, contextID, err.Error())
		tm.logger.Error("Create event task validation failed", zap.Error(err), zap.String("taskId", taskID))
		return errorTask
	}

	// Create initial working status
	workingMessage := CreateSuccessMessage(taskID, "📝 Creating calendar event...", nil)

	tm.logger.Info("Starting create event task",
		zap.String("taskId", taskID),
		zap.String("contextId", contextID),
		zap.Any("params", params))

	// Execute the actual calendar operation
	result, err := tm.agent.handleCreateEvent(ctx, params)
	if err != nil {
		errorTask := tm.errorHandler.CreateErrorTask(taskID, contextID, fmt.Sprintf("Failed to create event: %v", err))
		tm.logger.Error("Create event task failed", zap.Error(err), zap.String("taskId", taskID))
		return errorTask
	}

	// Create success message and artifacts
	successMessage := CreateSuccessMessage(taskID, "✅ Successfully created calendar event", map[string]interface{}{
		"result": result,
	})

	artifacts := []a2a.Artifact{
		{
			ArtifactID:  "artifact_" + taskID + "_created_event",
			Name:        stringPtr("Created Event"),
			Description: stringPtr("Details of the newly created calendar event"),
			Parts: []a2a.Part{
				CreateTextPart(result),
			},
		},
	}

	completedStatus := CreateTaskStatus(a2a.TaskStateCompleted, &successMessage)
	task := CreateTask(contextID, taskID, completedStatus, artifacts, []a2a.Message{workingMessage, successMessage})

	tm.logger.Info("Create event task completed successfully", zap.String("taskId", taskID))
	return task
}

// ProcessA2ARequest processes an A2A request and returns appropriate response
func (tm *A2ACalendarTaskManager) ProcessA2ARequest(ctx context.Context, request a2a.SendMessageRequest) a2a.SendMessageResponse {
	// Extract task information
	taskID := generateUniqueID()
	contextID := "context_" + generateUniqueID()

	// Process the message to determine what action to take
	message := request.Params.Message

	tm.logger.Info("Processing A2A request",
		zap.String("taskId", taskID),
		zap.String("contextId", contextID),
		zap.String("messageId", message.MessageID))

	// For demo purposes, let's assume we have logic to parse the message and determine the action
	// In a real implementation, you'd use NLP or pattern matching

	var task a2a.Task

	// This is a simplified example - in practice you'd parse the message content
	if len(message.Parts) > 0 {
		if textPart, ok := message.Parts[0].(a2a.TextPart); ok {
			if contains(textPart.Text, "list") && contains(textPart.Text, "events") {
				// Execute list events task
				params := map[string]interface{}{
					"timeMin": time.Now().Format(time.RFC3339),
					"timeMax": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				}
				task = tm.ExecuteListEventsTask(ctx, taskID, contextID, params)
			} else if contains(textPart.Text, "create") && contains(textPart.Text, "event") {
				// Execute create event task - this would need more sophisticated parsing
				params := map[string]interface{}{
					"summary":   "Example Event",
					"startTime": time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					"endTime":   time.Now().Add(2 * time.Hour).Format(time.RFC3339),
				}
				task = tm.ExecuteCreateEventTask(ctx, taskID, contextID, params)
			} else {
				// Unknown request
				errorTask := tm.errorHandler.CreateErrorTask(taskID, contextID, "Unknown calendar operation requested")
				task = errorTask
			}
		}
	}

	return a2a.SendMessageSuccessResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  task,
	}
}

// validateCreateEventParams validates parameters for creating an event
func (tm *A2ACalendarTaskManager) validateCreateEventParams(params map[string]interface{}) error {
	if summary, ok := params["summary"].(string); !ok || summary == "" {
		return fmt.Errorf("summary is required and must be a non-empty string")
	}

	if startTime, ok := params["startTime"].(string); !ok || startTime == "" {
		return fmt.Errorf("startTime is required and must be a valid RFC3339 timestamp")
	}

	if endTime, ok := params["endTime"].(string); !ok || endTime == "" {
		return fmt.Errorf("endTime is required and must be a valid RFC3339 timestamp")
	}

	return nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || (len(s) > len(substr) && s[len(s)-len(substr):] == substr)
}
