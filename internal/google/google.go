package google

import (
	"context"
	"fmt"
	"time"

	config "github.com/inference-gateway/google-calendar-agent/config"
	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"
	option "google.golang.org/api/option"
)

// CalendarService represents the google dependency interface
// Google Calendar API service for managing calendar events
type CalendarService interface {
	ListEvents(calendarID string, timeMin, timeMax time.Time) ([]*calendar.Event, error)
	CreateEvent(calendarID string, event *calendar.Event) (*calendar.Event, error)
	UpdateEvent(calendarID, eventID string, event *calendar.Event) (*calendar.Event, error)
	DeleteEvent(calendarID, eventID string) error
	GetEvent(calendarID, eventID string) (*calendar.Event, error)
	ListCalendars() ([]*calendar.CalendarListEntry, error)
	CheckConflicts(calendarID string, startTime, endTime time.Time) ([]*calendar.Event, error)
	GetCalendarID() string
}

// NewServiceFactory creates a new instance of CalendarService
func NewServiceFactory(logger *zap.Logger, cfg *config.Config) (CalendarService, error) {
	logger.Info("initializing Google Calendar dependency")

	// Check if we should use mock mode based on config
	if shouldUseMockMode(cfg) {
		logger.Info("Creating mock Google Calendar service")
		return NewMockCalendarService(logger, cfg), nil
	}

	logger.Info("Creating real Google Calendar service")
	return createRealCalendarService(context.Background(), logger, cfg)
}

// shouldUseMockMode determines if mock mode should be used based on config
func shouldUseMockMode(cfg *config.Config) bool {
	return cfg.GoogleCalendar.MockMode
}

// createRealCalendarService creates a real Google Calendar service using config
func createRealCalendarService(ctx context.Context, logger *zap.Logger, cfg *config.Config) (CalendarService, error) {
	var opts []option.ClientOption

	if cfg.Google.ServiceAccountJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(cfg.Google.ServiceAccountJSON))) //nolint:staticcheck // credential type is not known up front, migrating needs its own change
	} else if cfg.Google.CredentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.Google.CredentialsPath)) //nolint:staticcheck // credential type is not known up front, migrating needs its own change
	} else {
		return nil, fmt.Errorf("no Google credentials found: set GOOGLE_SERVICE_ACCOUNT_JSON or GOOGLE_CREDENTIALS_PATH")
	}

	scopes := []string{
		calendar.CalendarReadonlyScope,
		calendar.CalendarScope,
	}

	scopesOption := option.WithScopes(scopes...)
	allOptions := append([]option.ClientOption{scopesOption}, opts...)

	svc, err := calendar.NewService(ctx, allOptions...)
	if err != nil {
		return nil, fmt.Errorf("unable to create google calendar service: %w", err)
	}

	return &CalendarServiceImpl{service: svc, logger: logger, config: cfg}, nil
}

// CalendarServiceImpl implements the calendar service interface for Google Calendar API
type CalendarServiceImpl struct {
	service *calendar.Service
	logger  *zap.Logger
	config  *config.Config
}

// GetCalendarID returns the calendar ID from config or default to "primary"
func (g *CalendarServiceImpl) GetCalendarID() string {
	if g.config.GoogleCalendar.ID != "" {
		return g.config.GoogleCalendar.ID
	}
	return "primary"
}

// CreateEvent creates a new event in the calendar
func (g *CalendarServiceImpl) CreateEvent(calendarID string, event *calendar.Event) (*calendar.Event, error) {
	g.logger.Debug("creating event",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "create-event"),
		zap.String("calendarID", calendarID),
		zap.String("summary", event.Summary))

	createdEvent, err := g.service.Events.Insert(calendarID, event).Do()
	if err != nil {
		g.logger.Error("failed to create event",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "create-event"),
			zap.Error(err))
		return nil, fmt.Errorf("unable to create event: %w", err)
	}

	g.logger.Debug("Successfully created event", zap.String("eventId", createdEvent.Id))
	return createdEvent, nil
}

