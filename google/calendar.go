package google

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// CalendarService represents the interface for interacting with Google Calendar API
//
//go:generate counterfeiter -generate
//counterfeiter:generate -o ../tests/mocks . CalendarService
type CalendarService interface {
	ListEvents(calendarID string, timeMin, timeMax time.Time) ([]*calendar.Event, error)
	CreateEvent(calendarID string, event *calendar.Event) (*calendar.Event, error)
	UpdateEvent(calendarID, eventID string, event *calendar.Event) (*calendar.Event, error)
	DeleteEvent(calendarID, eventID string) error
	GetEvent(calendarID, eventID string) (*calendar.Event, error)
	ListCalendars() ([]*calendar.CalendarListEntry, error)
}

// CalendarServiceImpl implements the calendar service interface for Google Calendar API
type CalendarServiceImpl struct {
	service *calendar.Service
	logger  *zap.Logger
}

// NewCalendarService creates a new Google Calendar service
func NewCalendarService(ctx context.Context, logger *zap.Logger, opts ...option.ClientOption) (CalendarService, error) {
	scopesOption := option.WithScopes(
		calendar.CalendarReadonlyScope,
		calendar.CalendarScope,
	)

	allOptions := append([]option.ClientOption{scopesOption}, opts...)

	svc, err := calendar.NewService(ctx, allOptions...)
	if err != nil {
		return nil, fmt.Errorf("unable to create calendar service: %w", err)
	}
	return &CalendarServiceImpl{service: svc, logger: logger}, nil
}

// ListEvents retrieves events from the calendar within the specified time range
func (g *CalendarServiceImpl) ListEvents(calendarID string, timeMin, timeMax time.Time) ([]*calendar.Event, error) {
	g.logger.Debug("listing events",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "list-events"),
		zap.String("calendarID", calendarID),
		zap.Time("timeMin", timeMin),
		zap.Time("timeMax", timeMax))

	g.logger.Debug("google calendar api request parameters",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "list-events"),
		zap.String("calendarID", calendarID),
		zap.String("timeMinRFC3339", timeMin.Format(time.RFC3339)),
		zap.String("timeMaxRFC3339", timeMax.Format(time.RFC3339)),
		zap.String("orderBy", "startTime"),
		zap.Bool("singleEvents", true))

	events, err := g.service.Events.List(calendarID).
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		OrderBy("startTime").
		SingleEvents(true).
		Do()
	if err != nil {
		g.logger.Error("failed to retrieve events from google calendar api",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "list-events"),
			zap.String("calendarID", calendarID),
			zap.Error(err))
		return nil, fmt.Errorf("unable to retrieve events: %w", err)
	}

	g.logger.Debug("google calendar api response details",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "list-events"),
		zap.String("calendarID", calendarID),
		zap.String("kind", events.Kind),
		zap.String("etag", events.Etag),
		zap.String("summary", events.Summary),
		zap.String("description", events.Description),
		zap.String("timeZone", events.TimeZone),
		zap.String("accessRole", events.AccessRole),
		zap.String("nextPageToken", events.NextPageToken),
		zap.String("nextSyncToken", events.NextSyncToken),
		zap.Int("itemCount", len(events.Items)))

	for i, event := range events.Items {
		eventJson, _ := json.MarshalIndent(event, "", "  ")
		g.logger.Debug("google calendar api event details",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "list-events"),
			zap.String("calendarID", calendarID),
			zap.Int("eventIndex", i),
			zap.String("eventId", event.Id),
			zap.String("eventSummary", event.Summary),
			zap.String("eventStatus", event.Status),
			zap.String("eventJson", string(eventJson)))
	}

	g.logger.Info("successfully retrieved events",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "list-events"),
		zap.String("calendarID", calendarID),
		zap.Int("eventCount", len(events.Items)))

	return events.Items, nil
}

// CreateEvent creates a new event in the calendar
func (g *CalendarServiceImpl) CreateEvent(calendarID string, event *calendar.Event) (*calendar.Event, error) {
	g.logger.Debug("creating event",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "create-event"),
		zap.String("calendarID", calendarID),
		zap.String("eventSummary", event.Summary),
		zap.String("eventStart", event.Start.DateTime))

	eventJson, _ := json.MarshalIndent(event, "", "  ")
	g.logger.Debug("google calendar api create event request",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "create-event"),
		zap.String("calendarID", calendarID),
		zap.String("requestJson", string(eventJson)))

	createdEvent, err := g.service.Events.Insert(calendarID, event).Do()
	if err != nil {
		g.logger.Error("failed to create event in google calendar api",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "create-event"),
			zap.String("calendarID", calendarID),
			zap.String("eventSummary", event.Summary),
			zap.Error(err))
		return nil, fmt.Errorf("unable to create event: %w", err)
	}

	responseJson, _ := json.MarshalIndent(createdEvent, "", "  ")
	g.logger.Debug("google calendar api create event response",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "create-event"),
		zap.String("calendarID", calendarID),
		zap.String("responseJson", string(responseJson)))

	g.logger.Info("successfully created event",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "create-event"),
		zap.String("calendarID", calendarID),
		zap.String("eventID", createdEvent.Id),
		zap.String("eventSummary", createdEvent.Summary))

	return createdEvent, nil
}

