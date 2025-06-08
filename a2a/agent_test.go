package a2a

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/api/calendar/v3"

	"github.com/inference-gateway/google-calendar-agent/google/mocks"
)

func setupTestCalendarAgent(t *testing.T) (*CalendarAgent, *mocks.FakeCalendarService) {
	logger := zaptest.NewLogger(t)
	mockService := &mocks.FakeCalendarService{}
	agent := NewCalendarAgent(mockService, logger)
	return agent, mockService
}

// Helper function to parse error response
func parseErrorResponse(t *testing.T, responseBody []byte) (int, string) {
	var response map[string]interface{}
	err := json.Unmarshal(responseBody, &response)
	require.NoError(t, err)

	errorField, ok := response["error"].(map[string]interface{})
	require.True(t, ok, "Response should contain error field")

	code, ok := errorField["code"].(float64)
	require.True(t, ok, "Error should contain code field")

	message, ok := errorField["message"].(string)
	require.True(t, ok, "Error should contain message field")

	return int(code), message
}

func createTestJSONRPCRequest(method string, params map[string]interface{}) JSONRPCRequest {
	id := uuid.New().String()
	return JSONRPCRequest{
		ID:      id,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

func createMessageSendParams(text string) map[string]interface{} {
	return map[string]interface{}{
		"message": map[string]interface{}{
			"parts": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": text,
				},
			},
		},
	}
}

func TestNewCalendarAgent(t *testing.T) {
	testCases := []struct {
		name         string
		expectNonNil bool
	}{
		{
			name:         "successful agent creation",
			expectNonNil: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mocks.FakeCalendarService{}
			logger := zaptest.NewLogger(t)

			agent := NewCalendarAgent(mockService, logger)
			if tc.expectNonNil {
				assert.NotNil(t, agent)
				assert.NotNil(t, agent.calendarService)
				assert.NotNil(t, agent.logger)
			}
		})
	}
}

func TestCalendarAgent_HandleA2ARequest_InvalidJSON(t *testing.T) {
	testCases := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "invalid json syntax",
			requestBody:    `{"invalid": json}`,
			expectedStatus: http.StatusOK,
			expectedError:  "parse error",
		},
		{
			name:           "empty request body",
			requestBody:    "",
			expectedStatus: http.StatusOK,
			expectedError:  "parse error",
		},
		{
			name:           "non-json content",
			requestBody:    "not json at all",
			expectedStatus: http.StatusOK,
			expectedError:  "parse error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agent, _ := setupTestCalendarAgent(t)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/", agent.HandleA2ARequest)

			req, err := http.NewRequest("POST", "/", bytes.NewBufferString(tc.requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			_, message := parseErrorResponse(t, w.Body.Bytes())
			assert.Contains(t, message, tc.expectedError)
		})
	}
}

func TestCalendarAgent_HandleA2ARequest_MethodNotFound(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		expectedStatus int
		expectedCode   int
	}{
		{
			name:           "unknown method",
			method:         "unknown/method",
			expectedStatus: http.StatusOK,
			expectedCode:   -32601,
		},
		{
			name:           "empty method",
			method:         "",
			expectedStatus: http.StatusOK,
			expectedCode:   -32601,
		},
		{
			name:           "invalid method format",
			method:         "invalid_format",
			expectedStatus: http.StatusOK,
			expectedCode:   -32601,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agent, _ := setupTestCalendarAgent(t)

			request := createTestJSONRPCRequest(tc.method, make(map[string]interface{}))
			requestBody, err := json.Marshal(request)
			require.NoError(t, err)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/", agent.HandleA2ARequest)

			req, err := http.NewRequest("POST", "/", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			code, message := parseErrorResponse(t, w.Body.Bytes())
			assert.Equal(t, tc.expectedCode, code)
			assert.Contains(t, message, "method not found")
		})
	}
}

