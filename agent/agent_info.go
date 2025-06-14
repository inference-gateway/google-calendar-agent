package agent

import (
	"github.com/inference-gateway/a2a/adk/server/config"
)

// GoogleCalendarAgentInfo provides custom agent information for the Google Calendar agent
type GoogleCalendarAgentInfo struct{}

// GetAgentCard returns custom agent card information
func (g *GoogleCalendarAgentInfo) GetAgentCard(baseConfig config.Config) interface{} {
	return map[string]interface{}{
		"name":        baseConfig.AgentName,
		"description": baseConfig.AgentDescription,
		"version":     "1.0.0",
		"capabilities": map[string]interface{}{
			"streaming":         false,
			"taskManagement":    true,
			"pushNotifications": false,
		},
		"metadata": map[string]interface{}{
			"specialization": "google-calendar-management",
			"provider":       "Google Calendar API",
			"features": []string{
				"event-listing",
				"event-creation",
				"event-modification",
				"conflict-detection",
				"availability-search",
			},
		},
		"tools": []map[string]interface{}{
			{
				"name":        "list_calendar_events",
				"description": "List upcoming events from Google Calendar",
				"category":    "calendar",
			},
			{
				"name":        "create_calendar_event",
				"description": "Create a new event in Google Calendar",
				"category":    "calendar",
			},
			{
				"name":        "update_calendar_event",
				"description": "Update an existing event in Google Calendar",
				"category":    "calendar",
			},
			{
				"name":        "delete_calendar_event",
				"description": "Delete an event from Google Calendar",
				"category":    "calendar",
			},
			{
				"name":        "get_calendar_event",
				"description": "Get details of a specific event",
				"category":    "calendar",
			},
			{
				"name":        "find_available_time",
				"description": "Find available time slots in the calendar",
				"category":    "scheduling",
			},
			{
				"name":        "check_conflicts",
				"description": "Check for scheduling conflicts",
				"category":    "scheduling",
			},
		},
	}
}
