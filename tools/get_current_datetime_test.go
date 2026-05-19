package tools

import (
	"testing"
)

func TestResolveTimezone(t *testing.T) {
	tests := []struct {
		name       string
		gctz       string
		tz         string
		wantName   string
		wantSource string
	}{
		{
			name:       "GOOGLE_CALENDAR_TIMEZONE wins over TZ",
			gctz:       "Europe/Berlin",
			tz:         "America/Los_Angeles",
			wantName:   "Europe/Berlin",
			wantSource: "GOOGLE_CALENDAR_TIMEZONE",
		},
		{
			name:       "falls back to TZ when GOOGLE_CALENDAR_TIMEZONE unset",
			gctz:       "",
			tz:         "America/Los_Angeles",
			wantName:   "America/Los_Angeles",
			wantSource: "TZ",
		},
		{
			name:       "falls back to UTC when neither is set",
			gctz:       "",
			tz:         "",
			wantName:   "UTC",
			wantSource: "default",
		},
		{
			name:       "invalid GOOGLE_CALENDAR_TIMEZONE falls through to TZ",
			gctz:       "Not/A_Real_Zone",
			tz:         "Europe/Berlin",
			wantName:   "Europe/Berlin",
			wantSource: "TZ",
		},
		{
			name:       "CET shorthand is accepted",
			gctz:       "CET",
			tz:         "",
			wantName:   "CET",
			wantSource: "GOOGLE_CALENDAR_TIMEZONE",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GOOGLE_CALENDAR_TIMEZONE", tc.gctz)
			t.Setenv("TZ", tc.tz)

			loc, name, source := resolveTimezone()
			if loc == nil {
				t.Fatal("loc is nil")
			}
			if name != tc.wantName {
				t.Errorf("name = %q, want %q", name, tc.wantName)
			}
			if source != tc.wantSource {
				t.Errorf("source = %q, want %q", source, tc.wantSource)
			}
		})
	}
}