// ListEvents lists the events in the calendar
func (g *CalendarServiceImpl) ListEvents(calendarID string, timeMin, timeMax time.Time) ([]*calendar.Event, error) {
	g.logger.Debug("listing events",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "list-events"),
		zap.String("calendarID", calendarID),
		zap.Time("timeMin", timeMin),
		zap.Time("timeMax", timeMax))

	call := g.service.Events.List(calendarID).
		TimeMin(timeMin.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime")

	var items []*calendar.Event
	var err error
	if timeMax.IsZero() {
		var events *calendar.Events
		events, err = call.Do()
		if events != nil {
			items = events.Items
		}
	} else {
		items, _, err = collectAllEvents(call.TimeMax(timeMax.Format(time.RFC3339)))
	}
	if err != nil {
		g.logger.Error("failed to list events",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "list-events"),
			zap.Error(err))
		return nil, fmt.Errorf("unable to list events: %w", err)
	}

	g.logger.Debug("Successfully listed events", zap.Int("count", len(items)))
	return items, nil
}

// collectAllEvents follows nextPageToken until every page of a bounded listing is read.
// It also returns the calendar time zone reported by the API, used to place all-day events.
func collectAllEvents(call *calendar.EventsListCall) ([]*calendar.Event, string, error) {
	var items []*calendar.Event
	var timeZone string
	err := call.Pages(context.Background(), func(page *calendar.Events) error {
		items = append(items, page.Items...)
		if timeZone == "" {
			timeZone = page.TimeZone
		}
		return nil
	})
	return items, timeZone, err
}

// EventTimeRange returns the start and end instants of an event.
// Timed events use start.dateTime/end.dateTime. All-day events use start.date/end.date
// interpreted as midnight in loc, matching how Google Calendar blocks the whole day.
func EventTimeRange(event *calendar.Event, loc *time.Location) (time.Time, time.Time, bool) {
	if event == nil || event.Start == nil || event.End == nil {
		return time.Time{}, time.Time{}, false
	}
	start, startOK := parseEventDateTime(event.Start, loc)
	end, endOK := parseEventDateTime(event.End, loc)
	return start, end, startOK && endOK
}

// parseEventDateTime converts a single EventDateTime into an instant, preferring dateTime over date.
func parseEventDateTime(edt *calendar.EventDateTime, loc *time.Location) (time.Time, bool) {
	if edt.DateTime != "" {
		t, err := time.Parse(time.RFC3339, edt.DateTime)
		return t, err == nil
	}
	if edt.Date != "" {
		t, err := time.ParseInLocation(DateLayout, edt.Date, loc)
		return t, err == nil
	}
	return time.Time{}, false
}

// EventDateTimeValue returns dateTime for timed events and date for all-day events.
func EventDateTimeValue(edt *calendar.EventDateTime) string {
	if edt == nil {
		return ""
	}
	if edt.DateTime != "" {
		return edt.DateTime
	}
	return edt.Date
}

// IsAllDay reports whether an event spans whole days rather than a timed range.
func IsAllDay(event *calendar.Event) bool {
	return event != nil && event.Start != nil && event.Start.DateTime == "" && event.Start.Date != ""
}

// DateLayout is the YYYY-MM-DD format Google Calendar uses for all-day event dates.
const DateLayout = "2006-01-02"

// UpdateEvent updates an event by ID in the calendar
func (g *CalendarServiceImpl) UpdateEvent(calendarID, eventID string, event *calendar.Event) (*calendar.Event, error) {
	g.logger.Debug("updating event",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "update-event"),
		zap.String("calendarID", calendarID),
		zap.String("eventID", eventID),
		zap.String("summary", event.Summary))

	updatedEvent, err := g.service.Events.Update(calendarID, eventID, event).Do()
	if err != nil {
		g.logger.Error("failed to update event",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "update-event"),
			zap.String("eventID", eventID),
			zap.Error(err))
		return nil, fmt.Errorf("unable to update event: %w", err)
	}

	g.logger.Debug("Successfully updated event", zap.String("eventId", updatedEvent.Id))
	return updatedEvent, nil
}

