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
)

func TestFindAvailableTimeHandler(t *testing.T) {
	t.Setenv("GOOGLE_CALENDAR_TIMEZONE", "UTC")
	t.Setenv("TZ", "")

	timed := func(startRFC, endRFC string) *calendar.Event {
		return &calendar.Event{
			Id:    "timed-" + startRFC,
			Start: &calendar.EventDateTime{DateTime: startRFC},
			End:   &calendar.EventDateTime{DateTime: endRFC},
		}
	}
	allDay := func(startDate, endDate string) *calendar.Event {
		return &calendar.Event{
			Id:    "allday-" + startDate,
			Start: &calendar.EventDateTime{Date: startDate},
			End:   &calendar.EventDateTime{Date: endDate},
		}
	}

	type slotExpect struct {
		startTime string
		endTime   string
	}
	hourly := func(from string, n int) []slotExpect {
		start, _ := time.Parse(time.RFC3339, from)
		slots := make([]slotExpect, 0, n)
		for i := 0; i < n; i++ {
			t := start.Add(time.Duration(i) * time.Hour)
			slots = append(slots, slotExpect{t.Format(time.RFC3339), t.Add(time.Hour).Format(time.RFC3339)})
		}
		return slots
	}

	tests := []struct {
		name             string
		args             map[string]any
		events           []*calendar.Event
		calendarTimeZone string
		listErr          error
		wantErr          bool
		wantErrSub       string
		wantSlots        []slotExpect
		wantTruncated    bool
	}{
		{
			name: "a window wider than the cap is truncated",
			args: map[string]any{
				"startDate": "2026-05-20T00:00:00Z",
				"endDate":   "2026-05-23T00:00:00Z",
				"duration":  float64(60),
			},
			wantSlots:     hourly("2026-05-20T00:00:00Z", 50),
			wantTruncated: true,
		},
		{
			name: "all-day event is placed in the calendar time zone, not the user's",
			args: map[string]any{
				"startDate": "2026-05-23T02:00:00Z",
				"endDate":   "2026-05-23T04:00:00Z",
				"duration":  float64(60),
			},
			events:           []*calendar.Event{allDay("2026-05-23", "2026-05-24")},
			calendarTimeZone: "America/New_York",
			wantSlots:        hourly("2026-05-23T02:00:00Z", 2),
		},
		{
			name: "cancelled recurring instance does not block a slot",
			args: map[string]any{
				"startDate": "2026-05-23T09:00:00Z",
				"endDate":   "2026-05-23T11:00:00Z",
				"duration":  float64(60),
			},
			events: []*calendar.Event{
				{Id: "cancelled", Status: "cancelled", Start: &calendar.EventDateTime{DateTime: "2026-05-23T09:00:00Z"}, End: &calendar.EventDateTime{DateTime: "2026-05-23T10:00:00Z"}},
			},
			wantSlots: hourly("2026-05-23T09:00:00Z", 2),
		},
		{
			name: "happy path with timed events emits every slot before/between/after",
			args: map[string]any{
				"startDate": "2026-05-23T09:00:00Z",
				"endDate":   "2026-05-23T17:00:00Z",
				"duration":  float64(60),
			},
			events: []*calendar.Event{
				timed("2026-05-23T10:00:00Z", "2026-05-23T11:00:00Z"),
				timed("2026-05-23T14:00:00Z", "2026-05-23T15:00:00Z"),
			},
			wantSlots: append(append(hourly("2026-05-23T09:00:00Z", 1), hourly("2026-05-23T11:00:00Z", 3)...), hourly("2026-05-23T15:00:00Z", 2)...),
		},
		{
			name: "events returned out of order are sorted before slot detection",
			args: map[string]any{
				"startDate": "2026-05-23T09:00:00Z",
				"endDate":   "2026-05-23T17:00:00Z",
				"duration":  float64(60),
			},
			events: []*calendar.Event{
				timed("2026-05-23T14:00:00Z", "2026-05-23T15:00:00Z"),
				timed("2026-05-23T10:00:00Z", "2026-05-23T11:00:00Z"),
			},
			wantSlots: append(append(hourly("2026-05-23T09:00:00Z", 1), hourly("2026-05-23T11:00:00Z", 3)...), hourly("2026-05-23T15:00:00Z", 2)...),
		},
		{
			name: "all-day event blocks the entire day",
			args: map[string]any{
				"startDate": "2026-05-23T00:00:00Z",
				"endDate":   "2026-05-23T23:59:59Z",
				"duration":  float64(60),
			},
			events: []*calendar.Event{
				allDay("2026-05-23", "2026-05-24"),
			},
			wantSlots: nil,
		},
		{
			name: "all-day event blocks but leaves room before and after",
			args: map[string]any{
				"startDate": "2026-05-22T09:00:00Z",
				"endDate":   "2026-05-24T17:00:00Z",
				"duration":  float64(60),
			},
			events: []*calendar.Event{
				allDay("2026-05-23", "2026-05-24"),
			},
			wantSlots: append(hourly("2026-05-22T09:00:00Z", 15), hourly("2026-05-24T00:00:00Z", 17)...),
		},
		{
			name: "timed event nested in an all-day event does not open a gap inside the day",
			args: map[string]any{
				"startDate": "2026-05-23T09:00:00Z",
				"endDate":   "2026-05-24T03:00:00Z",
				"duration":  float64(60),
			},
			events: []*calendar.Event{
				allDay("2026-05-23", "2026-05-24"),
				timed("2026-05-23T10:00:00Z", "2026-05-23T11:00:00Z"),
			},
			wantSlots: hourly("2026-05-24T00:00:00Z", 3),
		},
		{
			name: "overlapping timed events are merged into one busy block",
			args: map[string]any{
				"startDate": "2026-05-23T09:00:00Z",
				"endDate":   "2026-05-23T14:00:00Z",
				"duration":  float64(60),
			},
			events: []*calendar.Event{
				timed("2026-05-23T09:00:00Z", "2026-05-23T12:00:00Z"),
				timed("2026-05-23T10:00:00Z", "2026-05-23T11:00:00Z"),
			},
			wantSlots: hourly("2026-05-23T12:00:00Z", 2),
		},
		{
			name: "startDate falling inside a busy period skips the pre-slot",
			args: map[string]any{
				"startDate": "2026-05-23T10:00:00Z",
				"endDate":   "2026-05-23T17:00:00Z",
				"duration":  float64(60),
			},
			events: []*calendar.Event{
				timed("2026-05-23T09:00:00Z", "2026-05-23T11:00:00Z"),
			},
			wantSlots: hourly("2026-05-23T11:00:00Z", 6),
		},
		{
			name: "back-to-back meetings leave no slots fitting the duration",
			args: map[string]any{
				"startDate": "2026-05-23T09:00:00Z",
				"endDate":   "2026-05-23T12:00:00Z",
				"duration":  float64(30),
			},
			events: []*calendar.Event{
				timed("2026-05-23T09:00:00Z", "2026-05-23T10:00:00Z"),
				timed("2026-05-23T10:00:00Z", "2026-05-23T11:00:00Z"),
				timed("2026-05-23T11:00:00Z", "2026-05-23T12:00:00Z"),
			},
			wantSlots: nil,
		},
		{
			name: "default duration of 60 minutes when omitted",
			args: map[string]any{
				"startDate": "2026-05-23T09:00:00Z",
				"endDate":   "2026-05-23T11:00:00Z",
			},
			events:    nil,
			wantSlots: hourly("2026-05-23T09:00:00Z", 2),
		},
		{
			name: "missing startDate returns error",
			args: map[string]any{
				"endDate": "2026-05-23T17:00:00Z",
			},
			wantErr:    true,
			wantErrSub: "startDate is required",
		},
		{
			name: "missing endDate returns error",
			args: map[string]any{
				"startDate": "2026-05-23T09:00:00Z",
			},
			wantErr:    true,
			wantErrSub: "endDate is required",
		},
		{
			name: "invalid startDate format returns error",
			args: map[string]any{
				"startDate": "tomorrow",
				"endDate":   "2026-05-23T17:00:00Z",
			},
			wantErr:    true,
			wantErrSub: "invalid startDate format",
		},
		{
			name: "ListEvents failure is wrapped",
			args: map[string]any{
				"startDate": "2026-05-23T09:00:00Z",
				"endDate":   "2026-05-23T17:00:00Z",
			},
			listErr:    errors.New("api down"),
			wantErr:    true,
			wantErrSub: "failed to list events for availability check",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubCalendarService{
				listEventsFn: func(calendarID string, timeMin, timeMax time.Time) ([]*calendar.Event, string, error) {
					if tc.listErr != nil {
						return nil, "", tc.listErr
					}
					return tc.events, tc.calendarTimeZone, nil
				},
			}
			tool := &FindAvailableTimeTool{logger: zap.NewNop(), google: stub}
			result, err := tool.FindAvailableTimeHandler(context.Background(), tc.args)

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

			var parsed struct {
				Success        bool             `json:"success"`
				SlotCount      int              `json:"slotCount"`
				Truncated      bool             `json:"truncated"`
				AvailableSlots []map[string]any `json:"availableSlots"`
			}
			if err := json.Unmarshal([]byte(result), &parsed); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}

			if !parsed.Success {
				t.Errorf("success = false, want true")
			}
			if parsed.Truncated != tc.wantTruncated {
				t.Errorf("truncated = %v, want %v", parsed.Truncated, tc.wantTruncated)
			}
			if got, want := parsed.SlotCount, len(tc.wantSlots); got != want {
				t.Errorf("slotCount = %d, want %d (slots=%+v)", got, want, parsed.AvailableSlots)
			}
			for i, want := range tc.wantSlots {
				if i >= len(parsed.AvailableSlots) {
					t.Errorf("missing slot %d: want %+v", i, want)
					continue
				}
				got := parsed.AvailableSlots[i]
				if got["startTime"] != want.startTime {
					t.Errorf("slot %d startTime = %v, want %v", i, got["startTime"], want.startTime)
				}
				if got["endTime"] != want.endTime {
					t.Errorf("slot %d endTime = %v, want %v", i, got["endTime"], want.endTime)
				}
			}
		})
	}
}