func TestCalendarAgent_HandleMessageSend_Success(t *testing.T) {
	testCases := []struct {
		name           string
		messageText    string
		mockEvents     []*calendar.Event
		mockCalendars  []*calendar.CalendarListEntry
		mockError      error
		expectedStatus int
		expectedState  string
	}{
		{
			name:        "list events request",
			messageText: "show my events today",
			mockEvents: []*calendar.Event{
				{
					Id:      "event1",
					Summary: "Test Meeting",
					Start: &calendar.EventDateTime{
						DateTime: time.Now().Format(time.RFC3339),
					},
					End: &calendar.EventDateTime{
						DateTime: time.Now().Add(time.Hour).Format(time.RFC3339),
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedState:  "completed",
		},
		{
			name:        "list calendars request",
			messageText: "show my calendars",
			mockCalendars: []*calendar.CalendarListEntry{
				{
					Id:      "primary",
					Summary: "Primary Calendar",
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedState:  "completed",
		},
		{
			name:           "create event request",
			messageText:    "schedule a meeting tomorrow at 2pm",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedState:  "completed",
		},
		{
			name:           "help request",
			messageText:    "what can you do?",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedState:  "completed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agent, mockService := setupTestCalendarAgent(t)

			if tc.mockEvents != nil {
				mockService.ListEventsReturns(tc.mockEvents, tc.mockError)
			}
			if tc.mockCalendars != nil {
				mockService.ListCalendarsReturns(tc.mockCalendars, tc.mockError)
			}

			if strings.Contains(tc.messageText, "schedule") || strings.Contains(tc.messageText, "create") {
				mockEvent := &calendar.Event{
					Id:      "test-event-id",
					Summary: "Meeting",
					Start: &calendar.EventDateTime{
						DateTime: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
					},
					End: &calendar.EventDateTime{
						DateTime: time.Now().Add(25 * time.Hour).Format(time.RFC3339),
					},
				}
				mockService.CreateEventReturns(mockEvent, tc.mockError)
			}

			params := createMessageSendParams(tc.messageText)
			request := createTestJSONRPCRequest("message/send", params)
			requestBody, err := json.Marshal(request)
			require.NoError(t, err)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/", agent.HandleA2ARequest)

			req, err := http.NewRequest("POST", "/", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var response JSONRPCSuccessResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			task, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "Response result should be a task object")

			status, ok := task["status"].(map[string]interface{})
			require.True(t, ok, "Task should have status")

			state, ok := status["state"].(string)
			require.True(t, ok, "Status should have state")
			assert.Equal(t, tc.expectedState, state)
		})
	}
}

func TestCalendarAgent_HandleMessageSend_InvalidParams(t *testing.T) {
	testCases := []struct {
		name           string
		params         map[string]interface{}
		expectedStatus int
		expectedCode   int
		expectedError  string
	}{
		{
			name:           "missing message param",
			params:         map[string]interface{}{},
			expectedStatus: http.StatusOK,
			expectedCode:   -32602,
			expectedError:  "invalid params: missing message",
		},
		{
			name: "invalid message type",
			params: map[string]interface{}{
				"message": "not an object",
			},
			expectedStatus: http.StatusOK,
			expectedCode:   -32602,
			expectedError:  "invalid params: missing message",
		},
		{
			name: "missing parts",
			params: map[string]interface{}{
				"message": map[string]interface{}{
					"other": "data",
				},
			},
			expectedStatus: http.StatusOK,
			expectedCode:   -32602,
			expectedError:  "invalid params: missing message parts",
		},
		{
			name: "invalid parts type",
			params: map[string]interface{}{
				"message": map[string]interface{}{
					"parts": "not an array",
				},
			},
			expectedStatus: http.StatusOK,
			expectedCode:   -32602,
			expectedError:  "invalid params: missing message parts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agent, _ := setupTestCalendarAgent(t)

			request := createTestJSONRPCRequest("message/send", tc.params)
			requestBody, err := json.Marshal(request)
			require.NoError(t, err)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/", agent.HandleA2ARequest)

			req, err := http.NewRequest("POST", "/", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			code, message := parseErrorResponse(t, w.Body.Bytes())
			assert.Equal(t, tc.expectedCode, code)
			assert.Contains(t, message, tc.expectedError)
		})
	}
}

func TestCalendarAgent_HandleMessageStream(t *testing.T) {
	agent, mockService := setupTestCalendarAgent(t)

	params := createMessageSendParams("help")
	request := createTestJSONRPCRequest("message/stream", params)
	requestBody, err := json.Marshal(request)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/", agent.HandleA2ARequest)

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(requestBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response JSONRPCSuccessResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 0, mockService.ListEventsCallCount())
}

func TestCalendarAgent_HandleTaskGet(t *testing.T) {
	agent, _ := setupTestCalendarAgent(t)

	request := createTestJSONRPCRequest("task/get", map[string]interface{}{
		"taskId": "test-task-id",
	})
	requestBody, err := json.Marshal(request)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/", agent.HandleA2ARequest)

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(requestBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	code, message := parseErrorResponse(t, w.Body.Bytes())
	assert.Equal(t, -32601, code)
	assert.Contains(t, message, "task/get not implemented")
}

func TestCalendarAgent_HandleTaskCancel(t *testing.T) {
	agent, _ := setupTestCalendarAgent(t)

	request := createTestJSONRPCRequest("task/cancel", map[string]interface{}{
		"taskId": "test-task-id",
	})
	requestBody, err := json.Marshal(request)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/", agent.HandleA2ARequest)

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(requestBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	code, message := parseErrorResponse(t, w.Body.Bytes())
	assert.Equal(t, -32601, code)
	assert.Contains(t, message, "task/cancel not implemented")
}

func TestCalendarAgent_ProcessCalendarRequest_ListEvents(t *testing.T) {
	testCases := []struct {
		name         string
		messageText  string
		mockEvents   []*calendar.Event
		mockError    error
		expectError  bool
		expectedText string
	}{
		{
			name:        "list today events",
			messageText: "show my events today",
			mockEvents: []*calendar.Event{
				{
					Id:      "event1",
					Summary: "Morning Meeting",
					Start: &calendar.EventDateTime{
						DateTime: time.Now().Format(time.RFC3339),
					},
					End: &calendar.EventDateTime{
						DateTime: time.Now().Add(time.Hour).Format(time.RFC3339),
					},
				},
			},
			mockError:    nil,
			expectError:  false,
			expectedText: "Here are your events for today:",
		},
		{
			name:         "no events found",
			messageText:  "show my events today",
			mockEvents:   []*calendar.Event{},
			mockError:    nil,
			expectError:  false,
			expectedText: "No events found for today.",
		},
		{
			name:         "calendar service error",
			messageText:  "show my events today",
			mockEvents:   nil,
			mockError:    fmt.Errorf("calendar API error"),
			expectError:  true,
			expectedText: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agent, mockService := setupTestCalendarAgent(t)

			mockService.ListEventsReturns(tc.mockEvents, tc.mockError)

			response, err := agent.processCalendarRequest(tc.messageText)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Contains(t, response.Text, tc.expectedText)

			if len(tc.mockEvents) > 0 {
				assert.NotNil(t, response.Data)
			}
		})
	}
}

func TestCalendarAgent_ProcessCalendarRequest_ListCalendars(t *testing.T) {
	testCases := []struct {
		name          string
		messageText   string
		mockCalendars []*calendar.CalendarListEntry
		mockError     error
		expectError   bool
		expectedText  string
	}{
		{
			name:        "list calendars success",
			messageText: "show my calendars",
			mockCalendars: []*calendar.CalendarListEntry{
				{
					Id:      "primary",
					Summary: "Primary Calendar",
				},
			},
			mockError:    nil,
			expectError:  false,
			expectedText: "Here are your available calendars:",
		},
		{
			name:          "calendar service error",
			messageText:   "show my calendars",
			mockCalendars: nil,
			mockError:     fmt.Errorf("calendar API error"),
			expectError:   true,
			expectedText:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agent, mockService := setupTestCalendarAgent(t)

			mockService.ListCalendarsReturns(tc.mockCalendars, tc.mockError)

			response, err := agent.processCalendarRequest(tc.messageText)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Contains(t, response.Text, tc.expectedText)
		})
	}
}

func TestCalendarAgent_RequestTypeDetection(t *testing.T) {
	agent, _ := setupTestCalendarAgent(t)

	testCases := []struct {
		name         string
		text         string
		expectedType string
	}{
		{
			name:         "show my events",
			text:         "show my events today",
			expectedType: "list-events",
		},
		{
			name:         "what's on today",
			text:         "what's on today",
			expectedType: "list-events",
		},
		{
			name:         "time only - today",
			text:         "today",
			expectedType: "list-events",
		},
		{
			name:         "time only - tomorrow",
			text:         "tomorrow",
			expectedType: "list-events",
		},
		{
			name:         "list calendars",
			text:         "list my calendars",
			expectedType: "list-calendars",
		},
		{
			name:         "what calendars",
			text:         "what calendars do I have",
			expectedType: "list-calendars",
		},
		{
			name:         "schedule meeting",
			text:         "schedule a meeting with John at 2pm tomorrow",
			expectedType: "create-event",
		},
		{
			name:         "book appointment",
			text:         "book an appointment",
			expectedType: "create-event",
		},
		{
			name:         "move meeting",
			text:         "move my 3pm meeting to 4pm",
			expectedType: "update-event",
		},
		{
			name:         "reschedule appointment",
			text:         "reschedule my appointment",
			expectedType: "update-event",
		},
		{
			name:         "cancel meeting",
			text:         "cancel my meeting",
			expectedType: "delete-event",
		},
		{
			name:         "delete appointment",
			text:         "delete my appointment",
			expectedType: "delete-event",
		},
		{
			name:         "unknown request",
			text:         "what can you do?",
			expectedType: "help",
		},
		{
			name:         "empty text",
			text:         "",
			expectedType: "help",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normalizedText := strings.ToLower(strings.TrimSpace(tc.text))

			var actualType string
			switch {
			case agent.isListCalendarsRequest(normalizedText):
				actualType = "list-calendars"
			case agent.isListEventsRequest(normalizedText):
				actualType = "list-events"
			case agent.isUpdateEventRequest(normalizedText):
				actualType = "update-event"
			case agent.isDeleteEventRequest(normalizedText):
				actualType = "delete-event"
			case agent.isCreateEventRequest(normalizedText):
				actualType = "create-event"
			default:
				actualType = "help"
			}

			assert.Equal(t, tc.expectedType, actualType)
		})
	}
}

func TestCalendarAgent_ParseEventDetails(t *testing.T) {
	agent, _ := setupTestCalendarAgent(t)

	testCases := []struct {
		name            string
		text            string
		expectedTitle   string
		expectValidTime bool
		expectValidDate bool
	}{
		{
			name:            "meeting with title and time",
			text:            `schedule a meeting "Weekly Standup" at 2pm tomorrow`,
			expectedTitle:   "Weekly Standup",
			expectValidTime: true,
			expectValidDate: true,
		},
		{
			name:            "simple time reference",
			text:            "meeting at 3pm",
			expectedTitle:   "Meeting",
			expectValidTime: true,
			expectValidDate: false,
		},
		{
			name:            "date reference only",
			text:            "meeting tomorrow",
			expectedTitle:   "Meeting",
			expectValidTime: false,
			expectValidDate: true,
		},
		{
			name:            "no time or date",
			text:            "schedule a meeting",
			expectedTitle:   "Meeting",
			expectValidTime: false,
			expectValidDate: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			details := agent.parseEventDetails(tc.text)

			assert.Equal(t, tc.expectedTitle, details.Title)

			if tc.expectValidTime {
				assert.False(t, details.StartTime.IsZero())
				assert.False(t, details.EndTime.IsZero())
				assert.True(t, details.EndTime.After(details.StartTime))
			}

			if tc.expectValidDate {
				assert.False(t, details.StartTime.IsZero())
			}
		})
	}
}

func TestCalendarAgent_TimeAndDateParsing(t *testing.T) {
	agent, _ := setupTestCalendarAgent(t)

	timeTestCases := []struct {
		name        string
		timeStr     string
		expectError bool
	}{
		{name: "2pm", timeStr: "2pm", expectError: false},
		{name: "14:30", timeStr: "14:30", expectError: false},
		{name: "3 PM", timeStr: "3 PM", expectError: false},
		{name: "invalid", timeStr: "invalid", expectError: true},
		{name: "empty", timeStr: "", expectError: true},
	}

	for _, tc := range timeTestCases {
		t.Run("time_"+tc.name, func(t *testing.T) {
			result, err := agent.parseTime(tc.timeStr)

			if tc.expectError {
				assert.Error(t, err)
				assert.True(t, result.IsZero())
			} else {
				assert.NoError(t, err)
				assert.False(t, result.IsZero())
			}
		})
	}

	dateTestCases := []struct {
		name        string
		dateStr     string
		expectError bool
	}{
		{name: "tomorrow", dateStr: "tomorrow", expectError: false},
		{name: "monday", dateStr: "monday", expectError: false},
		{name: "next friday", dateStr: "next friday", expectError: false},
		{name: "invalid", dateStr: "invalid", expectError: true},
		{name: "empty", dateStr: "", expectError: true},
	}

	for _, tc := range dateTestCases {
		t.Run("date_"+tc.name, func(t *testing.T) {
			result, err := agent.parseDate(tc.dateStr)

			if tc.expectError {
				assert.Error(t, err)
				assert.True(t, result.IsZero())
			} else {
				assert.NoError(t, err)
				assert.False(t, result.IsZero())
			}
		})
	}
}

func TestCalendarAgent_GetStringParam(t *testing.T) {
	agent, _ := setupTestCalendarAgent(t)

	testCases := []struct {
		name         string
		params       map[string]interface{}
		key          string
		defaultValue string
		expected     string
	}{
		{
			name: "existing string param",
			params: map[string]interface{}{
				"key1": "value1",
			},
			key:          "key1",
			defaultValue: "default",
			expected:     "value1",
		},
		{
			name: "non-existent param",
			params: map[string]interface{}{
				"key1": "value1",
			},
			key:          "key2",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name: "non-string param",
			params: map[string]interface{}{
				"key1": 123,
			},
			key:          "key1",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "empty params",
			params:       map[string]interface{}{},
			key:          "key1",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := agent.getStringParam(tc.params, tc.key, tc.defaultValue)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCalendarAgent_GetNextWeekday(t *testing.T) {
	agent, _ := setupTestCalendarAgent(t)

	testDate := time.Date(2025, 6, 9, 10, 0, 0, 0, time.UTC) // Monday

	testCases := []struct {
		name            string
		fromDate        time.Time
		targetWeekday   time.Weekday
		expectedDaysAdd int
	}{
		{
			name:            "monday to tuesday",
			fromDate:        testDate,
			targetWeekday:   time.Tuesday,
			expectedDaysAdd: 1,
		},
		{
			name:            "monday to next monday",
			fromDate:        testDate,
			targetWeekday:   time.Monday,
			expectedDaysAdd: 7,
		},
		{
			name:            "monday to friday",
			fromDate:        testDate,
			targetWeekday:   time.Friday,
			expectedDaysAdd: 4,
		},
		{
			name:            "monday to sunday",
			fromDate:        testDate,
			targetWeekday:   time.Sunday,
			expectedDaysAdd: 6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := agent.getNextWeekday(tc.fromDate, tc.targetWeekday)
			expected := tc.fromDate.Add(time.Duration(tc.expectedDaysAdd) * 24 * time.Hour)

			assert.Equal(t, expected.Year(), result.Year())
			assert.Equal(t, expected.Month(), result.Month())
			assert.Equal(t, expected.Day(), result.Day())
			assert.Equal(t, tc.targetWeekday, result.Weekday())
		})
	}
}
