package google_mocks

import (
	"time"

	"github.com/google/uuid"
	"google.golang.org/api/calendar/v3"
)

// MockCalendarService provides a mock implementation for testing
type MockCalendarService struct{}

func (m *MockCalendarService) ListEvents(calendarID string, timeMin, timeMax time.Time) ([]*calendar.Event, error) {
	return []*calendar.Event{}, nil
}

func (m *MockCalendarService) CreateEvent(calendarID string, event *calendar.Event) (*calendar.Event, error) {
	event.Id = uuid.New().String()
	return event, nil
}

func (m *MockCalendarService) UpdateEvent(calendarID, eventID string, event *calendar.Event) (*calendar.Event, error) {
	return event, nil
}

func (m *MockCalendarService) DeleteEvent(calendarID, eventID string) error {
	return nil
}

func (m *MockCalendarService) GetEvent(calendarID, eventID string) (*calendar.Event, error) {
	return &calendar.Event{Id: eventID}, nil
}

func (m *MockCalendarService) ListCalendars() ([]*calendar.CalendarListEntry, error) {
	return []*calendar.CalendarListEntry{
		{
			Id:      "primary",
			Summary: "Primary Calendar",
		},
		{
			Id:      "test@example.com",
			Summary: "Test Calendar",
		},
	}, nil
}
