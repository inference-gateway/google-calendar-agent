package a2a

import "testing"

func TestErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected string
	}{
		{
			name:     "TaskNotFound",
			code:     ErrorCodeTaskNotFound,
			expected: "Task not found",
		},
		{
			name:     "ContentTypeNotSupported",
			code:     ErrorCodeContentTypeNotSupported,
			expected: "Content type not supported",
		},
		{
			name:     "CalendarService",
			code:     ErrorCodeCalendarService,
			expected: "Calendar service error",
		},
		{
			name:     "InvalidParams",
			code:     ErrorCodeInvalidParams,
			expected: "Invalid method parameter(s)",
		},
		{
			name:     "InternalError",
			code:     ErrorCodeInternalError,
			expected: "Internal JSON-RPC error",
		},
		{
			name:     "UnknownError",
			code:     ErrorCode(-99999),
			expected: "Unknown error code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.code.String(); got != tt.expected {
				t.Errorf("ErrorCode.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorCodeClassification(t *testing.T) {
	tests := []struct {
		name            string
		code            ErrorCode
		expectedA2A     bool
		expectedJSONRPC bool
	}{
		{
			name:            "TaskNotFound is A2A error",
			code:            ErrorCodeTaskNotFound,
			expectedA2A:     true,
			expectedJSONRPC: false,
		},
		{
			name:            "ContentTypeNotSupported is A2A error",
			code:            ErrorCodeContentTypeNotSupported,
			expectedA2A:     true,
			expectedJSONRPC: false,
		},
		{
			name:            "CalendarService is A2A error",
			code:            ErrorCodeCalendarService,
			expectedA2A:     true,
			expectedJSONRPC: false,
		},
		{
			name:            "InvalidParams is JSON-RPC error",
			code:            ErrorCodeInvalidParams,
			expectedA2A:     false,
			expectedJSONRPC: true,
		},
		{
			name:            "InternalError is JSON-RPC error",
			code:            ErrorCodeInternalError,
			expectedA2A:     false,
			expectedJSONRPC: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.code.IsA2AError(); got != tt.expectedA2A {
				t.Errorf("ErrorCode.IsA2AError() = %v, want %v", got, tt.expectedA2A)
			}
			if got := tt.code.IsJSONRPCError(); got != tt.expectedJSONRPC {
				t.Errorf("ErrorCode.IsJSONRPCError() = %v, want %v", got, tt.expectedJSONRPC)
			}
		})
	}
}

func TestA2AErrorHandlerWithEnum(t *testing.T) {
	handler := NewA2AErrorHandler()

	taskErr := handler.HandleTaskNotFound("test-task-id")
	if taskErr.Code != int(ErrorCodeTaskNotFound) {
		t.Errorf("Expected task not found error code %d, got %d", int(ErrorCodeTaskNotFound), taskErr.Code)
	}

	contentErr := handler.HandleContentTypeNotSupported("application/xml")
	if contentErr.Code != int(ErrorCodeContentTypeNotSupported) {
		t.Errorf("Expected content type not supported error code %d, got %d", int(ErrorCodeContentTypeNotSupported), contentErr.Code)
	}

	invalidErr := handler.HandleInvalidParams("Invalid parameter", map[string]interface{}{"param": "value"})
	if invalidErr.Code != int(ErrorCodeInvalidParams) {
		t.Errorf("Expected invalid params error code %d, got %d", int(ErrorCodeInvalidParams), invalidErr.Code)
	}

	internalErr := handler.HandleInternalError("Internal error occurred", map[string]interface{}{"detail": "error"})
	if internalErr.Code != int(ErrorCodeInternalError) {
		t.Errorf("Expected internal error code %d, got %d", int(ErrorCodeInternalError), internalErr.Code)
	}

	calErr := handler.HandleCalendarServiceError("create", "cal-123", "Calendar unavailable")
	if calErr.Code != int(ErrorCodeCalendarService) {
		t.Errorf("Expected calendar service error code %d, got %d", int(ErrorCodeCalendarService), calErr.Code)
	}
}
