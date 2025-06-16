package toolbox

import (
	"encoding/json"
	"fmt"
	"time"
)

// Mock Response Helpers
// This file contains all mock implementations for Google Calendar operations.
// These methods provide realistic mock responses for testing and development environments.

// getMockEvents returns mock calendar events for testing
func (g *GoogleCalendarTools) getMockEvents() string {
	mockEvents := []map[string]interface{}{
		{
			"id":      "mock-event-1",
			"summary": "Team Meeting",
			"start":   map[string]string{"dateTime": time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
			"end":     map[string]string{"dateTime": time.Now().Add(2 * time.Hour).Format(time.RFC3339)},
		},
		{
			"id":      "mock-event-2",
			"summary": "Lunch with Client",
			"start":   map[string]string{"dateTime": time.Now().Add(4 * time.Hour).Format(time.RFC3339)},
			"end":     map[string]string{"dateTime": time.Now().Add(5 * time.Hour).Format(time.RFC3339)},
		},
	}
	result := map[string]interface{}{
		"events": mockEvents,
		"count":  len(mockEvents),
		"mock":   true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

// getMockCreateEvent returns a mock response for event creation
func (g *GoogleCalendarTools) getMockCreateEvent(args map[string]interface{}) string {
	result := map[string]interface{}{
		"success":   true,
		"eventId":   fmt.Sprintf("mock-created-event-%d", time.Now().Unix()),
		"message":   "Event would be created (mock mode)",
		"summary":   args["summary"],
		"startTime": args["startTime"],
		"endTime":   args["endTime"],
		"mock":      true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

// getMockUpdateEvent returns a mock response for event updates
func (g *GoogleCalendarTools) getMockUpdateEvent(args map[string]interface{}) string {
	result := map[string]interface{}{
		"success": true,
		"eventId": args["eventId"],
		"message": "Event would be updated (mock mode)",
		"mock":    true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

// getMockDeleteEvent returns a mock response for event deletion
func (g *GoogleCalendarTools) getMockDeleteEvent(args map[string]interface{}) string {
	result := map[string]interface{}{
		"success": true,
		"eventId": args["eventId"],
		"message": "Event would be deleted (mock mode)",
		"mock":    true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

// getMockGetEvent returns a mock response for getting event details
func (g *GoogleCalendarTools) getMockGetEvent(args map[string]interface{}) string {
	result := map[string]interface{}{
		"success": true,
		"event": map[string]interface{}{
			"id":      args["eventId"],
			"summary": "Mock Event",
			"start":   map[string]string{"dateTime": time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
			"end":     map[string]string{"dateTime": time.Now().Add(2 * time.Hour).Format(time.RFC3339)},
		},
		"mock": true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

// getMockAvailableTime returns mock available time slots
func (g *GoogleCalendarTools) getMockAvailableTime(args map[string]interface{}) string {
	duration := 60
	if val, ok := args["duration"].(float64); ok {
		duration = int(val)
	}

	start, _ := time.Parse(time.RFC3339, args["startDate"].(string))
	slots := []map[string]string{
		{
			"start": start.Add(2 * time.Hour).Format(time.RFC3339),
			"end":   start.Add(2*time.Hour + time.Duration(duration)*time.Minute).Format(time.RFC3339),
		},
		{
			"start": start.Add(4 * time.Hour).Format(time.RFC3339),
			"end":   start.Add(4*time.Hour + time.Duration(duration)*time.Minute).Format(time.RFC3339),
		},
	}
	result := map[string]interface{}{
		"availableSlots": slots,
		"count":          len(slots),
		"duration":       duration,
		"mock":           true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}

// getMockConflicts returns mock conflict checking results
func (g *GoogleCalendarTools) getMockConflicts(args map[string]interface{}) string {
	result := map[string]interface{}{
		"hasConflicts":   false,
		"conflictCount":  0,
		"conflictEvents": []interface{}{},
		"timeRange": map[string]string{
			"start": args["startTime"].(string),
			"end":   args["endTime"].(string),
		},
		"mock": true,
	}
	response, _ := json.Marshal(result)
	return string(response)
}