// DeleteEvent deletes an event by ID
func (g *CalendarServiceImpl) DeleteEvent(calendarID, eventID string) error {
	g.logger.Debug("deleting event",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "delete-event"),
		zap.String("calendarID", calendarID),
		zap.String("eventID", eventID))

	err := g.service.Events.Delete(calendarID, eventID).Do()
	if err != nil {
		g.logger.Error("failed to delete event",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "delete-event"),
			zap.String("eventID", eventID),
			zap.Error(err))
		return fmt.Errorf("unable to delete event: %w", err)
	}

	g.logger.Debug("Successfully deleted event", zap.String("eventId", eventID))
	return nil
}

// GetEvent get a specific event by ID
func (g *CalendarServiceImpl) GetEvent(calendarID, eventID string) (*calendar.Event, error) {
	g.logger.Debug("getting event",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "get-event"),
		zap.String("calendarID", calendarID),
		zap.String("eventID", eventID))

	event, err := g.service.Events.Get(calendarID, eventID).Do()
	if err != nil {
		g.logger.Error("failed to get event",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "get-event"),
			zap.String("eventID", eventID),
			zap.Error(err))
		return nil, fmt.Errorf("unable to get event: %w", err)
	}

	g.logger.Debug("Successfully retrieved event", zap.String("eventId", event.Id))
	return event, nil
}

// ListCalendars returns the list of calendars accessible to the configured credential
func (g *CalendarServiceImpl) ListCalendars() ([]*calendar.CalendarListEntry, error) {
	g.logger.Debug("listing calendars",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "list-calendars"))

	list, err := g.service.CalendarList.List().Do()
	if err != nil {
		g.logger.Error("failed to list calendars",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "list-calendars"),
			zap.Error(err))
		return nil, fmt.Errorf("unable to list calendars: %w", err)
	}

	g.logger.Debug("Successfully listed calendars", zap.Int("count", len(list.Items)))
	return list.Items, nil
}

// CheckConflicts checks for conflicts of events in the calendar by given start and end time
func (g *CalendarServiceImpl) CheckConflicts(calendarID string, startTime, endTime time.Time) ([]*calendar.Event, error) {
	g.logger.Debug("checking conflicts",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "check-conflicts"),
		zap.String("calendarID", calendarID),
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime))

	items, calendarTimeZone, err := collectAllEvents(g.service.Events.List(calendarID).
		TimeMin(startTime.Format(time.RFC3339)).
		TimeMax(endTime.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime"))
	if err != nil {
		g.logger.Error("failed to check conflicts",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "check-conflicts"),
			zap.Error(err))
		return nil, fmt.Errorf("unable to check conflicts: %w", err)
	}

	conflicts := overlappingEvents(items, startTime, endTime, locationOrDefault(calendarTimeZone, startTime.Location()))

	g.logger.Debug("Successfully checked conflicts", zap.Int("conflictCount", len(conflicts)))
	return conflicts, nil
}

// overlappingEvents returns the events whose time range intersects [startTime, endTime).
// All-day events are placed in loc.
func overlappingEvents(events []*calendar.Event, startTime, endTime time.Time, loc *time.Location) []*calendar.Event {
	var conflicts []*calendar.Event
	for _, event := range events {
		eventStart, eventEnd, ok := EventTimeRange(event, loc)
		if ok && eventStart.Before(endTime) && eventEnd.After(startTime) {
			conflicts = append(conflicts, event)
		}
	}
	return conflicts
}

// locationOrDefault loads the named IANA time zone, falling back when the name is empty or unknown.
func locationOrDefault(name string, fallback *time.Location) *time.Location {
	if name == "" {
		return fallback
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return fallback
	}
	return loc
}

// MockCalendarService implements CalendarService for testing
type MockCalendarService struct {
	logger *zap.Logger
	config *config.Config
}

// NewMockCalendarService creates a new mock calendar service
func NewMockCalendarService(logger *zap.Logger, cfg *config.Config) *MockCalendarService {
	return &MockCalendarService{
		logger: logger,
		config: cfg,
	}
}

