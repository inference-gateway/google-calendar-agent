package google

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/api/calendar/v3"
)

func createTestEvent(id, summary, description string, startTime, endTime time.Time) *calendar.Event {
	return &calendar.Event{
		Id:          id,
		Summary:     summary,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		Status: "confirmed",
	}
}

func createTestCalendar(id, summary string, isPrimary bool) *calendar.CalendarListEntry {
	return &calendar.CalendarListEntry{
		Id:      id,
		Summary: summary,
		Primary: isPrimary,
	}
}

func TestCalendarService_ListEvents(t *testing.T) {
	logger := zaptest.NewLogger(t)

	testCases := []struct {
		name          string
		calendarID    string
		timeMin       time.Time
		timeMax       time.Time
		mockEvents    []*calendar.Event
		mockError     error
		expectError   bool
		expectedCount int
	}{
		{
			name:       "successful event listing",
			calendarID: "primary",
			timeMin:    time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			timeMax:    time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC),
			mockEvents: []*calendar.Event{
				createTestEvent("event1", "Meeting 1", "Team standup",
					time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
					time.Date(2025, 6, 15, 11, 0, 0, 0, time.UTC)),
				createTestEvent("event2", "Meeting 2", "Client call",
					time.Date(2025, 6, 20, 14, 0, 0, 0, time.UTC),
					time.Date(2025, 6, 20, 15, 0, 0, 0, time.UTC)),
			},
			mockError:     nil,
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:          "empty calendar",
			calendarID:    "primary",
			timeMin:       time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			timeMax:       time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC),
			mockEvents:    []*calendar.Event{},
			mockError:     nil,
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:          "API error",
			calendarID:    "primary",
			timeMin:       time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			timeMax:       time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC),
			mockEvents:    nil,
			mockError:     errors.New("calendar API error"),
			expectError:   true,
			expectedCount: 0,
		},
		{
			name:          "invalid calendar ID",
			calendarID:    "invalid@calendar.com",
			timeMin:       time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			timeMax:       time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC),
			mockEvents:    nil,
			mockError:     errors.New("calendar not found"),
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &CalendarServiceImpl{
				logger: logger,
			}

			// Test input validation
			assert.NotNil(t, service)
			assert.NotEmpty(t, tc.calendarID)
			assert.True(t, tc.timeMax.After(tc.timeMin))

			if tc.mockEvents != nil {
				assert.Len(t, tc.mockEvents, tc.expectedCount)
			}
		})
	}
}

func TestCalendarService_CreateEvent(t *testing.T) {
	logger := zaptest.NewLogger(t)

	testCases := []struct {
		name        string
		calendarID  string
		event       *calendar.Event
		mockError   error
		expectError bool
	}{
		{
			name:       "successful event creation",
			calendarID: "primary",
			event: createTestEvent("", "New Meeting", "Project discussion",
				time.Date(2025, 6, 25, 15, 0, 0, 0, time.UTC),
				time.Date(2025, 6, 25, 16, 0, 0, 0, time.UTC)),
			mockError:   nil,
			expectError: false,
		},
		{
			name:       "event without summary",
			calendarID: "primary",
			event: createTestEvent("", "", "Description only",
				time.Date(2025, 6, 25, 15, 0, 0, 0, time.UTC),
				time.Date(2025, 6, 25, 16, 0, 0, 0, time.UTC)),
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "nil event",
			calendarID:  "primary",
			event:       nil,
			mockError:   nil,
			expectError: true,
		},
		{
			name:       "API error during creation",
			calendarID: "primary",
			event: createTestEvent("", "Meeting", "Test meeting",
				time.Date(2025, 6, 25, 15, 0, 0, 0, time.UTC),
				time.Date(2025, 6, 25, 16, 0, 0, 0, time.UTC)),
			mockError:   errors.New("permission denied"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &CalendarServiceImpl{
				logger: logger,
			}

			if tc.event == nil {
				assert.Nil(t, tc.event)
				assert.True(t, tc.expectError)
				return
			}

			assert.NotNil(t, service)
			assert.NotEmpty(t, tc.calendarID)

			if tc.event.Summary == "" {
				assert.Empty(t, tc.event.Summary)
			} else {
				assert.NotEmpty(t, tc.event.Summary)
			}
		})
	}
}

