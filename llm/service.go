package llm

import (
	"context"
	"time"
)

// Service defines the interface for LLM operations
//
//go:generate counterfeiter -generate
//counterfeiter:generate -o mocks . Service
type Service interface {
	// ProcessNaturalLanguage processes natural language input and returns structured output
	ProcessNaturalLanguage(ctx context.Context, input string) (*ProcessingResult, error)

	// IsEnabled returns true if the LLM service is enabled
	IsEnabled() bool

	// GetProvider returns the configured provider name
	GetProvider() string

	// GetModel returns the configured model name
	GetModel() string
}

// ProcessingResult represents the result of natural language processing
type ProcessingResult struct {
	// Intent represents the detected intent (e.g., "create_event", "list_events", "update_event", "delete_event")
	Intent string `json:"intent"`

	// Confidence represents the confidence score (0.0 to 1.0)
	Confidence float64 `json:"confidence"`

	// Parameters contains extracted parameters from the input
	Parameters map[string]interface{} `json:"parameters"`

	// Response is the formatted response to return to the user
	Response string `json:"response"`

	// RawResponse is the raw response from the LLM
	RawResponse string `json:"raw_response"`

	// ProcessingTime is the time taken to process the request
	ProcessingTime time.Duration `json:"processing_time"`

	// TokensUsed represents token usage information
	TokensUsed *TokenUsage `json:"tokens_used,omitempty"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	// PromptTokens is the number of tokens in the prompt
	PromptTokens int `json:"prompt_tokens"`

	// CompletionTokens is the number of tokens in the completion
	CompletionTokens int `json:"completion_tokens"`

	// TotalTokens is the total number of tokens used
	TotalTokens int `json:"total_tokens"`
}

// EventRequest represents a structured calendar event request
type EventRequest struct {
	// Action is the action to perform (create, update, delete, list)
	Action string `json:"action"`

	// Title is the event title
	Title string `json:"title,omitempty"`

	// Description is the event description
	Description string `json:"description,omitempty"`

	// StartTime is the event start time
	StartTime *time.Time `json:"start_time,omitempty"`

	// EndTime is the event end time
	EndTime *time.Time `json:"end_time,omitempty"`

	// Location is the event location
	Location string `json:"location,omitempty"`

	// Attendees is a list of attendee email addresses
	Attendees []string `json:"attendees,omitempty"`

	// EventID is the ID of an existing event (for updates/deletes)
	EventID string `json:"event_id,omitempty"`

	// TimeRange specifies the time range for listing events
	TimeRange *TimeRange `json:"time_range,omitempty"`
}

// TimeRange represents a time range for queries
type TimeRange struct {
	// Start is the start of the time range
	Start time.Time `json:"start"`

	// End is the end of the time range
	End time.Time `json:"end"`
}