// GetCalendarID returns the calendar ID from config or default to "primary"
func (m *MockCalendarService) GetCalendarID() string {
	if m.config.GoogleCalendar.ID != "" {
		return m.config.GoogleCalendar.ID
	}
	return "primary"
}

// CreateEvent creates a mock event
func (m *MockCalendarService) CreateEvent(calendarID string, event *calendar.Event) (*calendar.Event, error) {
	m.logger.Debug("Mock: creating event", zap.String("summary", event.Summary))

	event.Id = fmt.Sprintf("mock-event-%d", time.Now().Unix())
	event.Status = "confirmed"

	return event, nil
}

// Implement other mock methods...
func (m *MockCalendarService) ListEvents(calendarID string, timeMin, timeMax time.Time) ([]*calendar.Event, error) {
	m.logger.Debug("Mock: listing events", zap.String("calendarID", calendarID))

	events := []*calendar.Event{
		{
			Id:          "mock-event-1",
			Summary:     "Mock Meeting 1",
			Description: "This is a mock meeting",
			Status:      "confirmed",
			Start: &calendar.EventDateTime{
				DateTime: timeMin.Add(time.Hour).Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: timeMin.Add(2 * time.Hour).Format(time.RFC3339),
			},
			Location: "Mock Office",
			HtmlLink: "https://calendar.google.com/mock1",
		},
		{
			Id:          "mock-event-2",
			Summary:     "Mock Event 2",
			Description: "Another mock event",
			Status:      "confirmed",
			Start: &calendar.EventDateTime{
				DateTime: timeMin.Add(3 * time.Hour).Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: timeMin.Add(4 * time.Hour).Format(time.RFC3339),
			},
			Location: "Mock Location 2",
			HtmlLink: "https://calendar.google.com/mock2",
		},
	}

	return events, nil
}
func (m *MockCalendarService) UpdateEvent(calendarID, eventID string, event *calendar.Event) (*calendar.Event, error) {
	m.logger.Debug("Mock: updating event", zap.String("eventId", eventID), zap.String("summary", event.Summary))

	event.Id = eventID
	event.Status = "confirmed"

	return event, nil
}
func (m *MockCalendarService) DeleteEvent(calendarID, eventID string) error {
	return nil
}
func (m *MockCalendarService) GetEvent(calendarID, eventID string) (*calendar.Event, error) {
	m.logger.Debug("Mock: getting event", zap.String("eventId", eventID))

	return &calendar.Event{
		Id:          eventID,
		Summary:     "Mock Event",
		Description: "This is a mock event for testing",
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: "2024-01-01T10:00:00Z",
		},
		End: &calendar.EventDateTime{
			DateTime: "2024-01-01T11:00:00Z",
		},
		Location: "Mock Location",
		HtmlLink: "https://calendar.google.com/mock",
	}, nil
}
func (m *MockCalendarService) ListCalendars() ([]*calendar.CalendarListEntry, error) {
	return []*calendar.CalendarListEntry{}, nil
}
func (m *MockCalendarService) CheckConflicts(calendarID string, startTime, endTime time.Time) ([]*calendar.Event, error) {
	m.logger.Debug("Mock: checking conflicts",
		zap.String("calendarID", calendarID),
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime))

	mockConflictStart := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 10, 0, 0, 0, startTime.Location())
	mockConflictEnd := mockConflictStart.Add(time.Hour)

	if startTime.Before(mockConflictEnd) && endTime.After(mockConflictStart) {
		return []*calendar.Event{
			{
				Id:      "mock-conflict-event",
				Summary: "Mock Existing Meeting",
				Status:  "confirmed",
				Start: &calendar.EventDateTime{
					DateTime: mockConflictStart.Format(time.RFC3339),
				},
				End: &calendar.EventDateTime{
					DateTime: mockConflictEnd.Format(time.RFC3339),
				},
				Location: "Mock Conference Room",
			},
		}, nil
	}

	return []*calendar.Event{}, nil
}
