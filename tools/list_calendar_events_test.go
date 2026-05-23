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

func TestListCalendarEventsHandler(t *testing.T) {
	mockEvents := []*calendar.Event{
		{
			Id:          "e1",
			Summary:     "Standup",
			Description: "Daily sync",
			Status:      "confirmed",
			Start:       &calendar.EventDateTime{DateTime: "2026-05-23T10:00:00Z"},
			End:         &calendar.EventDateTime{DateTime: "2026-05-23T10:30:00Z"},
		},
		{
			Id:          "e2",
			Summary:     "Design review",
			Description: "API discussion",
			Status:      "confirmed",
			Start:       &calendar.EventDateTime{DateTime: "2026-05-23T14:00:00Z"},
			End:         &calendar.EventDateTime{DateTime: "2026-05-23T15:00:00Z"},
		},
		{
			Id:      "e3",
			Summary: "Lunch",
			Status:  "confirmed",
			Start:   &calendar.EventDateTime{DateTime: "2026-05-23T12:00:00Z"},
			End:     &calendar.EventDateTime{DateTime: "2026-05-23T13:00:00Z"},
		},
	}

	tests := []struct {
		name         string
		args         map[string]any
		events       []*calendar.Event
		listErr      error
		wantErr      bool
		wantErrSub   string
		wantCount    int
		wantTimeMin  string
		wantTimeMax  string
		wantSummary0 string
	}{
		{
			name:        "happy path returns all events",
			args:        map[string]any{},
			events:      mockEvents,
			wantCount:   3,
			wantTimeMax: "0001-01-01T00:00:00Z",
		},
		{
			name:        "maxResults caps the result count",
			args:        map[string]any{"maxResults": float64(2)},
			events:      mockEvents,
			wantCount:   2,
			wantTimeMax: "0001-01-01T00:00:00Z",
		},
		{
			name:         "query filters by case-insensitive substring of summary or description",
			args:         map[string]any{"query": "DESIGN"},
			events:       mockEvents,
			wantCount:    1,
			wantSummary0: "Design review",
		},
		{
			name:        "query matches description",
			args:        map[string]any{"query": "api"},
			events:      mockEvents,
			wantCount:   1,
			wantTimeMax: "0001-01-01T00:00:00Z",
		},
		{
			name: "valid timeMin and timeMax are forwarded to ListEvents",
			args: map[string]any{
				"timeMin": "2026-05-23T00:00:00Z",
				"timeMax": "2026-05-23T23:59:59Z",
			},
			events:      mockEvents,
			wantCount:   3,
			wantTimeMin: "2026-05-23T00:00:00Z",
			wantTimeMax: "2026-05-23T23:59:59Z",
		},
		{
			name:       "wrong-typed maxResults returns error",
			args:       map[string]any{"maxResults": "ten"},
			wantErr:    true,
			wantErrSub: "maxResults must be a number",
		},
		{
			name:       "wrong-typed query returns error",
			args:       map[string]any{"query": 42},
			wantErr:    true,
			wantErrSub: "query must be a string",
		},
		{
			name:       "wrong-typed timeMin returns error",
			args:       map[string]any{"timeMin": 12345},
			wantErr:    true,
			wantErrSub: "timeMin must be a string",
		},
		{
			name:       "invalid RFC3339 timeMin returns error",
			args:       map[string]any{"timeMin": "tomorrow"},
			wantErr:    true,
			wantErrSub: "invalid timeMin format",
		},
		{
			name:       "invalid RFC3339 timeMax returns error",
			args:       map[string]any{"timeMax": "next-week"},
			wantErr:    true,
			wantErrSub: "invalid timeMax format",
		},
		{
			name:       "ListEvents failure is wrapped",
			args:       map[string]any{},
			listErr:    errors.New("api 500"),
			wantErr:    true,
			wantErrSub: "failed to list calendar events",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedTimeMin, capturedTimeMax time.Time
			stub := &stubCalendarService{
				listEventsFn: func(calendarID string, timeMin, timeMax time.Time) ([]*calendar.Event, error) {
					capturedTimeMin = timeMin
					capturedTimeMax = timeMax
					if tc.listErr != nil {
						return nil, tc.listErr
					}
					return tc.events, nil
				},
			}
			tool := &ListCalendarEventsTool{logger: zap.NewNop(), google: stub}
			result, err := tool.ListCalendarEventsHandler(context.Background(), tc.args)

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
				Success bool             `json:"success"`
				Count   int              `json:"count"`
				Events  []map[string]any `json:"events"`
			}
			if err := json.Unmarshal([]byte(result), &parsed); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}
			if !parsed.Success {
				t.Errorf("success = false, want true")
			}
			if parsed.Count != tc.wantCount {
				t.Errorf("count = %d, want %d (events=%+v)", parsed.Count, tc.wantCount, parsed.Events)
			}
			if tc.wantSummary0 != "" && len(parsed.Events) > 0 {
				if parsed.Events[0]["summary"] != tc.wantSummary0 {
					t.Errorf("events[0].summary = %v, want %v", parsed.Events[0]["summary"], tc.wantSummary0)
				}
			}
			if tc.wantTimeMin != "" {
				if got := capturedTimeMin.Format(time.RFC3339); got != tc.wantTimeMin {
					t.Errorf("timeMin forwarded = %s, want %s", got, tc.wantTimeMin)
				}
			}
			if tc.wantTimeMax != "" {
				if got := capturedTimeMax.Format(time.RFC3339); got != tc.wantTimeMax {
					t.Errorf("timeMax forwarded = %s, want %s", got, tc.wantTimeMax)
				}
			}
		})
	}
}