// UpdateEvent updates an existing event in the calendar
func (g *CalendarServiceImpl) UpdateEvent(calendarID, eventID string, event *calendar.Event) (*calendar.Event, error) {
	g.logger.Debug("updating event",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "update-event"),
		zap.String("calendarID", calendarID),
		zap.String("eventID", eventID),
		zap.String("eventSummary", event.Summary))

	updatedEvent, err := g.service.Events.Update(calendarID, eventID, event).Do()
	if err != nil {
		g.logger.Error("failed to update event in google calendar api",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "update-event"),
			zap.String("calendarID", calendarID),
			zap.String("eventID", eventID),
			zap.Error(err))
		return nil, fmt.Errorf("unable to update event: %w", err)
	}

	g.logger.Info("successfully updated event",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "update-event"),
		zap.String("calendarID", calendarID),
		zap.String("eventID", eventID),
		zap.String("eventSummary", updatedEvent.Summary))

	return updatedEvent, nil
}

// DeleteEvent removes an event from the calendar
func (g *CalendarServiceImpl) DeleteEvent(calendarID, eventID string) error {
	g.logger.Debug("deleting event",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "delete-event"),
		zap.String("calendarID", calendarID),
		zap.String("eventID", eventID))

	err := g.service.Events.Delete(calendarID, eventID).Do()
	if err != nil {
		g.logger.Error("failed to delete event from google calendar api",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "delete-event"),
			zap.String("calendarID", calendarID),
			zap.String("eventID", eventID),
			zap.Error(err))
		return fmt.Errorf("unable to delete event: %w", err)
	}

	g.logger.Info("successfully deleted event",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "delete-event"),
		zap.String("calendarID", calendarID),
		zap.String("eventID", eventID))

	return nil
}

// GetEvent retrieves a specific event from the calendar
func (g *CalendarServiceImpl) GetEvent(calendarID, eventID string) (*calendar.Event, error) {
	g.logger.Debug("getting event",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "get-event"),
		zap.String("calendarID", calendarID),
		zap.String("eventID", eventID))

	event, err := g.service.Events.Get(calendarID, eventID).Do()
	if err != nil {
		g.logger.Error("failed to get event from google calendar api",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "get-event"),
			zap.String("calendarID", calendarID),
			zap.String("eventID", eventID),
			zap.Error(err))
		return nil, fmt.Errorf("unable to get event: %w", err)
	}

	g.logger.Info("successfully retrieved event",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "get-event"),
		zap.String("calendarID", calendarID),
		zap.String("eventID", eventID),
		zap.String("eventSummary", event.Summary))

	return event, nil
}

// ListCalendars retrieves all available calendars
func (g *CalendarServiceImpl) ListCalendars() ([]*calendar.CalendarListEntry, error) {
	g.logger.Debug("listing calendars",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "list-calendars"))

	calendarList, err := g.service.CalendarList.List().Do()
	if err != nil {
		g.logger.Error("failed to list calendars from google calendar api",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "list-calendars"),
			zap.Error(err))
		return nil, fmt.Errorf("unable to list calendars: %w", err)
	}

	g.logger.Debug("google calendar api calendars response details",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "list-calendars"),
		zap.String("kind", calendarList.Kind),
		zap.String("etag", calendarList.Etag),
		zap.String("nextPageToken", calendarList.NextPageToken),
		zap.String("nextSyncToken", calendarList.NextSyncToken),
		zap.Int("itemCount", len(calendarList.Items)))

	for i, cal := range calendarList.Items {
		calendarJson, _ := json.MarshalIndent(cal, "", "  ")
		g.logger.Debug("google calendar api calendar details",
			zap.String("component", "google-calendar-service"),
			zap.String("operation", "list-calendars"),
			zap.Int("calendarIndex", i),
			zap.String("calendarId", cal.Id),
			zap.String("calendarSummary", cal.Summary),
			zap.String("calendarDescription", cal.Description),
			zap.String("calendarTimeZone", cal.TimeZone),
			zap.String("calendarAccessRole", cal.AccessRole),
			zap.Bool("calendarPrimary", cal.Primary),
			zap.Bool("calendarSelected", cal.Selected),
			zap.String("calendarJson", string(calendarJson)))
	}

	g.logger.Info("successfully retrieved calendars",
		zap.String("component", "google-calendar-service"),
		zap.String("operation", "list-calendars"),
		zap.Int("calendarCount", len(calendarList.Items)))

	return calendarList.Items, nil
}
