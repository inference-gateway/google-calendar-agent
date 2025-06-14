package agent

import (
	"encoding/json"

	a2a "github.com/inference-gateway/google-calendar-agent/a2a"
)

// A2AErrorHandler provides utilities for handling errors using A2A types
type A2AErrorHandler struct{}

// NewA2AErrorHandler creates a new error handler
func NewA2AErrorHandler() *A2AErrorHandler {
	return &A2AErrorHandler{}
}

// HandleTaskNotFound creates a task not found error response
func (h *A2AErrorHandler) HandleTaskNotFound(taskID string) a2a.TaskNotFoundError {
	data := interface{}(map[string]interface{}{"taskId": taskID})
	return a2a.TaskNotFoundError{
		Code:    -32000, // A2A error code for task not found
		Message: "The requested task was not found",
		Data:    &data,
	}
}

// HandleContentTypeNotSupported creates a content type not supported error
func (h *A2AErrorHandler) HandleContentTypeNotSupported(contentType string) a2a.ContentTypeNotSupportedError {
	data := interface{}(map[string]interface{}{"contentType": contentType})
	return a2a.ContentTypeNotSupportedError{
		Code:    -32001, // A2A error code for content type not supported
		Message: "The requested content type is not supported by this agent",
		Data:    &data,
	}
}

// HandleInvalidParams creates an invalid params error
func (h *A2AErrorHandler) HandleInvalidParams(message string, params map[string]interface{}) a2a.InvalidParamsError {
	data := interface{}(params)
	return a2a.InvalidParamsError{
		Code:    -32602, // JSON-RPC standard error code for invalid params
		Message: message,
		Data:    &data,
	}
}

// HandleInternalError creates an internal error response
func (h *A2AErrorHandler) HandleInternalError(message string, details map[string]interface{}) a2a.InternalError {
	data := interface{}(details)
	return a2a.InternalError{
		Code:    -32603, // JSON-RPC standard error code for internal error
		Message: message,
		Data:    &data,
	}
}

// CreateErrorResponse creates a JSON-RPC error response
func (h *A2AErrorHandler) CreateErrorResponse(id interface{}, error interface{}) a2a.JSONRPCErrorResponse {
	return a2a.JSONRPCErrorResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   error,
	}
}

// CreateErrorTask creates a task with error status
func (h *A2AErrorHandler) CreateErrorTask(taskID, contextID, errorMessage string) a2a.Task {
	errorMsg := CreateErrorMessage(taskID, errorMessage)
	status := CreateTaskStatus(a2a.TaskStateFailed, &errorMsg)

	return a2a.Task{
		ID:        taskID,
		ContextID: contextID,
		Kind:      "task",
		Status:    status,
		History:   []a2a.Message{errorMsg},
	}
}

// WrapErrorAsJSON converts an A2A error to JSON string
func (h *A2AErrorHandler) WrapErrorAsJSON(err interface{}) (string, error) {
	jsonBytes, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		return "", marshalErr
	}
	return string(jsonBytes), nil
}

// CalendarServiceError represents calendar-specific errors with A2A structure
type CalendarServiceError struct {
	Code       int                     `json:"code"`
	Message    string                  `json:"message"`
	Data       *map[string]interface{} `json:"data,omitempty"`
	CalendarID string                  `json:"calendarId,omitempty"`
	Operation  string                  `json:"operation,omitempty"`
}

// HandleCalendarServiceError creates a calendar service error
func (h *A2AErrorHandler) HandleCalendarServiceError(operation, calendarID, message string) CalendarServiceError {
	data := map[string]interface{}{
		"operation":  operation,
		"calendarId": calendarID,
		"timestamp":  generateUniqueID(), // Using our existing ID generator for timestamp
	}

	return CalendarServiceError{
		Code:       -32004, // Custom error code for calendar service errors
		Message:    message,
		Data:       &data,
		CalendarID: calendarID,
		Operation:  operation,
	}
}
