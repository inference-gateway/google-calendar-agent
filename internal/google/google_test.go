package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	config "github.com/inference-gateway/google-calendar-agent/config"
	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"
	option "google.golang.org/api/option"
)

func newTestService(t *testing.T, pages map[string]*calendar.Events) (*CalendarServiceImpl, *int) {
	t.Helper()
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		page, ok := pages[r.URL.Query().Get("pageToken")]
		if !ok {
			http.Error(w, "unknown page", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(page)
	}))
	t.Cleanup(srv.Close)

	svc, err := calendar.NewService(context.Background(),
		option.WithEndpoint(srv.URL+"/"),
		option.WithoutAuthentication(),
		option.WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("failed to create calendar service: %v", err)
	}
	return &CalendarServiceImpl{service: svc, logger: zap.NewNop(), config: &config.Config{}}, &requests
}

func timedEvent(id, start, end string) *calendar.Event {
	return &calendar.Event{
		Id:    id,
		Start: &calendar.EventDateTime{DateTime: start},
		End:   &calendar.EventDateTime{DateTime: end},
	}
}

func allDayEvent(id, start, end string) *calendar.Event {
	return &calendar.Event{
		Id:    id,
		Start: &calendar.EventDateTime{Date: start},
		End:   &calendar.EventDateTime{Date: end},
	}
}

func eventIDs(events []*calendar.Event) []string {
	ids := make([]string, 0, len(events))
	for _, e := range events {
		ids = append(ids, e.Id)
	}
	return ids
}

func TestCheckConflicts(t *testing.T) {
	tests := []struct {
		name    string
		pages   map[string]*calendar.Events
		start   string
		end     string
		wantIDs []string
	}{
		{
			name: "all-day event is reported as a conflict",
			pages: map[string]*calendar.Events{
				"": {TimeZone: "UTC", Items: []*calendar.Event{allDayEvent("holiday", "2026-05-23", "2026-05-24")}},
			},
			start:   "2026-05-23T10:00:00Z",
			end:     "2026-05-23T11:00:00Z",
			wantIDs: []string{"holiday"},
		},
		{
			name: "all-day event is placed in the calendar time zone",
			pages: map[string]*calendar.Events{
				"": {TimeZone: "America/New_York", Items: []*calendar.Event{allDayEvent("holiday", "2026-05-23", "2026-05-24")}},
			},
			start:   "2026-05-23T02:00:00Z",
			end:     "2026-05-23T03:00:00Z",
			wantIDs: []string{},
		},
		{
			name: "events on later pages are included",
			pages: map[string]*calendar.Events{
				"":      {TimeZone: "UTC", NextPageToken: "page2", Items: []*calendar.Event{timedEvent("first", "2026-05-23T08:00:00Z", "2026-05-23T09:00:00Z")}},
				"page2": {Items: []*calendar.Event{timedEvent("second", "2026-05-23T10:30:00Z", "2026-05-23T11:30:00Z")}},
			},
			start:   "2026-05-23T10:00:00Z",
			end:     "2026-05-23T11:00:00Z",
			wantIDs: []string{"second"},
		},
		{
			name: "touching events are not conflicts",
			pages: map[string]*calendar.Events{
				"": {TimeZone: "UTC", Items: []*calendar.Event{timedEvent("before", "2026-05-23T09:00:00Z", "2026-05-23T10:00:00Z")}},
			},
			start:   "2026-05-23T10:00:00Z",
			end:     "2026-05-23T11:00:00Z",
			wantIDs: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := newTestService(t, tc.pages)
			start, _ := time.Parse(time.RFC3339, tc.start)
			end, _ := time.Parse(time.RFC3339, tc.end)

			conflicts, err := svc.CheckConflicts("primary", start, end)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := eventIDs(conflicts)
			if len(got) != len(tc.wantIDs) {
				t.Fatalf("conflicts = %v, want %v", got, tc.wantIDs)
			}
			for i := range got {
				if got[i] != tc.wantIDs[i] {
					t.Errorf("conflicts = %v, want %v", got, tc.wantIDs)
				}
			}
		})
	}
}

func TestListEventsPagination(t *testing.T) {
	pages := map[string]*calendar.Events{
		"":      {NextPageToken: "page2", Items: []*calendar.Event{timedEvent("first", "2026-05-23T08:00:00Z", "2026-05-23T09:00:00Z")}},
		"page2": {Items: []*calendar.Event{timedEvent("second", "2026-05-23T10:00:00Z", "2026-05-23T11:00:00Z")}},
	}
	timeMin, _ := time.Parse(time.RFC3339, "2026-05-23T00:00:00Z")

	t.Run("bounded window reads every page", func(t *testing.T) {
		svc, requests := newTestService(t, pages)
		events, err := svc.ListEvents("primary", timeMin, timeMin.Add(24*time.Hour))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 || *requests != 2 {
			t.Errorf("events = %v after %d requests, want 2 events after 2 requests", eventIDs(events), *requests)
		}
	})

	t.Run("unbounded window reads a single page", func(t *testing.T) {
		svc, requests := newTestService(t, pages)
		events, err := svc.ListEvents("primary", timeMin, time.Time{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 1 || *requests != 1 {
			t.Errorf("events = %v after %d requests, want 1 event after 1 request", eventIDs(events), *requests)
		}
	})
}

func TestEventDateTimeValue(t *testing.T) {
	tests := []struct {
		name string
		edt  *calendar.EventDateTime
		want string
	}{
		{"nil", nil, ""},
		{"timed", &calendar.EventDateTime{DateTime: "2026-05-23T10:00:00Z", TimeZone: "UTC"}, "2026-05-23T10:00:00Z"},
		{"all-day", &calendar.EventDateTime{Date: "2026-05-23"}, "2026-05-23"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := EventDateTimeValue(tc.edt); got != tc.want {
				t.Errorf("EventDateTimeValue() = %q, want %q", got, tc.want)
			}
		})
	}
}
