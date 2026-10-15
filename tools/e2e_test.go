//go:build e2e

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	config "github.com/inference-gateway/google-calendar-agent/config"
	google "github.com/inference-gateway/google-calendar-agent/internal/google"
	envconfig "github.com/sethvargo/go-envconfig"
	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"
)

type e2eHarness struct {
	t          *testing.T
	svc        google.CalendarService
	calendarID string
	prefix     string
	day        time.Time
}

func newE2EHarness(t *testing.T) *e2eHarness {
	t.Helper()
	var cfg config.Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.GoogleCalendar.MockMode {
		t.Skip("GOOGLE_CALENDAR_MOCK_MODE is set, e2e needs the real API")
	}
	svc, err := google.NewServiceFactory(zap.NewNop(), &cfg)
	if err != nil {
		t.Fatalf("create calendar service: %v", err)
	}
	day := time.Now().UTC().AddDate(0, 0, 30).Truncate(24 * time.Hour)
	h := &e2eHarness{t: t, svc: svc, calendarID: svc.GetCalendarID(), prefix: fmt.Sprintf("e2e-%d", time.Now().Unix()), day: day}
	t.Cleanup(h.sweep)
	return h
}

func (h *e2eHarness) sweep() {
	events, _, err := h.svc.ListEvents(h.calendarID, h.day.AddDate(0, 0, -1), h.day.AddDate(0, 0, 4))
	if err != nil {
		h.t.Logf("sweep: list failed: %v", err)
		return
	}
	for _, e := range events {
		if strings.HasPrefix(e.Summary, "e2e-") {
			if err := h.svc.DeleteEvent(h.calendarID, e.Id); err != nil {
				h.t.Logf("sweep: delete %s failed: %v", e.Id, err)
			}
		}
	}
}

func (h *e2eHarness) at(hour int) string {
	return h.day.Add(time.Duration(hour) * time.Hour).Format(time.RFC3339)
}

func (h *e2eHarness) date(offsetDays int) string {
	return h.day.AddDate(0, 0, offsetDays).Format(google.DateLayout)
}

func call(t *testing.T, fn func(context.Context, map[string]any) (string, error), args map[string]any) map[string]any {
	t.Helper()
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("unmarshal %q: %v", out, err)
	}
	return parsed
}

