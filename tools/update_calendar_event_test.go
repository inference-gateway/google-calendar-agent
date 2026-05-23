package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"

	google "github.com/inference-gateway/google-calendar-agent/internal/google"
)

// stubCalendarService is a configurable test stub for google.CalendarService.
// Each method delegates to a function field so individual tests dictate behavior.
type stubCalendarService struct {
	getEventFn       func(calendarID, eventID string) (*calendar.Event, error)
	updateEventFn    func(calendarID, eventID string, event *calendar.Event) (*calendar.Event, error)
	createEventFn    func(calendarID string, event *calendar.Event) (*calendar.Event, error)
	deleteEventFn    func(calendarID, eventID string) error
	listEventsFn     func(calendarID string, timeMin, timeMax time.Time) ([]*calendar.Event, error)
	checkConflictsFn func(calendarID string, startTime, endTime time.Time) ([]*calendar.Event, error)
	listCalendarsFn  func() ([]*calendar.CalendarListEntry, error)
	calendarID       string
}

var _ google.CalendarService = (*stubCalendarService)(nil)

func (s *stubCalendarService) ListEvents(calendarID string, timeMin, timeMax time.Time) ([]*calendar.Event, error) {
	if s.listEventsFn == nil {
		return nil, errors.New("ListEvents unexpectedly called")
	}
	return s.listEventsFn(calendarID, timeMin, timeMax)
}

func (s *stubCalendarService) CreateEvent(calendarID string, event *calendar.Event) (*calendar.Event, error) {
	if s.createEventFn == nil {
		return nil, errors.New("CreateEvent unexpectedly called")
	}
	return s.createEventFn(calendarID, event)
}

func (s *stubCalendarService) UpdateEvent(calendarID, eventID string, event *calendar.Event) (*calendar.Event, error) {
	if s.updateEventFn == nil {
		return nil, errors.New("UpdateEvent unexpectedly called")
	}
	return s.updateEventFn(calendarID, eventID, event)
}

func (s *stubCalendarService) DeleteEvent(calendarID, eventID string) error {
	if s.deleteEventFn == nil {
		return errors.New("DeleteEvent unexpectedly called")
	}
	return s.deleteEventFn(calendarID, eventID)
}

func (s *stubCalendarService) GetEvent(calendarID, eventID string) (*calendar.Event, error) {
	if s.getEventFn == nil {
		return nil, errors.New("GetEvent unexpectedly called")
	}
	return s.getEventFn(calendarID, eventID)
}

func (s *stubCalendarService) ListCalendars() ([]*calendar.CalendarListEntry, error) {
	if s.listCalendarsFn == nil {
		return nil, errors.New("ListCalendars unexpectedly called")
	}
	return s.listCalendarsFn()
}

func (s *stubCalendarService) CheckConflicts(calendarID string, startTime, endTime time.Time) ([]*calendar.Event, error) {
	if s.checkConflictsFn == nil {
		return nil, errors.New("CheckConflicts unexpectedly called")
	}
	return s.checkConflictsFn(calendarID, startTime, endTime)
}

func (s *stubCalendarService) GetCalendarID() string {
	if s.calendarID == "" {
		return "primary"
	}
	return s.calendarID
}

