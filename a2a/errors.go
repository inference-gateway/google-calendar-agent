package a2a

import (
	"encoding/json"
)

// ErrorCode represents A2A and JSON-RPC error codes
type ErrorCode int

// Error code constants following JSON-RPC 2.0 and A2A specifications
const (
	// A2A specific error codes
	ErrorCodeTaskNotFound            ErrorCode = -32000
	ErrorCodeContentTypeNotSupported ErrorCode = -32001
	ErrorCodeCalendarService         ErrorCode = -32004

	// JSON-RPC standard error codes
	ErrorCodeInvalidParams ErrorCode = -32602
	ErrorCodeInternalError ErrorCode = -32603
)

// String returns a human-readable description of the error code
func (e ErrorCode) String() string {
	switch e {
	case ErrorCodeTaskNotFound:
		return "Task not found"
	case ErrorCodeContentTypeNotSupported:
		return "Content type not supported"
	case ErrorCodeCalendarService:
		return "Calendar service error"
	case ErrorCodeInvalidParams:
		return "Invalid method parameter(s)"
	case ErrorCodeInternalError:
		return "Internal JSON-RPC error"
	default:
		return "Unknown error code"
	}
}

// IsA2AError returns true if the error code is A2A-specific (not standard JSON-RPC)
func (e ErrorCode) IsA2AError() bool {
	switch e {
	case ErrorCodeTaskNotFound, ErrorCodeContentTypeNotSupported, ErrorCodeCalendarService:
		return true
	default:
		return false
	}
}

// IsJSONRPCError returns true if the error code is a standard JSON-RPC error
func (e ErrorCode) IsJSONRPCError() bool {
	switch e {
	case ErrorCodeInvalidParams, ErrorCodeInternalError:
		return true
	default:
		return false
	}
}

// A2AErrorHandler provides utilities for handling errors using A2A types
type A2AErrorHandler struct{}

// NewA2AErrorHandler creates a new error handler
func NewA2AErrorHandler() *A2AErrorHandler {
	return &A2AErrorHandler{}
}

// HandleTaskNotFound creates a task not found error response
func (h *A2AErrorHandler) HandleTaskNotFound(taskID string) TaskNotFoundError {
	data := interface{}(map[string]interface{}{"taskId": taskID})
	return TaskNotFoundError{
		Code:    int(ErrorCodeTaskNotFound),
		Message: "The requested task was not found",
		Data:    &data,
	}
}

// HandleContentTypeNotSupported creates a content type not supported error
func (h *A2AErrorHandler) HandleContentTypeNotSupported(contentType string) ContentTypeNotSupportedError {
	data := interface{}(map[string]interface{}{"contentType": contentType})
	return ContentTypeNotSupportedError{
		Code:    int(ErrorCodeContentTypeNotSupported),
		Message: "The requested content type is not supported by this agent",
		Data:    &data,
	}
}

// HandleInvalidParams creates an invalid params error
func (h *A2AErrorHandler) HandleInvalidParams(message string, params map[string]interface{}) InvalidParamsError {
	data := interface{}(params)
	return InvalidParamsError{
		Code:    int(ErrorCodeInvalidParams),
		Message: message,
		Data:    &data,
	}
}

// HandleInternalError creates an internal error response
func (h *A2AErrorHandler) HandleInternalError(message string, details map[string]interface{}) InternalError {
	data := interface{}(details)
	return InternalError{
		Code:    int(ErrorCodeInternalError),
		Message: message,
		Data:    &data,
	}
}

// CreateErrorResponse creates a JSON-RPC error response
func (h *A2AErrorHandler) CreateErrorResponse(id interface{}, error interface{}) JSONRPCErrorResponse {
	return JSONRPCErrorResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   error,
	}
}

// CreateErrorTask creates a task with error status
func (h *A2AErrorHandler) CreateErrorTask(taskID, contextID, errorMessage string) Task {
	errorMsg := CreateErrorMessage(taskID, errorMessage)
	status := CreateTaskStatus(TaskStateFailed, &errorMsg)

	return Task{
		ID:        taskID,
		ContextID: contextID,
		Kind:      "task",
		Status:    status,
		History:   []Message{errorMsg},
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
		"timestamp":  generateUniqueID(),
	}

	return CalendarServiceError{
		Code:       int(ErrorCodeCalendarService),
		Message:    message,
		Data:       &data,
		CalendarID: calendarID,
		Operation:  operation,
	}
}