func TestE2E(t *testing.T) {
	h := newE2EHarness(t)
	logger := zap.NewNop()
	create := &CreateCalendarEventTool{logger: logger, google: h.svc}
	check := &CheckConflictsTool{logger: logger, google: h.svc}
	update := &UpdateCalendarEventTool{logger: logger, google: h.svc}
	find := &FindAvailableTimeTool{logger: logger, google: h.svc}
	list := &ListCalendarEventsTool{logger: logger, google: h.svc}
	del := &DeleteCalendarEventTool{logger: logger, google: h.svc}
	t.Logf("calendar %s, test day %s, prefix %s", h.calendarID, h.day.Format(google.DateLayout), h.prefix)

	var timedID, allDayID, tzID string

	t.Run("create timed event", func(t *testing.T) {
		res := call(t, create.CreateCalendarEventHandler, map[string]any{
			"summary": h.prefix + " timed", "startTime": h.at(10), "endTime": h.at(11),
		})
		timedID, _ = res["eventId"].(string)
		if timedID == "" {
			t.Fatalf("no eventId in %v", res)
		}
	})

	t.Run("timed conflict is detected and touching window is not", func(t *testing.T) {
		res := call(t, check.CheckConflictsHandler, map[string]any{"startTime": h.at(10), "endTime": h.at(11)})
		conflicts, _ := res["conflicts"].([]any)
		if !containsEvent(conflicts, timedID) {
			t.Errorf("expected %s in conflicts, got %v", timedID, res)
		}
		if c := conflicts[0].(map[string]any); c["allDay"] != false {
			t.Errorf("allDay = %v, want false", c["allDay"])
		}
		res = call(t, check.CheckConflictsHandler, map[string]any{"startTime": h.at(11), "endTime": h.at(12)})
		if containsEvent(res["conflicts"], timedID) {
			t.Errorf("touching window should not conflict: %v", res)
		}
	})

	t.Run("all-day event conflicts inside the day with date values", func(t *testing.T) {
		ev, err := h.svc.CreateEvent(h.calendarID, &calendar.Event{
			Summary: h.prefix + " all-day",
			Start:   &calendar.EventDateTime{Date: h.date(1)},
			End:     &calendar.EventDateTime{Date: h.date(2)},
		})
		if err != nil {
			t.Fatalf("create all-day: %v", err)
		}
		allDayID = ev.Id
		noon := h.day.AddDate(0, 0, 1).Add(12 * time.Hour)
		res := call(t, check.CheckConflictsHandler, map[string]any{
			"startTime": noon.Format(time.RFC3339), "endTime": noon.Add(time.Hour).Format(time.RFC3339),
		})
		conflicts, _ := res["conflicts"].([]any)
		if !containsEvent(conflicts, allDayID) {
			t.Fatalf("expected all-day %s in conflicts, got %v", allDayID, res)
		}
		c := findEvent(conflicts, allDayID)
		if c["allDay"] != true || c["startTime"] != h.date(1) || c["endTime"] != h.date(2) {
			t.Errorf("all-day conflict shape = %v", c)
		}
	})

	t.Run("find_available_time finds no slot inside the all-day event", func(t *testing.T) {
		noon := h.day.AddDate(0, 0, 1).Add(12 * time.Hour)
		res := call(t, find.FindAvailableTimeHandler, map[string]any{
			"startDate": noon.Format(time.RFC3339), "endDate": noon.Add(2 * time.Hour).Format(time.RFC3339), "duration": float64(60),
		})
		if res["slotCount"] != float64(0) {
			t.Errorf("slotCount = %v, want 0: %v", res["slotCount"], res)
		}
	})

	t.Run("find_available_time skips the timed event", func(t *testing.T) {
		res := call(t, find.FindAvailableTimeHandler, map[string]any{
			"startDate": h.at(9), "endDate": h.at(12), "duration": float64(60),
		})
		slots, _ := res["availableSlots"].([]any)
		for _, s := range slots {
			start := s.(map[string]any)["startTime"].(string)
			if start == h.at(10) {
				t.Errorf("slot at %s overlaps the timed event: %v", start, res)
			}
		}
		if len(slots) == 0 {
			t.Errorf("expected free slots around the timed event: %v", res)
		}
	})

	t.Run("moving an all-day event with dates keeps it all-day", func(t *testing.T) {
		res := call(t, update.UpdateCalendarEventHandler, map[string]any{
			"eventId": allDayID, "startTime": h.date(2), "endTime": h.date(3),
		})
		if res["allDay"] != true || res["startTime"] != h.date(2) {
			t.Errorf("update result = %v", res)
		}
	})

	t.Run("mixing a date start with a timed end is rejected before the API call", func(t *testing.T) {
		_, err := update.UpdateCalendarEventHandler(context.Background(), map[string]any{
			"eventId": allDayID, "startTime": h.at(10),
		})
		if err == nil || !strings.Contains(err.Error(), "must both be dates") {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("rescheduling keeps the event time zone", func(t *testing.T) {
		ev, err := h.svc.CreateEvent(h.calendarID, &calendar.Event{
			Summary: h.prefix + " berlin",
			Start:   &calendar.EventDateTime{DateTime: h.at(14), TimeZone: "Europe/Berlin"},
			End:     &calendar.EventDateTime{DateTime: h.at(15), TimeZone: "Europe/Berlin"},
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		tzID = ev.Id
		call(t, update.UpdateCalendarEventHandler, map[string]any{
			"eventId": tzID, "startTime": h.at(16), "endTime": h.at(17),
		})
		got, err := h.svc.GetEvent(h.calendarID, tzID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Start.TimeZone != "Europe/Berlin" || got.End.TimeZone != "Europe/Berlin" {
			t.Errorf("time zones = %q/%q, want Europe/Berlin", got.Start.TimeZone, got.End.TimeZone)
		}
	})

	t.Run("list_calendar_events returns the test events in a bounded window", func(t *testing.T) {
		res := call(t, list.ListCalendarEventsHandler, map[string]any{
			"timeMin": h.at(0), "timeMax": h.day.AddDate(0, 0, 4).Format(time.RFC3339), "maxResults": float64(100),
		})
		events, _ := res["events"].([]any)
		for _, id := range []string{timedID, allDayID, tzID} {
			if !containsEvent(events, id) {
				t.Errorf("event %s missing from listing: %v", id, res)
			}
		}
	})

	t.Run("delete cleans up", func(t *testing.T) {
		for _, id := range []string{timedID, allDayID, tzID} {
			call(t, del.DeleteCalendarEventHandler, map[string]any{"eventId": id})
		}
		res := call(t, check.CheckConflictsHandler, map[string]any{"startTime": h.at(10), "endTime": h.at(11)})
		if containsEvent(res["conflicts"], timedID) {
			t.Errorf("deleted event still conflicts: %v", res)
		}
	})
}

func containsEvent(list any, id string) bool {
	return findEvent(list, id) != nil
}

func findEvent(list any, id string) map[string]any {
	items, _ := list.([]any)
	for _, item := range items {
		if m, ok := item.(map[string]any); ok && m["eventId"] == id {
			return m
		}
	}
	return nil
}