func TestUpdateCalendarEventHandler(t *testing.T) {
	baseEvent := func() *calendar.Event {
		return &calendar.Event{
			Id:          "evt-1",
			Summary:     "Original",
			Description: "Original description",
			Location:    "Original location",
			Status:      "confirmed",
			HtmlLink:    "https://example.com/evt-1",
			Start:       &calendar.EventDateTime{DateTime: "2026-05-23T10:00:00Z"},
			End:         &calendar.EventDateTime{DateTime: "2026-05-23T11:00:00Z"},
		}
	}

	tests := []struct {
		name          string
		args          map[string]any
		getEventFn    func(calendarID, eventID string) (*calendar.Event, error)
		updateEventFn func(calendarID, eventID string, event *calendar.Event) (*calendar.Event, error)
		wantErr       bool
		wantErrSub    string
		wantSummary   string
		wantStart     string
		wantEnd       string
	}{
		{
			name: "happy path updates all fields",
			args: map[string]any{
				"eventId":     "evt-1",
				"summary":     "New title",
				"description": "New description",
				"location":    "New location",
				"startTime":   "2026-05-23T12:00:00Z",
				"endTime":     "2026-05-23T13:00:00Z",
			},
			getEventFn: func(calendarID, eventID string) (*calendar.Event, error) {
				return baseEvent(), nil
			},
			updateEventFn: func(calendarID, eventID string, event *calendar.Event) (*calendar.Event, error) {
				return event, nil
			},
			wantSummary: "New title",
			wantStart:   "2026-05-23T12:00:00Z",
			wantEnd:     "2026-05-23T13:00:00Z",
		},
		{
			name: "partial update leaves untouched fields alone",
			args: map[string]any{
				"eventId": "evt-1",
				"summary": "New title",
			},
			getEventFn: func(calendarID, eventID string) (*calendar.Event, error) {
				return baseEvent(), nil
			},
			updateEventFn: func(calendarID, eventID string, event *calendar.Event) (*calendar.Event, error) {
				return event, nil
			},
			wantSummary: "New title",
			wantStart:   "2026-05-23T10:00:00Z",
			wantEnd:     "2026-05-23T11:00:00Z",
		},
		{
			name:       "missing eventId returns error before any API call",
			args:       map[string]any{},
			wantErr:    true,
			wantErrSub: "eventId is required",
		},
		{
			name: "wrong-typed summary returns error and does not panic",
			args: map[string]any{
				"eventId": "evt-1",
				"summary": 42,
			},
			getEventFn: func(calendarID, eventID string) (*calendar.Event, error) {
				return baseEvent(), nil
			},
			wantErr:    true,
			wantErrSub: "summary must be a string",
		},
		{
			name: "wrong-typed endTime returns error and does not panic",
			args: map[string]any{
				"eventId": "evt-1",
				"endTime": []any{"2026-05-23T13:00:00Z"},
			},
			getEventFn: func(calendarID, eventID string) (*calendar.Event, error) {
				return baseEvent(), nil
			},
			wantErr:    true,
			wantErrSub: "endTime must be a string",
		},
		{
			name: "wrong-typed startTime returns error and does not panic",
			args: map[string]any{
				"eventId":   "evt-1",
				"startTime": 0.5,
			},
			getEventFn: func(calendarID, eventID string) (*calendar.Event, error) {
				return baseEvent(), nil
			},
			wantErr:    true,
			wantErrSub: "startTime must be a string",
		},
		{
			name: "GetEvent error is wrapped and returned",
			args: map[string]any{
				"eventId": "evt-1",
			},
			getEventFn: func(calendarID, eventID string) (*calendar.Event, error) {
				return nil, errors.New("boom")
			},
			wantErr:    true,
			wantErrSub: "failed to get existing calendar event",
		},
		{
			name: "UpdateEvent error is wrapped and returned",
			args: map[string]any{
				"eventId": "evt-1",
				"summary": "New",
			},
			getEventFn: func(calendarID, eventID string) (*calendar.Event, error) {
				return baseEvent(), nil
			},
			updateEventFn: func(calendarID, eventID string, event *calendar.Event) (*calendar.Event, error) {
				return nil, errors.New("rate limited")
			},
			wantErr:    true,
			wantErrSub: "failed to update calendar event",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubCalendarService{
				getEventFn:    tc.getEventFn,
				updateEventFn: tc.updateEventFn,
			}
			tool := &UpdateCalendarEventTool{logger: zap.NewNop(), google: stub}
			result, err := tool.UpdateCalendarEventHandler(context.Background(), tc.args)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (result=%q)", tc.wantErrSub, result)
				}
				if !strings.Contains(err.Error(), tc.wantErrSub) {
					t.Errorf("error = %q, want substring %q", err.Error(), tc.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var parsed map[string]any
			if err := json.Unmarshal([]byte(result), &parsed); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}
			if got := parsed["summary"]; got != tc.wantSummary {
				t.Errorf("summary = %v, want %v", got, tc.wantSummary)
			}
			if got := parsed["startTime"]; got != tc.wantStart {
				t.Errorf("startTime = %v, want %v", got, tc.wantStart)
			}
			if got := parsed["endTime"]; got != tc.wantEnd {
				t.Errorf("endTime = %v, want %v", got, tc.wantEnd)
			}
			if got := parsed["success"]; got != true {
				t.Errorf("success = %v, want true", got)
			}
		})
	}
}
