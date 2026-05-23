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

func TestCheckConflictsHandler(t *testing.T) {
	oneConflict := []*calendar.Event{
		{
			Id:       "conflict-1",
			Summary:  "Overlapping meeting",
			Location: "Room 1",
			Start:    &calendar.EventDateTime{DateTime: "2026-05-23T10:30:00Z"},
			End:      &calendar.EventDateTime{DateTime: "2026-05-23T11:30:00Z"},
		},
	}

	tests := []struct {
		name             string
		args             map[string]any
		conflicts        []*calendar.Event
		checkErr         error
		wantErr          bool
		wantErrSub       string
		wantHasConflicts bool
		wantCount        int
		wantFirstSummary string
	}{
		{
			name: "happy path with no conflicts",
			args: map[string]any{
				"startTime": "2026-05-23T10:00:00Z",
				"endTime":   "2026-05-23T11:00:00Z",
			},
			conflicts:        nil,
			wantHasConflicts: false,
			wantCount:        0,
		},
		{
			name: "happy path with one conflict",
			args: map[string]any{
				"startTime": "2026-05-23T10:00:00Z",
				"endTime":   "2026-05-23T11:00:00Z",
			},
			conflicts:        oneConflict,
			wantHasConflicts: true,
			wantCount:        1,
			wantFirstSummary: "Overlapping meeting",
		},
		{
			name:       "missing startTime returns error",
			args:       map[string]any{"endTime": "2026-05-23T11:00:00Z"},
			wantErr:    true,
			wantErrSub: "startTime is required",
		},
		{
			name:       "missing endTime returns error",
			args:       map[string]any{"startTime": "2026-05-23T10:00:00Z"},
			wantErr:    true,
			wantErrSub: "endTime is required",
		},
		{
			name: "invalid startTime format returns error",
			args: map[string]any{
				"startTime": "tomorrow",
				"endTime":   "2026-05-23T11:00:00Z",
			},
			wantErr:    true,
			wantErrSub: "invalid startTime format",
		},
		{
			name: "invalid endTime format returns error",
			args: map[string]any{
				"startTime": "2026-05-23T10:00:00Z",
				"endTime":   "later",
			},
			wantErr:    true,
			wantErrSub: "invalid endTime format",
		},
		{
			name: "CheckConflicts error is wrapped and returned",
			args: map[string]any{
				"startTime": "2026-05-23T10:00:00Z",
				"endTime":   "2026-05-23T11:00:00Z",
			},
			checkErr:   errors.New("api timeout"),
			wantErr:    true,
			wantErrSub: "failed to check conflicts",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubCalendarService{
				checkConflictsFn: func(calendarID string, startTime, endTime time.Time) ([]*calendar.Event, error) {
					if tc.checkErr != nil {
						return nil, tc.checkErr
					}
					return tc.conflicts, nil
				},
			}
			tool := &CheckConflictsTool{logger: zap.NewNop(), google: stub}
			result, err := tool.CheckConflictsHandler(context.Background(), tc.args)

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
				Success       bool             `json:"success"`
				HasConflicts  bool             `json:"hasConflicts"`
				ConflictCount int              `json:"conflictCount"`
				Conflicts     []map[string]any `json:"conflicts"`
			}
			if err := json.Unmarshal([]byte(result), &parsed); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}
			if !parsed.Success {
				t.Errorf("success = false, want true")
			}
			if parsed.HasConflicts != tc.wantHasConflicts {
				t.Errorf("hasConflicts = %v, want %v", parsed.HasConflicts, tc.wantHasConflicts)
			}
			if parsed.ConflictCount != tc.wantCount {
				t.Errorf("conflictCount = %d, want %d", parsed.ConflictCount, tc.wantCount)
			}
			if tc.wantFirstSummary != "" && len(parsed.Conflicts) > 0 {
				if parsed.Conflicts[0]["summary"] != tc.wantFirstSummary {
					t.Errorf("conflicts[0].summary = %v, want %v", parsed.Conflicts[0]["summary"], tc.wantFirstSummary)
				}
			}
		})
	}
}
