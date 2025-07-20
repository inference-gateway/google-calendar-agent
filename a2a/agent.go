package a2a

import (
	adk "github.com/inference-gateway/a2a/adk"
	config "github.com/inference-gateway/a2a/adk/server/config"
)

// GetAgentCard returns custom agent card information
func GetAgentCard(baseConfig config.Config) adk.AgentCard {
	return adk.AgentCard{
		Name:        baseConfig.AgentName,
		Description: baseConfig.AgentDescription,
		Version:     "0.4.7",
		URL:         baseConfig.AgentURL,
		Provider: &adk.AgentProvider{
			Organization: "Inference Gateway",
			URL:          "https://github.com/inference-gateway",
		},
		Capabilities: adk.AgentCapabilities{
			Streaming:              boolPtr(true),
			PushNotifications:      boolPtr(true),
			StateTransitionHistory: boolPtr(true),
		},
		DefaultInputModes:  []string{"text"},
		DefaultOutputModes: []string{"text", "json"},
		Skills: []adk.AgentSkill{
			{
				ID:          "list_calendar_events",
				Name:        "List Calendar Events",
				Description: "List upcoming events from Google Calendar",
				Tags:        []string{"calendar", "events", "list"},
				Examples: []string{
					"List my calendar events for today",
					"Show me my upcoming meetings",
					"What events do I have this week?",
				},
				InputModes:  []string{"text"},
				OutputModes: []string{"text", "json"},
			},
			{
				ID:          "create_calendar_event",
				Name:        "Create Calendar Event",
				Description: "Create a new event in Google Calendar",
				Tags:        []string{"calendar", "events", "create"},
				Examples: []string{
					"Create a meeting for tomorrow at 2pm",
					"Schedule a dentist appointment for next Friday",
					"Add a reminder for my anniversary",
				},
				InputModes:  []string{"text"},
				OutputModes: []string{"text", "json"},
			},
			{
				ID:          "update_calendar_event",
				Name:        "Update Calendar Event",
				Description: "Update an existing event in Google Calendar",
				Tags:        []string{"calendar", "events", "update"},
				Examples: []string{
					"Change my 3pm meeting to 4pm",
					"Update the location of my appointment",
					"Add attendees to my meeting",
				},
				InputModes:  []string{"text"},
				OutputModes: []string{"text", "json"},
			},
			{
				ID:          "delete_calendar_event",
				Name:        "Delete Calendar Event",
				Description: "Delete an event from Google Calendar",
				Tags:        []string{"calendar", "events", "delete"},
				Examples: []string{
					"Cancel my 2pm meeting",
					"Delete tomorrow's dentist appointment",
					"Remove the duplicate event",
				},
				InputModes:  []string{"text"},
				OutputModes: []string{"text", "json"},
			},
			{
				ID:          "get_calendar_event",
				Name:        "Get Calendar Event",
				Description: "Get details of a specific event in Google Calendar",
				Tags:        []string{"calendar", "events", "details"},
				Examples: []string{
					"Get details of my 3pm meeting",
					"Show me information about my appointment",
					"What's the location of my next meeting?",
				},
				InputModes:  []string{"text"},
				OutputModes: []string{"text", "json"},
			},
			{
				ID:          "find_available_time",
				Name:        "Find Available Time",
				Description: "Find free time slots in Google Calendar",
				Tags:        []string{"calendar", "availability", "scheduling"},
				Examples: []string{
					"When am I free this week?",
					"Find a 1-hour slot for a meeting",
					"What's my availability tomorrow?",
				},
				InputModes:  []string{"text"},
				OutputModes: []string{"text", "json"},
			},
			{
				ID:          "check_conflicts",
				Name:        "Check Conflicts",
				Description: "Check for scheduling conflicts in Google Calendar",
				Tags:        []string{"calendar", "conflicts", "scheduling"},
				Examples: []string{
					"Do I have any conflicts this week?",
					"Check if I'm double-booked tomorrow",
					"Are there any overlapping meetings?",
				},
				InputModes:  []string{"text"},
				OutputModes: []string{"text", "json"},
			},
		},
	}
}

// boolPtr returns a pointer to a boolean value
func boolPtr(b bool) *bool {
	return &b
}
