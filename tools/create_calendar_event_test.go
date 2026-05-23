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

func TestCreateCalendarEventHandler(t *testing.T) {
	echoCreate := func(calendarID string, event *calendar.Event) (*calendar.Event, error) {
		event.Id = "evt-created"
		event.HtmlLink = "https://example.com/evt-created"
		return event, nil
	}

	tests := []struct {
		name          string
		args          map[string]any
		createEventFn func(calendarID string, event *calendar.Event) (*calendar.Event, error)
		wantErr       bool
		wantErrSub    string
		wantSummary   string
		wantAttendees []any
		wantDesc      string
		wantLocation  string
	}{
		{
			name: "happy path with required fields only",
			args: map[string]any{
				"summary":   "Standup",
				"startTime": "2026-05-23T10:00:00Z",
				"endTime":   "2026-05-23T10:30:00Z",
			},
			createEventFn: echoCreate,
			wantSummary:   "Standup",
		},
		{
			name: "happy path with all optional fields",
			args: map[string]any{
				"summary":     "Design review",
				"startTime":   "2026-05-23T14:00:00Z",
				"endTime":     "2026-05-23T15:00:00Z",
				"description": "Discuss new API",
				"location":    "Room 1",
				"attendees":   []any{"a@example.com", "b@example.com"},
			},
			createEventFn: echoCreate,
			wantSummary:   "Design review",
			wantDesc:      "Discuss new API",
			wantLocation:  "Room 1",
			wantAttendees: []any{"a@example.com", "b@example.com"},
		},
		{
			name: "non-string attendee entries are silently skipped",
			args: map[string]any{
				"summary":   "Mixed attendees",
				"startTime": "2026-05-23T14:00:00Z",
				"endTime":   "2026-05-23T15:00:00Z",
				"attendees": []any{"a@example.com", 42, "b@example.com"},
			},
			createEventFn: echoCreate,
			wantSummary:   "Mixed attendees",
			wantAttendees: []any{"a@example.com", "b@example.com"},
		},
		{
			name:       "missing summary returns error",
			args:       map[string]any{"startTime": "2026-05-23T10:00:00Z", "endTime": "2026-05-23T11:00:00Z"},
			wantErr:    true,
			wantErrSub: "summary is required",
		},
		{
			name:       "missing startTime returns error",
			args:       map[string]any{"summary": "s", "endTime": "2026-05-23T11:00:00Z"},
			wantErr:    true,
			wantErrSub: "startTime is required",
		},
		{
			name:       "missing endTime returns error",
			args:       map[string]any{"summary": "s", "startTime": "2026-05-23T10:00:00Z"},
			wantErr:    true,
			wantErrSub: "endTime is required",
		},
		{
			name: "wrong-typed description returns error and does not panic",
			args: map[string]any{
				"summary":     "s",
				"startTime":   "2026-05-23T10:00:00Z",
				"endTime":     "2026-05-23T11:00:00Z",
				"description": 42,
			},
			wantErr:    true,
			wantErrSub: "description must be a string",
		},
		{
			name: "wrong-typed location returns error and does not panic",
			args: map[string]any{
				"summary":   "s",
				"startTime": "2026-05-23T10:00:00Z",
				"endTime":   "2026-05-23T11:00:00Z",
				"location":  []any{"a"},
			},
			wantErr:    true,
			wantErrSub: "location must be a string",
		},
		{
			name: "CreateEvent error is wrapped and returned",
			args: map[string]any{
				"summary":   "s",
				"startTime": "2026-05-23T10:00:00Z",
				"endTime":   "2026-05-23T11:00:00Z",
			},
			createEventFn: func(calendarID string, event *calendar.Event) (*calendar.Event, error) {
				return nil, errors.New("quota exceeded")
			},
			wantErr:    true,
			wantErrSub: "failed to create calendar event",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubCalendarService{createEventFn: tc.createEventFn}
			tool := &CreateCalendarEventTool{logger: zap.NewNop(), google: stub}
			result, err := tool.CreateCalendarEventHandler(context.Background(), tc.args)

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
			if parsed["success"] != true {
				t.Errorf("success = %v, want true", parsed["success"])
			}
			if parsed["summary"] != tc.wantSummary {
				t.Errorf("summary = %v, want %v", parsed["summary"], tc.wantSummary)
			}
			if tc.wantDesc != "" && parsed["description"] != tc.wantDesc {
				t.Errorf("description = %v, want %v", parsed["description"], tc.wantDesc)
			}
			if tc.wantLocation != "" && parsed["location"] != tc.wantLocation {
				t.Errorf("location = %v, want %v", parsed["location"], tc.wantLocation)
			}
			if tc.wantAttendees != nil {
				got, ok := parsed["attendees"].([]any)
				if !ok {
					t.Fatalf("attendees missing or wrong type: %T %v", parsed["attendees"], parsed["attendees"])
				}
				if len(got) != len(tc.wantAttendees) {
					t.Errorf("attendees len = %d, want %d", len(got), len(tc.wantAttendees))
				}
				for i := range tc.wantAttendees {
					if i < len(got) && got[i] != tc.wantAttendees[i] {
						t.Errorf("attendee[%d] = %v, want %v", i, got[i], tc.wantAttendees[i])
					}
				}
			}
		})
	}
}
