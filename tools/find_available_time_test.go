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

	tests := []struct {
		name       string
		args       map[string]any
		events     []*calendar.Event
		listErr    error
		wantErr    bool
		wantErrSub string
		wantSlots  []slotExpect
	}{
		{
			name: "happy path with timed events emits before/between/after slots",
			args: map[string]any{
				"startDate": "2026-05-23T09:00:00Z",
				"endDate":   "2026-05-23T17:00:00Z",
				"duration":  float64(60),
			},
			events: []*calendar.Event{
				timed("2026-05-23T10:00:00Z", "2026-05-23T11:00:00Z"),
				timed("2026-05-23T14:00:00Z", "2026-05-23T15:00:00Z"),
			},
			wantSlots: []slotExpect{
				{"2026-05-23T09:00:00Z", "2026-05-23T10:00:00Z"},
				{"2026-05-23T11:00:00Z", "2026-05-23T12:00:00Z"},
				{"2026-05-23T15:00:00Z", "2026-05-23T16:00:00Z"},
			},
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
			wantSlots: []slotExpect{
				{"2026-05-23T09:00:00Z", "2026-05-23T10:00:00Z"},
				{"2026-05-23T11:00:00Z", "2026-05-23T12:00:00Z"},
				{"2026-05-23T15:00:00Z", "2026-05-23T16:00:00Z"},
			},
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
			wantSlots: []slotExpect{
				{"2026-05-22T09:00:00Z", "2026-05-22T10:00:00Z"},
				{"2026-05-24T00:00:00Z", "2026-05-24T01:00:00Z"},
			},
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
			wantSlots: []slotExpect{
				{"2026-05-23T11:00:00Z", "2026-05-23T12:00:00Z"},
			},
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
			events: nil,
			wantSlots: []slotExpect{
				{"2026-05-23T09:00:00Z", "2026-05-23T10:00:00Z"},
			},
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
				listEventsFn: func(calendarID string, timeMin, timeMax time.Time) ([]*calendar.Event, error) {
					if tc.listErr != nil {
						return nil, tc.listErr
					}
					return tc.events, nil
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
				AvailableSlots []map[string]any `json:"availableSlots"`
			}
			if err := json.Unmarshal([]byte(result), &parsed); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}

			if !parsed.Success {
				t.Errorf("success = false, want true")
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
