package a2a

import (
	"fmt"

	uuid "github.com/google/uuid"
	calendar "google.golang.org/api/calendar/v3"
)

// CalendarEventResponse represents a structured response for calendar events
type CalendarEventResponse struct {
	Event   *calendar.Event   `json:"event,omitempty"`
	Events  []*calendar.Event `json:"events,omitempty"`
	Message string            `json:"message"`
	Success bool              `json:"success"`
}

// CalendarAvailabilityResponse represents a response for availability queries
type CalendarAvailabilityResponse struct {
	AvailableSlots []TimeSlot `json:"availableSlots"`
	Message        string     `json:"message"`
	Success        bool       `json:"success"`
}

// TimeSlot represents an available time slot
type TimeSlot struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// CalendarConflictResponse represents a response for conflict checks
type CalendarConflictResponse struct {
	Conflicts []ConflictInfo `json:"conflicts"`
	Message   string         `json:"message"`
	Success   bool           `json:"success"`
}

// ConflictInfo represents information about a scheduling conflict
type ConflictInfo struct {
	Event        *calendar.Event `json:"event"`
	ConflictType string          `json:"conflictType"`
	Details      string          `json:"details"`
}

// CreateTextPart creates an A2A TextPart with the given content
func CreateTextPart(text string) TextPart {
	return TextPart{
		Kind: "text",
		Text: text,
	}
}

// CreateDataPart creates an A2A DataPart with structured data
func CreateDataPart(data map[string]interface{}) DataPart {
	return DataPart{
		Kind: "data",
		Data: data,
	}
}

// CreateSuccessMessage creates a success message using A2A types
func CreateSuccessMessage(taskID, content string, data map[string]interface{}) Message {
	parts := []Part{CreateTextPart(content)}

	if data != nil {
		parts = append(parts, CreateDataPart(data))
	}

	return Message{
		Kind:      "message",
		MessageID: generateMessageID(),
		Role:      "assistant",
		TaskID:    &taskID,
		Parts:     parts,
	}
}

// CreateErrorMessage creates an error message using A2A types
func CreateErrorMessage(taskID, errorMsg string) Message {
	return Message{
		Kind:      "message",
		MessageID: generateMessageID(),
		Role:      "assistant",
		TaskID:    &taskID,
		Parts: []Part{
			CreateTextPart("❌ Error: " + errorMsg),
		},
	}
}

// CreateTaskStatus creates a task status using A2A types
func CreateTaskStatus(state TaskState, message *Message) TaskStatus {
	return TaskStatus{
		State:   state,
		Message: message,
	}
}

// CreateCalendarEventArtifact creates an A2A artifact for calendar events
func CreateCalendarEventArtifact(event *calendar.Event, artifactType string) Artifact {
	metadata := map[string]interface{}{
		"eventId":  event.Id,
		"summary":  event.Summary,
		"created":  event.Created,
		"updated":  event.Updated,
		"status":   event.Status,
		"htmlLink": event.HtmlLink,
	}

	return Artifact{
		ArtifactID:  "artifact_" + event.Id,
		Name:        &event.Summary,
		Description: &artifactType,
		Metadata:    metadata,
		Parts: []Part{
			CreateDataPart(map[string]interface{}{
				"event": event,
			}),
		},
	}
}

// CreateCalendarEventsArtifact creates an A2A artifact for multiple calendar events
func CreateCalendarEventsArtifact(events []*calendar.Event, description string) Artifact {
	artifactName := fmt.Sprintf("Calendar Events (%d)", len(events))

	metadata := map[string]interface{}{
		"eventCount": len(events),
		"type":       "event_list",
	}

	return Artifact{
		ArtifactID:  "artifact_events_" + generateUniqueID(),
		Name:        &artifactName,
		Description: &description,
		Metadata:    metadata,
		Parts: []Part{
			CreateDataPart(map[string]interface{}{
				"events": events,
			}),
		},
	}
}

// CreateAvailabilityArtifact creates an A2A artifact for availability information
func CreateAvailabilityArtifact(availableSlots []TimeSlot, description string) Artifact {
	artifactName := fmt.Sprintf("Available Time Slots (%d)", len(availableSlots))

	metadata := map[string]interface{}{
		"slotCount": len(availableSlots),
		"type":      "availability",
	}

	return Artifact{
		ArtifactID:  "artifact_availability_" + generateUniqueID(),
		Name:        &artifactName,
		Description: &description,
		Metadata:    metadata,
		Parts: []Part{
			CreateDataPart(map[string]interface{}{
				"availableSlots": availableSlots,
			}),
		},
	}
}

// CreateTask creates a complete A2A task
func CreateTask(contextID, taskID string, status TaskStatus, artifacts []Artifact, history []Message) Task {
	return Task{
		ID:        taskID,
		ContextID: contextID,
		Kind:      "task",
		Status:    status,
		Artifacts: artifacts,
		History:   history,
	}
}

// Helper function to generate unique message IDs using UUID
func generateMessageID() string {
	return "msg_" + uuid.New().String()
}

// Helper function to generate unique IDs using UUID
func generateUniqueID() string {
	return uuid.New().String()
}