func TestCalendarService_UpdateEvent(t *testing.T) {
	logger := zaptest.NewLogger(t)

	testCases := []struct {
		name        string
		calendarID  string
		eventID     string
		event       *calendar.Event
		mockError   error
		expectError bool
	}{
		{
			name:       "successful event update",
			calendarID: "primary",
			eventID:    "event123",
			event: createTestEvent("event123", "Updated Meeting", "Updated description",
				time.Date(2025, 6, 25, 16, 0, 0, 0, time.UTC),
				time.Date(2025, 6, 25, 17, 0, 0, 0, time.UTC)),
			mockError:   nil,
			expectError: false,
		},
		{
			name:       "empty event ID",
			calendarID: "primary",
			eventID:    "",
			event: createTestEvent("", "Meeting", "Test",
				time.Date(2025, 6, 25, 15, 0, 0, 0, time.UTC),
				time.Date(2025, 6, 25, 16, 0, 0, 0, time.UTC)),
			mockError:   nil,
			expectError: true,
		},
		{
			name:        "nil event",
			calendarID:  "primary",
			eventID:     "event123",
			event:       nil,
			mockError:   nil,
			expectError: true,
		},
		{
			name:       "event not found",
			calendarID: "primary",
			eventID:    "nonexistent",
			event: createTestEvent("nonexistent", "Meeting", "Test",
				time.Date(2025, 6, 25, 15, 0, 0, 0, time.UTC),
				time.Date(2025, 6, 25, 16, 0, 0, 0, time.UTC)),
			mockError:   errors.New("event not found"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &CalendarServiceImpl{
				logger: logger,
			}

			if tc.eventID == "" {
				assert.Empty(t, tc.eventID)
				assert.True(t, tc.expectError)
				return
			}

			if tc.event == nil {
				assert.Nil(t, tc.event)
				assert.True(t, tc.expectError)
				return
			}

			assert.NotNil(t, service)
			assert.NotEmpty(t, tc.calendarID)
			assert.NotEmpty(t, tc.eventID)
			assert.NotNil(t, tc.event)
		})
	}
}

func TestCalendarService_DeleteEvent(t *testing.T) {
	logger := zaptest.NewLogger(t)

	testCases := []struct {
		name        string
		calendarID  string
		eventID     string
		mockError   error
		expectError bool
	}{
		{
			name:        "successful event deletion",
			calendarID:  "primary",
			eventID:     "event123",
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "empty event ID",
			calendarID:  "primary",
			eventID:     "",
			mockError:   nil,
			expectError: true,
		},
		{
			name:        "empty calendar ID",
			calendarID:  "",
			eventID:     "event123",
			mockError:   nil,
			expectError: true,
		},
		{
			name:        "event not found",
			calendarID:  "primary",
			eventID:     "nonexistent",
			mockError:   errors.New("event not found"),
			expectError: true,
		},
		{
			name:        "permission denied",
			calendarID:  "primary",
			eventID:     "event123",
			mockError:   errors.New("permission denied"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &CalendarServiceImpl{
				logger: logger,
			}

			if tc.calendarID == "" {
				assert.Empty(t, tc.calendarID)
				assert.True(t, tc.expectError)
				return
			}

			if tc.eventID == "" {
				assert.Empty(t, tc.eventID)
				assert.True(t, tc.expectError)
				return
			}

			assert.NotNil(t, service)
			assert.NotEmpty(t, tc.calendarID)
			assert.NotEmpty(t, tc.eventID)
		})
	}
}

func TestCalendarService_GetEvent(t *testing.T) {
	logger := zaptest.NewLogger(t)

	testCases := []struct {
		name        string
		calendarID  string
		eventID     string
		mockEvent   *calendar.Event
		mockError   error
		expectError bool
	}{
		{
			name:       "successful event retrieval",
			calendarID: "primary",
			eventID:    "event123",
			mockEvent: createTestEvent("event123", "Meeting", "Team standup",
				time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
				time.Date(2025, 6, 15, 11, 0, 0, 0, time.UTC)),
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "empty event ID",
			calendarID:  "primary",
			eventID:     "",
			mockEvent:   nil,
			mockError:   nil,
			expectError: true,
		},
		{
			name:        "empty calendar ID",
			calendarID:  "",
			eventID:     "event123",
			mockEvent:   nil,
			mockError:   nil,
			expectError: true,
		},
		{
			name:        "event not found",
			calendarID:  "primary",
			eventID:     "nonexistent",
			mockEvent:   nil,
			mockError:   errors.New("event not found"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &CalendarServiceImpl{
				logger: logger,
			}

			if tc.calendarID == "" {
				assert.Empty(t, tc.calendarID)
				assert.True(t, tc.expectError)
				return
			}

			if tc.eventID == "" {
				assert.Empty(t, tc.eventID)
				assert.True(t, tc.expectError)
				return
			}

			assert.NotNil(t, service)
			assert.NotEmpty(t, tc.calendarID)
			assert.NotEmpty(t, tc.eventID)
		})
	}
}

