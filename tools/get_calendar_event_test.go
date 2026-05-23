package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"
)

func TestGetCalendarEventHandler(t *testing.T) {
	fullEvent := &calendar.Event{
		Id:          "evt-1",
		Summary:     "Quarterly review",
		Description: "Discuss Q2 metrics",
		Location:    "HQ-3F",
		Status:      "confirmed",
		HtmlLink:    "https://example.com/evt-1",
		Start:       &calendar.EventDateTime{DateTime: "2026-05-23T10:00:00Z"},
		End:         &calendar.EventDateTime{DateTime: "2026-05-23T11:00:00Z"},
		Attendees: []*calendar.EventAttendee{
			{Email: "a@example.com"},
			{Email: "b@example.com"},
		},
	}
	minimalEvent := &calendar.Event{
		Id:      "evt-min",
		Summary: "Bare event",
		Status:  "confirmed",
	}

	tests := []struct {
		name       string
		args       map[string]any
		getEventFn func(calendarID, eventID string) (*calendar.Event, error)
		wantErr    bool
		wantErrSub string
		wantFields map[string]any
		wantNoKey  []string
	}{
		{
			name: "happy path with full event",
			args: map[string]any{"eventId": "evt-1"},
			getEventFn: func(calendarID, eventID string) (*calendar.Event, error) {
				return fullEvent, nil
			},
			wantFields: map[string]any{
				"success":     true,
				"eventId":     "evt-1",
				"summary":     "Quarterly review",
				"description": "Discuss Q2 metrics",
				"location":    "HQ-3F",
				"htmlLink":    "https://example.com/evt-1",
				"startTime":   "2026-05-23T10:00:00Z",
				"endTime":     "2026-05-23T11:00:00Z",
			},
		},
		{
			name: "minimal event omits empty optional fields",
			args: map[string]any{"eventId": "evt-min"},
			getEventFn: func(calendarID, eventID string) (*calendar.Event, error) {
				return minimalEvent, nil
			},
			wantFields: map[string]any{
				"success": true,
				"eventId": "evt-min",
				"summary": "Bare event",
			},
			wantNoKey: []string{"description", "location", "htmlLink", "attendees"},
		},
		{
			name:       "missing eventId returns error",
			args:       map[string]any{},
			wantErr:    true,
			wantErrSub: "eventId is required",
		},
		{
			name: "GetEvent error is wrapped and returned",
			args: map[string]any{"eventId": "evt-1"},
			getEventFn: func(calendarID, eventID string) (*calendar.Event, error) {
				return nil, errors.New("not found")
			},
			wantErr:    true,
			wantErrSub: "failed to get calendar event",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubCalendarService{getEventFn: tc.getEventFn}
			tool := &GetCalendarEventTool{logger: zap.NewNop(), google: stub}
			result, err := tool.GetCalendarEventHandler(context.Background(), tc.args)

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
			for key, want := range tc.wantFields {
				if parsed[key] != want {
					t.Errorf("%s = %v, want %v", key, parsed[key], want)
				}
			}
			for _, key := range tc.wantNoKey {
				if _, ok := parsed[key]; ok {
					t.Errorf("expected key %q to be absent, but found %v", key, parsed[key])
				}
			}
		})
	}
}
