package google_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/calendar/v3"

	"github.com/inference-gateway/google-calendar-agent/google/mocks"
)

func TestCalendarServiceMocking(t *testing.T) {
	testCases := []struct {
		name        string
		setupMock   func(*mocks.FakeCalendarService)
		expectError bool
		validate    func(*testing.T, []*calendar.Event, error)
	}{
		{
			name: "successful event listing",
			setupMock: func(mock *mocks.FakeCalendarService) {
				events := []*calendar.Event{
					{
						Id:      "event1",
						Summary: "Test Event 1",
						Start: &calendar.EventDateTime{
							DateTime: time.Now().Format(time.RFC3339),
						},
					},
					{
						Id:      "event2",
						Summary: "Test Event 2",
						Start: &calendar.EventDateTime{
							DateTime: time.Now().Add(time.Hour).Format(time.RFC3339),
						},
					},
				}
				mock.ListEventsReturns(events, nil)
			},
			expectError: false,
			validate: func(t *testing.T, events []*calendar.Event, err error) {
				assert.NoError(t, err)
				assert.Len(t, events, 2)
				assert.Equal(t, "event1", events[0].Id)
				assert.Equal(t, "Test Event 1", events[0].Summary)
			},
		},
		{
			name: "API error handling",
			setupMock: func(mock *mocks.FakeCalendarService) {
				mock.ListEventsReturns(nil, errors.New("API quota exceeded"))
			},
			expectError: true,
			validate: func(t *testing.T, events []*calendar.Event, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "API quota exceeded")
				assert.Nil(t, events)
			},
		},
		{
			name: "empty result set",
			setupMock: func(mock *mocks.FakeCalendarService) {
				mock.ListEventsReturns([]*calendar.Event{}, nil)
			},
			expectError: false,
			validate: func(t *testing.T, events []*calendar.Event, err error) {
				assert.NoError(t, err)
				assert.Empty(t, events)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mocks.FakeCalendarService{}
			tc.setupMock(mockService)

			events, err := mockService.ListEvents(
				"primary",
				time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC),
			)

			tc.validate(t, events, err)

			assert.Equal(t, 1, mockService.ListEventsCallCount())
			calendarID, timeMin, timeMax := mockService.ListEventsArgsForCall(0)
			assert.Equal(t, "primary", calendarID)
			assert.Equal(t, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), timeMin)
			assert.Equal(t, time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC), timeMax)
		})
	}
}

func TestCalendarServiceConcurrency(t *testing.T) {
	mockService := &mocks.FakeCalendarService{}

	events1 := []*calendar.Event{{Id: "event1", Summary: "Event 1"}}
	events2 := []*calendar.Event{{Id: "event2", Summary: "Event 2"}}

	mockService.ListEventsReturnsOnCall(0, events1, nil)
	mockService.ListEventsReturnsOnCall(1, events2, nil)

	type result struct {
		events []*calendar.Event
		err    error
	}
	results := make(chan result, 2)

	for i := 0; i < 2; i++ {
		go func() {
			events, err := mockService.ListEvents("primary", time.Now(), time.Now().Add(time.Hour))
			results <- result{events: events, err: err}
		}()
	}

	var allEvents [][]*calendar.Event
	for i := 0; i < 2; i++ {
		select {
		case res := <-results:
			assert.NoError(t, res.err)
			allEvents = append(allEvents, res.events)
		case <-time.After(time.Second):
			t.Fatal("Test timed out")
		}
	}

	// Verify both calls were made
	assert.Equal(t, 2, mockService.ListEventsCallCount())
	assert.Len(t, allEvents, 2)
}

func TestCalendarServiceCreateEventMocking(t *testing.T) {
	testCases := []struct {
		name        string
		inputEvent  *calendar.Event
		mockReturn  *calendar.Event
		mockError   error
		expectError bool
	}{
		{
			name: "successful event creation",
			inputEvent: &calendar.Event{
				Summary:     "New Meeting",
				Description: "Team standup",
			},
			mockReturn: &calendar.Event{
				Id:          "created-event-123",
				Summary:     "New Meeting",
				Description: "Team standup",
				Created:     time.Now().Format(time.RFC3339),
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name: "creation failure",
			inputEvent: &calendar.Event{
				Summary: "Invalid Event",
			},
			mockReturn:  nil,
			mockError:   errors.New("insufficient permissions"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mocks.FakeCalendarService{}
			mockService.CreateEventReturns(tc.mockReturn, tc.mockError)

			result, err := mockService.CreateEvent("primary", tc.inputEvent)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.mockReturn.Id, result.Id)
				assert.Equal(t, tc.mockReturn.Summary, result.Summary)
			}

			assert.Equal(t, 1, mockService.CreateEventCallCount())
			calendarID, event := mockService.CreateEventArgsForCall(0)
			assert.Equal(t, "primary", calendarID)
			assert.Equal(t, tc.inputEvent, event)
		})
	}
}
