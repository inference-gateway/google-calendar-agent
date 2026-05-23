package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	zap "go.uber.org/zap"
)

func TestDeleteCalendarEventHandler(t *testing.T) {
	tests := []struct {
		name          string
		args          map[string]any
		deleteEventFn func(calendarID, eventID string) error
		wantErr       bool
		wantErrSub    string
		wantEventID   string
	}{
		{
			name: "happy path deletes the event",
			args: map[string]any{"eventId": "evt-1"},
			deleteEventFn: func(calendarID, eventID string) error {
				if eventID != "evt-1" {
					t.Errorf("DeleteEvent called with eventID=%q, want evt-1", eventID)
				}
				return nil
			},
			wantEventID: "evt-1",
		},
		{
			name:       "missing eventId returns error",
			args:       map[string]any{},
			wantErr:    true,
			wantErrSub: "eventId is required",
		},
		{
			name:       "empty eventId returns error",
			args:       map[string]any{"eventId": ""},
			wantErr:    true,
			wantErrSub: "eventId is required",
		},
		{
			name:       "non-string eventId returns error",
			args:       map[string]any{"eventId": 42},
			wantErr:    true,
			wantErrSub: "eventId is required",
		},
		{
			name: "DeleteEvent error is wrapped and returned",
			args: map[string]any{"eventId": "evt-1"},
			deleteEventFn: func(calendarID, eventID string) error {
				return errors.New("permission denied")
			},
			wantErr:    true,
			wantErrSub: "failed to delete calendar event",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubCalendarService{deleteEventFn: tc.deleteEventFn}
			tool := &DeleteCalendarEventTool{logger: zap.NewNop(), google: stub}
			result, err := tool.DeleteCalendarEventHandler(context.Background(), tc.args)

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
			if parsed["eventId"] != tc.wantEventID {
				t.Errorf("eventId = %v, want %v", parsed["eventId"], tc.wantEventID)
			}
		})
	}
}