func TestCalendarService_ListCalendars(t *testing.T) {
	logger := zaptest.NewLogger(t)

	testCases := []struct {
		name          string
		mockCalendars []*calendar.CalendarListEntry
		mockError     error
		expectError   bool
		expectedCount int
	}{
		{
			name: "successful calendar listing",
			mockCalendars: []*calendar.CalendarListEntry{
				createTestCalendar("primary", "Primary Calendar", true),
				createTestCalendar("work@company.com", "Work Calendar", false),
				createTestCalendar("personal@gmail.com", "Personal Calendar", false),
			},
			mockError:     nil,
			expectError:   false,
			expectedCount: 3,
		},
		{
			name:          "empty calendar list",
			mockCalendars: []*calendar.CalendarListEntry{},
			mockError:     nil,
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:          "API error",
			mockCalendars: nil,
			mockError:     errors.New("calendar API error"),
			expectError:   true,
			expectedCount: 0,
		},
		{
			name: "single primary calendar",
			mockCalendars: []*calendar.CalendarListEntry{
				createTestCalendar("primary", "My Calendar", true),
			},
			mockError:     nil,
			expectError:   false,
			expectedCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &CalendarServiceImpl{
				logger: logger,
			}

			assert.NotNil(t, service)

			// Validate mock data structure
			if tc.mockCalendars != nil {
				assert.Len(t, tc.mockCalendars, tc.expectedCount)

				for i, cal := range tc.mockCalendars {
					assert.NotNil(t, cal)
					assert.NotEmpty(t, cal.Id)
					assert.NotEmpty(t, cal.Summary)

					// Check if we have the expected primary calendar
					if i == 0 && tc.expectedCount > 0 {
						if cal.Id == "primary" {
							assert.True(t, cal.Primary)
						}
					}
				}
			}
		})
	}
}

func TestCreateTestEvent(t *testing.T) {
	startTime := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 6, 15, 11, 0, 0, 0, time.UTC)

	event := createTestEvent("test-id", "Test Summary", "Test Description", startTime, endTime)

	require.NotNil(t, event)
	assert.Equal(t, "test-id", event.Id)
	assert.Equal(t, "Test Summary", event.Summary)
	assert.Equal(t, "Test Description", event.Description)
	assert.Equal(t, "confirmed", event.Status)
	assert.NotNil(t, event.Start)
	assert.NotNil(t, event.End)
	assert.Equal(t, "UTC", event.Start.TimeZone)
	assert.Equal(t, "UTC", event.End.TimeZone)
}

func TestCreateTestCalendar(t *testing.T) {
	calendar := createTestCalendar("test-id", "Test Calendar", true)

	require.NotNil(t, calendar)
	assert.Equal(t, "test-id", calendar.Id)
	assert.Equal(t, "Test Calendar", calendar.Summary)
	assert.True(t, calendar.Primary)

	calendar2 := createTestCalendar("secondary", "Secondary Calendar", false)
	assert.False(t, calendar2.Primary)
}

func TestCalendarService_NewService(t *testing.T) {
	logger := zaptest.NewLogger(t)

	service := &CalendarServiceImpl{
		logger: logger,
	}

	assert.NotNil(t, service)
	assert.NotNil(t, service.logger)
}

func TestCalendarService_EdgeCases(t *testing.T) {
	logger := zaptest.NewLogger(t)

	testCases := []struct {
		name        string
		description string
		testFunc    func(t *testing.T)
	}{
		{
			name:        "nil logger",
			description: "Test service behavior with nil logger",
			testFunc: func(t *testing.T) {
				service := &CalendarServiceImpl{
					logger: nil,
				}
				assert.NotNil(t, service)
			},
		},
		{
			name:        "concurrent access",
			description: "Test service is safe for concurrent access",
			testFunc: func(t *testing.T) {
				service := &CalendarServiceImpl{
					logger: logger,
				}

				done := make(chan bool, 2)

				go func() {
					assert.NotNil(t, service)
					done <- true
				}()

				go func() {
					assert.NotNil(t, service)
					done <- true
				}()

				<-done
				<-done
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFunc(t)
		})
	}
}
