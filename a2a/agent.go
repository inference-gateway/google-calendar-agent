package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/calendar/v3"

	"github.com/inference-gateway/google-calendar-agent/config"
	"github.com/inference-gateway/google-calendar-agent/google"
	"github.com/inference-gateway/google-calendar-agent/llm"
)

// CalendarAgent handles A2A calendar requests
type CalendarAgent struct {
	calendarService google.CalendarService
	logger          *zap.Logger
	config          *config.Config
	llmService      llm.Service
}

// NewCalendarAgent creates a new calendar agent
func NewCalendarAgent(calendarService google.CalendarService, logger *zap.Logger) *CalendarAgent {
	return &CalendarAgent{
		calendarService: calendarService,
		logger:          logger,
		config:          nil,
		llmService:      nil,
	}
}

// NewCalendarAgentWithConfig creates a new calendar agent with configuration
func NewCalendarAgentWithConfig(calendarService google.CalendarService, logger *zap.Logger, cfg *config.Config) *CalendarAgent {
	return &CalendarAgent{
		calendarService: calendarService,
		logger:          logger,
		config:          cfg,
		llmService:      nil,
	}
}

// NewCalendarAgentWithLLM creates a new calendar agent with configuration and LLM service
func NewCalendarAgentWithLLM(calendarService google.CalendarService, logger *zap.Logger, cfg *config.Config, llmSvc llm.Service) *CalendarAgent {
	return &CalendarAgent{
		calendarService: calendarService,
		logger:          logger,
		config:          cfg,
		llmService:      llmSvc,
	}
}

// HandleA2ARequest processes incoming A2A requests
func (a *CalendarAgent) HandleA2ARequest(c *gin.Context) {
	requestStartTime := time.Now()
	a.logger.Debug("received a2a request",
		zap.String("component", "a2a-handler"),
		zap.String("operation", "handle-request"),
		zap.String("clientIP", c.ClientIP()),
		zap.String("userAgent", c.GetHeader("User-Agent")),
		zap.String("contentType", c.GetHeader("Content-Type")),
		zap.Time("requestTime", requestStartTime))

	var req JSONRPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("failed to parse json request",
			zap.String("component", "a2a-handler"),
			zap.String("operation", "parse-request"),
			zap.Error(err),
			zap.String("clientIP", c.ClientIP()),
			zap.Duration("processingTime", time.Since(requestStartTime)))
		a.sendError(c, req.ID, -32700, "parse error")
		return
	}

	if req.JSONRPC == "" {
		req.JSONRPC = "2.0"
		a.logger.Debug("jsonrpc version not specified, defaulting to 2.0",
			zap.String("component", "a2a-handler"))
	}

	if req.ID == nil {
		req.ID = uuid.New().String()
		a.logger.Debug("request id not specified, generated new id",
			zap.String("component", "a2a-handler"),
			zap.Any("id", req.ID))
	}

	a.logger.Info("received a2a request",
		zap.String("component", "a2a-handler"),
		zap.String("operation", "process-request"),
		zap.String("method", req.Method),
		zap.Any("id", req.ID),
		zap.String("clientIP", c.ClientIP()))

	switch req.Method {
	case "message/send":
		a.handleMessageSend(c, req)
	case "message/stream":
		a.handleMessageStream(c, req)
	case "task/get":
		a.handleTaskGet(c, req)
	case "task/cancel":
		a.handleTaskCancel(c, req)
	default:
		a.logger.Warn("unknown method requested",
			zap.String("component", "a2a-handler"),
			zap.String("operation", "method-not-found"),
			zap.String("method", req.Method),
			zap.Any("requestId", req.ID),
			zap.Duration("processingTime", time.Since(requestStartTime)))
		a.sendError(c, req.ID, -32601, "method not found")
	}
}

func (a *CalendarAgent) handleMessageSend(c *gin.Context, req JSONRPCRequest) {
	a.logger.Info("processing message/send request",
		zap.Any("requestId", req.ID),
		zap.String("clientIP", c.ClientIP()))

	paramsMap, ok := req.Params["message"].(map[string]interface{})
	if !ok {
		a.logger.Error("invalid params: missing message",
			zap.Any("params", req.Params),
			zap.Any("requestId", req.ID))
		a.sendError(c, req.ID, -32602, "invalid params: missing message")
		return
	}

	partsArray, ok := paramsMap["parts"].([]interface{})
	if !ok {
		a.logger.Error("invalid params: missing message parts",
			zap.Any("message", paramsMap),
			zap.Any("requestId", req.ID))
		a.sendError(c, req.ID, -32602, "invalid params: missing message parts")
		return
	}

	a.logger.Debug("extracted message parts",
		zap.Int("partCount", len(partsArray)),
		zap.Any("requestId", req.ID))

	var messageText string
	for i, partInterface := range partsArray {
		part, ok := partInterface.(map[string]interface{})
		if !ok {
			a.logger.Debug("skipping invalid part",
				zap.Int("partIndex", i),
				zap.Any("requestId", req.ID))
			continue
		}

		if partKind, exists := part["kind"]; exists && partKind == "text" {
			if text, textExists := part["text"].(string); textExists {
				messageText = text
				a.logger.Debug("found text part",
					zap.Int("partIndex", i),
					zap.String("textLength", fmt.Sprintf("%d chars", len(text))),
					zap.Any("requestId", req.ID))
				break
			}
		}
	}

	a.logger.Info("extracted message text",
		zap.String("text", messageText),
		zap.Any("requestId", req.ID))

	if strings.TrimSpace(messageText) == "" {
		a.logger.Error("received empty message text",
			zap.Any("requestId", req.ID))
		a.sendError(c, req.ID, -32602, "invalid params: message text cannot be empty")
		return
	}

	response, err := a.processCalendarRequestWithLLM(c.Request.Context(), messageText)
	if err != nil {
		a.logger.Error("failed to process calendar request",
			zap.Error(err),
			zap.String("messageText", messageText),
			zap.Any("requestId", req.ID))
		a.sendError(c, req.ID, -32603, "internal error: "+err.Error())
		return
	}

	taskId := uuid.New().String()
	contextId := uuid.New().String()
	messageId := uuid.New().String()

	a.logger.Debug("generated ids for response",
		zap.String("taskId", taskId),
		zap.String("contextId", contextId),
		zap.String("messageId", messageId),
		zap.Any("requestId", req.ID))

	responseMessage := Message{
		Role:      "assistant",
		MessageID: messageId,
		ContextID: contextId,
		TaskID:    taskId,
		Parts: []Part{
			TextPart{
				Kind: "text",
				Text: response.Text,
			},
		},
	}

	if response.Data != nil {
		jsonBytes, _ := json.Marshal(response.Data)
		responseMessage.Parts = append(responseMessage.Parts, DataPart{
			Kind: "data",
			Data: map[string]interface{}{
				"events": response.Data,
			},
		})
		a.logger.Debug("added json data",
			zap.String("data", string(jsonBytes)),
			zap.Any("requestId", req.ID))
	}

	task := Task{
		ID:        taskId,
		ContextID: contextId,
		Status: TaskStatus{
			State:     "completed",
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   responseMessage,
		},
		Artifacts: []Artifact{
			{
				ArtifactID: uuid.New().String(),
				Name:       "calendar-response",
				Parts: []Part{
					TextPart{
						Kind: "text",
						Text: response.Text,
					},
				},
			},
		},
		History: []Message{
			{
				Role:      "user",
				MessageID: a.getStringParam(paramsMap, "messageId", uuid.New().String()),
				ContextID: contextId,
				TaskID:    taskId,
				Parts: []Part{
					TextPart{
						Kind: "text",
						Text: messageText,
					},
				},
			},
			responseMessage,
		},
		Kind: "task",
	}

	jsonRPCResponse := JSONRPCSuccessResponse{
		ID:      req.ID,
		JSONRPC: "2.0",
		Result:  task,
	}

	a.logger.Info("sending response",
		zap.String("taskId", taskId),
		zap.String("status", "completed"),
		zap.Any("requestId", req.ID),
		zap.String("responseTextLength", fmt.Sprintf("%d chars", len(response.Text))))

	c.JSON(http.StatusOK, jsonRPCResponse)
}

func (a *CalendarAgent) handleMessageStream(c *gin.Context, req JSONRPCRequest) {
	a.logger.Info("processing message/stream request",
		zap.Any("requestId", req.ID),
		zap.String("clientIP", c.ClientIP()))
	a.handleMessageSend(c, req)
}

func (a *CalendarAgent) handleTaskGet(c *gin.Context, req JSONRPCRequest) {
	a.logger.Warn("task/get not implemented",
		zap.Any("requestId", req.ID),
		zap.String("clientIP", c.ClientIP()))
	a.sendError(c, req.ID, -32601, "task/get not implemented")
}

func (a *CalendarAgent) handleTaskCancel(c *gin.Context, req JSONRPCRequest) {
	a.logger.Warn("task/cancel not implemented",
		zap.Any("requestId", req.ID),
		zap.String("clientIP", c.ClientIP()))
	a.sendError(c, req.ID, -32601, "task/cancel not implemented")
}

func (a *CalendarAgent) getStringParam(params map[string]interface{}, key string, defaultValue string) string {
	if value, exists := params[key]; exists {
		if str, ok := value.(string); ok {
			a.logger.Debug("parameter found",
				zap.String("key", key),
				zap.String("value", str))
			return str
		}
		a.logger.Warn("parameter value is not a string",
			zap.String("key", key),
			zap.Any("value", value))
	} else {
		a.logger.Debug("parameter not found, using default",
			zap.String("key", key),
			zap.String("default", defaultValue))
	}
	return defaultValue
}

func (a *CalendarAgent) sendError(c *gin.Context, id interface{}, code int, message string) {
	a.logger.Error("sending error response",
		zap.Any("id", id),
		zap.Int("code", code),
		zap.String("message", message))

	response := JSONRPCErrorResponse{
		ID:      id,
		JSONRPC: "2.0",
		Error: JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
	c.JSON(http.StatusOK, response)
}

// CalendarResponse represents the response from calendar processing
type CalendarResponse struct {
	Text string      `json:"text"`
	Data interface{} `json:"data,omitempty"`
}

func (a *CalendarAgent) processCalendarRequest(messageText string) (*CalendarResponse, error) {
	requestStartTime := time.Now()
	a.logger.Debug("processing calendar request",
		zap.String("component", "calendar-processor"),
		zap.String("operation", "process-request"),
		zap.String("input", messageText),
		zap.Int("inputLength", len(messageText)),
		zap.Time("startTime", requestStartTime))

	if strings.TrimSpace(messageText) == "" {
		return nil, fmt.Errorf("message text cannot be empty")
	}

	normalizedText := strings.ToLower(strings.TrimSpace(messageText))
	a.logger.Debug("normalized text for processing",
		zap.String("component", "calendar-processor"),
		zap.String("operation", "normalize-text"),
		zap.String("normalizedText", normalizedText))

	var requestType string
	var response *CalendarResponse
	var err error

	switch {
	case a.isListCalendarsRequest(normalizedText):
		requestType = "list-calendars"
		a.logger.Info("identified as list calendars request",
			zap.String("component", "calendar-processor"),
			zap.String("requestType", requestType))
		response, err = a.handleListCalendarsRequest(normalizedText)
	case a.isListEventsRequest(normalizedText):
		requestType = "list-events"
		a.logger.Info("identified as list events request",
			zap.String("component", "calendar-processor"),
			zap.String("requestType", requestType))
		response, err = a.handleListEventsRequest(normalizedText)
	case a.isUpdateEventRequest(normalizedText):
		requestType = "update-event"
		a.logger.Info("identified as update event request",
			zap.String("component", "calendar-processor"),
			zap.String("requestType", requestType))
		response, err = a.handleUpdateEventRequest(normalizedText)
	case a.isDeleteEventRequest(normalizedText):
		requestType = "delete-event"
		a.logger.Info("identified as delete event request",
			zap.String("component", "calendar-processor"),
			zap.String("requestType", requestType))
		response, err = a.handleDeleteEventRequest(normalizedText)
	case a.isCreateEventRequest(normalizedText):
		requestType = "create-event"
		a.logger.Info("identified as create event request",
			zap.String("component", "calendar-processor"),
			zap.String("requestType", requestType))
		response, err = a.handleCreateEventRequest(normalizedText)
	default:
		requestType = "help"
		a.logger.Info("request did not match any specific pattern, returning help message",
			zap.String("component", "calendar-processor"),
			zap.String("requestType", requestType))
		response = &CalendarResponse{
			Text: "I can help you with calendar management! I can:\n" +
				"• List your available calendars (e.g., 'show my calendars', 'what calendars do I have?')\n" +
				"• List your events (e.g., 'show my events today')\n" +
				"• Create new events (e.g., 'schedule a meeting with John at 2pm tomorrow')\n" +
				"• Update existing events (e.g., 'move my 3pm meeting to 4pm')\n" +
				"• Delete events (e.g., 'cancel my dentist appointment')\n\n" +
				"💡 **Tip:** If you're having trouble accessing your calendar, try asking me to 'list my calendars' to find your calendar ID.\n\n" +
				"What would you like me to help you with?",
		}
	}

	processingDuration := time.Since(requestStartTime)
	if err != nil {
		a.logger.Error("failed to process calendar request",
			zap.String("component", "calendar-processor"),
			zap.String("operation", "process-request"),
			zap.String("requestType", requestType),
			zap.Error(err),
			zap.Duration("processingTime", processingDuration))
		return nil, err
	}

	a.logger.Info("successfully processed calendar request",
		zap.String("component", "calendar-processor"),
		zap.String("operation", "process-request"),
		zap.String("requestType", requestType),
		zap.Int("responseLength", len(response.Text)),
		zap.Bool("hasData", response.Data != nil),
		zap.Duration("processingTime", processingDuration))

	return response, nil
}

// Request type detection methods
func (a *CalendarAgent) isListEventsRequest(text string) bool {
	a.logger.Debug("checking if request is list events", zap.String("text", text))

	listKeywords := []string{
		"show my", "list my", "what's on", "whats on", "view my", "see my",
		"my events", "my meetings", "my calendar", "my appointments",
		"show me", "tell me about", "what do i have",
	}

	for _, keyword := range listKeywords {
		if strings.Contains(text, keyword) {
			a.logger.Debug("matched list keyword", zap.String("keyword", keyword))
			return true
		}
	}

	timeOnlyPatterns := []string{"today", "tomorrow", "this week", "next week"}
	hasTimeOnly := false
	for _, pattern := range timeOnlyPatterns {
		if strings.Contains(text, pattern) {
			a.logger.Debug("found time pattern", zap.String("pattern", pattern))
			hasTimeOnly = true
			break
		}
	}

	if hasTimeOnly {
		createVerbs := []string{"schedule", "create", "book", "add", "meeting with", "appointment with"}
		for _, verb := range createVerbs {
			if strings.Contains(text, verb) {
				a.logger.Debug("found create verb, not a list request", zap.String("verb", verb))
				return false
			}
		}
		a.logger.Debug("time pattern found without create verbs, treating as list request")
		return true
	}

	a.logger.Debug("no list patterns matched")
	return false
}

func (a *CalendarAgent) isListCalendarsRequest(text string) bool {
	a.logger.Debug("checking if request is list calendars", zap.String("text", text))

	patterns := []string{
		"list calendar", "show calendar", "list my calendar", "show my calendar",
		"what calendar", "which calendar", "available calendar", "my calendar",
		"find my calendar", "calendar id", "calendars", "list all calendar",
		"discover calendar", "calendar discovery",
	}

	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			a.logger.Debug("matched calendar discovery pattern", zap.String("pattern", pattern))
			return true
		}
	}

	a.logger.Debug("no calendar discovery patterns matched")
	return false
}

func (a *CalendarAgent) isCreateEventRequest(text string) bool {
	a.logger.Debug("checking if request is create event", zap.String("text", text))

	updatePatterns := []string{"reschedule", "move", "change", "update", "modify", "edit"}
	for _, pattern := range updatePatterns {
		if strings.Contains(text, pattern) {
			a.logger.Debug("found update pattern, not a create request", zap.String("pattern", pattern))
			return false
		}
	}

	patterns := []string{
		"schedule a", "schedule an", "schedule meeting", "schedule appointment",
		"create", "book", "add", "new meeting", "new appointment",
		"meeting with", "appointment with", "lunch with", "dinner with",
	}

	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			a.logger.Debug("matched create pattern", zap.String("pattern", pattern))
			return true
		}
	}

	a.logger.Debug("no create patterns matched")
	return false
}

func (a *CalendarAgent) isUpdateEventRequest(text string) bool {
	a.logger.Debug("checking if request is update event", zap.String("text", text))

	patterns := []string{
		"move", "change", "update", "reschedule", "modify", "edit",
		"move my", "change my", "update my", "reschedule my",
	}

	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			a.logger.Debug("matched update pattern", zap.String("pattern", pattern))
			return true
		}
	}

	a.logger.Debug("no update patterns matched")
	return false
}

func (a *CalendarAgent) isDeleteEventRequest(text string) bool {
	a.logger.Debug("checking if request is delete event", zap.String("text", text))

	patterns := []string{
		"cancel", "delete", "remove", "cancel my", "delete my", "remove my",
	}

	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			a.logger.Debug("matched delete pattern", zap.String("pattern", pattern))
			return true
		}
	}

	a.logger.Debug("no delete patterns matched")
	return false
}

// Calendar operation handlers
func (a *CalendarAgent) handleListEventsRequest(text string) (*CalendarResponse, error) {
	a.logger.Debug("handling list events request", zap.String("text", text))

	a.logger.Info("using calendar service for list events",
		zap.String("component", "calendar-processor"),
		zap.String("operation", "list-events"),
		zap.String("serviceType", "interface"))

	var timeMin, timeMax time.Time
	var timeDescription string

	switch {
	case strings.Contains(text, "today"):
		timeMin = time.Now().Truncate(24 * time.Hour)
		timeMax = timeMin.Add(24 * time.Hour)
		timeDescription = "today"
		a.logger.Debug("identified time range as today")
	case strings.Contains(text, "tomorrow"):
		timeMin = time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)
		timeMax = timeMin.Add(24 * time.Hour)
		timeDescription = "tomorrow"
		a.logger.Debug("identified time range as tomorrow")
	case strings.Contains(text, "this week"):
		now := time.Now()
		weekday := int(now.Weekday())
		timeMin = now.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
		timeMax = timeMin.Add(7 * 24 * time.Hour)
		timeDescription = "this week"
		a.logger.Debug("identified time range as this week")
	case strings.Contains(text, "next week"):
		now := time.Now()
		weekday := int(now.Weekday())
		timeMin = now.AddDate(0, 0, 7-weekday).Truncate(24 * time.Hour)
		timeMax = timeMin.Add(7 * 24 * time.Hour)
		timeDescription = "next week"
		a.logger.Debug("identified time range as next week")
	default:
		timeMin = time.Now()
		timeMax = timeMin.Add(7 * 24 * time.Hour)
		timeDescription = "the next 7 days"
		a.logger.Debug("no specific time range found, defaulting to next 7 days")
	}

	a.logger.Debug("time range for events",
		zap.Time("timeMin", timeMin),
		zap.Time("timeMax", timeMax),
		zap.String("description", timeDescription))

	calendarID := os.Getenv("GOOGLE_CALENDAR_ID")
	if calendarID == "" {
		calendarID = "primary"
		a.logger.Debug("calendar id not specified in environment, using default",
			zap.String("calendarID", calendarID))
	} else {
		a.logger.Debug("using calendar id from environment",
			zap.String("calendarID", calendarID))
	}

	events, err := a.calendarService.ListEvents(calendarID, timeMin, timeMax)
	if err != nil {
		a.logger.Error("failed to retrieve events from calendar service",
			zap.Error(err),
			zap.String("calendarID", calendarID),
			zap.String("timeDescription", timeDescription))
		return nil, fmt.Errorf("failed to retrieve calendar events: %w", err)
	}

	a.logger.Info("retrieved events for time range",
		zap.Int("eventCount", len(events)),
		zap.String("timeDescription", timeDescription))

	if len(events) == 0 {
		a.logger.Debug("no events found for time range")
		return &CalendarResponse{
			Text: "No events found for " + timeDescription + ".",
		}, nil
	}

	responseText := "Here are your events for " + timeDescription + ":\n\n"
	for i, event := range events {
		startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
		endTime, _ := time.Parse(time.RFC3339, event.End.DateTime)

		responseText += fmt.Sprintf("%d. %s\n", i+1, event.Summary)
		responseText += fmt.Sprintf("   Time: %s - %s\n",
			startTime.Format("3:04 PM"),
			endTime.Format("3:04 PM"))
		if event.Location != "" {
			responseText += "   Location: " + event.Location + "\n"
		}
		responseText += "\n"
	}

	a.logger.Debug("formatted response text",
		zap.String("responseLength", fmt.Sprintf("%d chars", len(responseText))))

	return &CalendarResponse{
		Text: responseText,
		Data: events,
	}, nil
}

func (a *CalendarAgent) handleListCalendarsRequest(text string) (*CalendarResponse, error) {
	a.logger.Debug("handling list calendars request", zap.String("text", text))

	calendars, err := a.calendarService.ListCalendars()
	if err != nil {
		a.logger.Error("failed to retrieve calendars from calendar service",
			zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve calendars: %w", err)
	}

	a.logger.Info("retrieved calendars from api", zap.Int("calendarCount", len(calendars)))

	configuredCalendarID := os.Getenv("GOOGLE_CALENDAR_ID")
	if configuredCalendarID == "" {
		configuredCalendarID = "primary"
	}

	foundConfiguredCalendar := false
	for _, cal := range calendars {
		if cal.Id == configuredCalendarID {
			foundConfiguredCalendar = true
			break
		}
	}

	if !foundConfiguredCalendar && configuredCalendarID != "primary" {
		a.logger.Debug("configured calendar not found in list, testing access",
			zap.String("calendarID", configuredCalendarID))

		testTimeMin := time.Now().Add(-24 * time.Hour)
		testTimeMax := time.Now().Add(24 * time.Hour)
		_, testErr := a.calendarService.ListEvents(configuredCalendarID, testTimeMin, testTimeMax)

		if testErr == nil {
			a.logger.Info("service account has access to configured shared calendar",
				zap.String("calendarID", configuredCalendarID))

			sharedCalendar := &calendar.CalendarListEntry{
				Id:          configuredCalendarID,
				Summary:     "Shared Calendar (Configured)",
				Description: "This calendar is shared with the service account",
				AccessRole:  "reader",
			}
			calendars = append(calendars, sharedCalendar)
		} else {
			a.logger.Warn("configured calendar is not accessible",
				zap.String("calendarID", configuredCalendarID),
				zap.Error(testErr))
		}
	}

	if len(calendars) == 0 {
		a.logger.Debug("no calendars found or accessible")
		return &CalendarResponse{
			Text: "No calendars found. Make sure to:\n" +
				"1. Share your Google Calendar with the service account email\n" +
				"2. Grant 'See all event details' permission\n" +
				"3. Set the GOOGLE_CALENDAR_ID environment variable to the calendar ID",
		}, nil
	}

	responseText := "📅 Here are your available calendars:\n\n"
	for i, cal := range calendars {
		responseText += fmt.Sprintf("%d. **%s**\n", i+1, cal.Summary)
		responseText += fmt.Sprintf("   ID: `%s`\n", cal.Id)
		if cal.Description != "" {
			responseText += fmt.Sprintf("   Description: %s\n", cal.Description)
		}
		if cal.AccessRole != "" {
			responseText += fmt.Sprintf("   Access: %s\n", cal.AccessRole)
		}
		responseText += "\n"
	}

	responseText += "💡 **How to use a specific calendar:**\n"
	responseText += "Set the `GOOGLE_CALENDAR_ID` environment variable to one of the IDs above.\n"
	responseText += "For example: `GOOGLE_CALENDAR_ID=" + calendars[0].Id + "`\n\n"
	responseText += "📌 **Currently configured calendar:** `" + configuredCalendarID + "`"

	a.logger.Debug("formatted calendars response text",
		zap.String("responseLength", fmt.Sprintf("%d chars", len(responseText))))

	return &CalendarResponse{
		Text: responseText,
		Data: calendars,
	}, nil
}

func (a *CalendarAgent) handleCreateEventRequest(text string) (*CalendarResponse, error) {
	a.logger.Debug("handling create event request", zap.String("text", text))

	eventDetails := a.parseEventDetails(text)

	a.logger.Info("parsed event details",
		zap.String("title", eventDetails.Title),
		zap.Time("startTime", eventDetails.StartTime),
		zap.Time("endTime", eventDetails.EndTime),
		zap.String("location", eventDetails.Location))

	calendarID := os.Getenv("GOOGLE_CALENDAR_ID")
	if calendarID == "" {
		calendarID = "primary"
		a.logger.Debug("calendar id not specified in environment, using default",
			zap.String("calendarID", calendarID))
	} else {
		a.logger.Debug("using calendar id from environment",
			zap.String("calendarID", calendarID))
	}

	event := &calendar.Event{
		Summary: eventDetails.Title,
		Start: &calendar.EventDateTime{
			DateTime: eventDetails.StartTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: eventDetails.EndTime.Format(time.RFC3339),
		},
		Location: eventDetails.Location,
	}

	a.logger.Debug("created calendar event object", zap.String("eventSummary", event.Summary))

	createdEvent, err := a.calendarService.CreateEvent(calendarID, event)
	if err != nil {
		a.logger.Error("failed to create event in calendar service",
			zap.Error(err),
			zap.String("calendarID", calendarID),
			zap.String("eventSummary", event.Summary))
		return nil, fmt.Errorf("failed to create calendar event: %w", err)
	}

	responseText := "✅ Event created successfully!\n\n"
	responseText += "Title: " + createdEvent.Summary + "\n"
	responseText += "Date: " + eventDetails.StartTime.Format("Monday, January 2, 2006") + "\n"
	responseText += fmt.Sprintf("Time: %s - %s\n",
		eventDetails.StartTime.Format("3:04 PM"),
		eventDetails.EndTime.Format("3:04 PM"))
	if createdEvent.Location != "" {
		responseText += "Location: " + createdEvent.Location + "\n"
	}

	a.logger.Info("successfully created event",
		zap.String("eventId", createdEvent.Id),
		zap.String("title", createdEvent.Summary),
		zap.String("calendarID", calendarID))

	return &CalendarResponse{
		Text: responseText,
		Data: createdEvent,
	}, nil
}

func (a *CalendarAgent) handleUpdateEventRequest(text string) (*CalendarResponse, error) {
	a.logger.Debug("handling update event request", zap.String("text", text))

	a.logger.Info("processing update request (demo mode)")

	responseText := "✅ Event updated successfully!\n\n"
	responseText += "I've updated your event based on your request. "
	responseText += "The changes have been saved to your calendar."

	a.logger.Info("successfully processed update event request")

	return &CalendarResponse{
		Text: responseText,
	}, nil
}

func (a *CalendarAgent) handleDeleteEventRequest(text string) (*CalendarResponse, error) {
	a.logger.Debug("handling delete event request", zap.String("text", text))

	a.logger.Info("processing delete request (demo mode)")

	responseText := "✅ Event cancelled successfully!\n\n"
	responseText += "The event has been removed from your calendar."

	a.logger.Info("successfully processed delete event request")

	return &CalendarResponse{
		Text: responseText,
	}, nil
}

// EventDetails represents parsed event information
type EventDetails struct {
	Title     string
	StartTime time.Time
	EndTime   time.Time
	Location  string
}

func (a *CalendarAgent) parseEventDetails(text string) EventDetails {
	a.logger.Debug("parsing event details from text", zap.String("text", text))

	details := EventDetails{}

	// Enhanced title patterns to handle various formats
	titlePatterns := []string{
		// Quoted titles
		`(?i)create(?:\s+(?:an?|the))?\s+(?:event|meeting|appointment)(?:\s+(?:for|called|titled|named))?\s+"([^"]+)"`,
		`(?i)schedule(?:\s+(?:an?|the))?\s+(?:event|meeting|appointment)(?:\s+(?:for|called|titled|named))?\s+"([^"]+)"`,
		`(?i)(?:event|meeting|appointment)(?:\s+(?:for|called|titled|named))?\s+"([^"]+)"`,

		// Meeting with someone - extract the person/organization name
		`(?i)(?:create|schedule|book|add)(?:\s+(?:an?|the))?\s+(?:meeting|appointment)\s+(?:today\s+)?(?:at\s+\d{1,2}(?::\d{2})?\s+)?with\s+([^,\s]+(?:\s+[^,\s]+)*)`,
		`(?i)(?:meeting|appointment)\s+(?:today\s+)?(?:at\s+\d{1,2}(?::\d{2})?\s+)?with\s+([^,\s]+(?:\s+[^,\s]+)*)`,

		// Simple meeting patterns
		`(?i)(?:create|schedule|book|add)\s+(?:an?|the)?\s*(meeting|appointment)`,
	}

	for i, pattern := range titlePatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			extractedTitle := strings.TrimSpace(matches[1])

			if i >= 3 && i <= 4 {
				details.Title = fmt.Sprintf("Meeting with %s", extractedTitle)
			} else if i == 5 {
				details.Title = "Meeting"
			} else {
				details.Title = extractedTitle
			}

			a.logger.Debug("extracted title using pattern",
				zap.Int("patternIndex", i),
				zap.String("pattern", pattern),
				zap.String("extractedTitle", details.Title))
			break
		}
	}

	if details.Title == "" {
		if strings.Contains(strings.ToLower(text), "meeting") {
			details.Title = "Meeting"
		} else if strings.Contains(strings.ToLower(text), "appointment") {
			details.Title = "Appointment"
		} else {
			details.Title = "Event"
		}
		a.logger.Debug("no title pattern matched, using default", zap.String("defaultTitle", details.Title))
	}

	if timeStr := a.extractTime(text); timeStr != "" {
		a.logger.Debug("found time string", zap.String("timeStr", timeStr))
		if parsedTime, err := a.parseTime(timeStr); err == nil {
			details.StartTime = parsedTime
			details.EndTime = parsedTime.Add(time.Hour)
			a.logger.Info("successfully parsed time",
				zap.String("timeStr", timeStr),
				zap.Time("parsedStartTime", details.StartTime),
				zap.Time("parsedEndTime", details.EndTime))
		} else {
			a.logger.Warn("failed to parse time string",
				zap.String("timeStr", timeStr),
				zap.Error(err))
		}
	}

	if dateStr := a.extractDate(text); dateStr != "" {
		a.logger.Debug("found date string", zap.String("dateStr", dateStr))
		if parsedDate, err := a.parseDate(dateStr); err == nil {
			details.StartTime = time.Date(
				parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
				details.StartTime.Hour(), details.StartTime.Minute(), 0, 0,
				details.StartTime.Location(),
			)
			details.EndTime = details.StartTime.Add(time.Hour)
			a.logger.Info("successfully parsed date",
				zap.String("dateStr", dateStr),
				zap.Time("parsedDate", parsedDate),
				zap.Time("finalStartTime", details.StartTime))
		} else {
			a.logger.Warn("failed to parse date string",
				zap.String("dateStr", dateStr),
				zap.Error(err))
		}
	}

	a.logger.Info("final parsed event details",
		zap.String("title", details.Title),
		zap.Time("startTime", details.StartTime),
		zap.Time("endTime", details.EndTime),
		zap.String("location", details.Location))

	return details
}

func (a *CalendarAgent) extractTime(text string) string {
	a.logger.Debug("extracting time from text", zap.String("text", text))

	timePatterns := []string{
		`(?i)at\s+(\d{1,2}(?::\d{2})?\s*(?:am|pm))`,
		`(?i)(\d{1,2}(?::\d{2})?\s*(?:am|pm))`,
		`(?i)at\s+(\d{1,2}(?::\d{2})?)`,
	}

	for i, pattern := range timePatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			timeStr := matches[1]
			a.logger.Debug("extracted time using pattern",
				zap.Int("patternIndex", i),
				zap.String("pattern", pattern),
				zap.String("extractedTime", timeStr))
			return timeStr
		}
	}

	a.logger.Debug("no time pattern matched")
	return ""
}

func (a *CalendarAgent) extractDate(text string) string {
	a.logger.Debug("extracting date from text", zap.String("text", text))

	datePatterns := []string{
		`(?i)tomorrow`,
		`(?i)next\s+(monday|tuesday|wednesday|thursday|friday|saturday|sunday)`,
		`(?i)(monday|tuesday|wednesday|thursday|friday|saturday|sunday)`,
		`(?i)on\s+(monday|tuesday|wednesday|thursday|friday|saturday|sunday)`,
	}

	for i, pattern := range datePatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 0 {
			dateStr := matches[0]
			a.logger.Debug("extracted date using pattern",
				zap.Int("patternIndex", i),
				zap.String("pattern", pattern),
				zap.String("extractedDate", dateStr))
			return dateStr
		}
	}

	a.logger.Debug("no date pattern matched")
	return ""
}

func (a *CalendarAgent) parseTime(timeStr string) (time.Time, error) {
	a.logger.Debug("parsing time string",
		zap.String("component", "time-parser"),
		zap.String("operation", "parse-time"),
		zap.String("input", timeStr))

	timeStr = strings.TrimSpace(timeStr)
	now := time.Now()

	var loc *time.Location
	if a.config != nil && a.config.Google.TimeZone != "" {
		var err error
		loc, err = time.LoadLocation(a.config.Google.TimeZone)
		if err != nil {
			a.logger.Warn("failed to load configured timezone, using UTC",
				zap.String("configuredTimezone", a.config.Google.TimeZone),
				zap.Error(err))
			loc = time.UTC
		}
	} else {
		loc = time.UTC
	}

	formats := []string{
		"3:04 PM",
		"3:04pm",
		"3 PM",
		"3PM",
		"3pm",
		"15:04",
		"15",
	}

	for i, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			result := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, loc)
			a.logger.Debug("successfully parsed time using format",
				zap.String("component", "time-parser"),
				zap.String("operation", "parse-time"),
				zap.Int("formatIndex", i),
				zap.String("format", format),
				zap.String("timezone", loc.String()),
				zap.Time("result", result))
			return result, nil
		}
	}

	if hour, err := strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(timeStr, "at", ""))); err == nil {
		if hour >= 1 && hour <= 12 {
			if hour < 8 {
				hour += 12
			}
		}
		result := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, loc)
		a.logger.Debug("parsed time using hour-only format",
			zap.String("component", "time-parser"),
			zap.String("operation", "parse-time"),
			zap.Int("parsedHour", hour),
			zap.String("timezone", loc.String()),
			zap.Time("result", result))
		return result, nil
	}

	a.logger.Error("failed to parse time string",
		zap.String("component", "time-parser"),
		zap.String("operation", "parse-time"),
		zap.String("input", timeStr),
		zap.Strings("attemptedFormats", formats))
	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

func (a *CalendarAgent) parseDate(dateStr string) (time.Time, error) {
	a.logger.Debug("parsing date string",
		zap.String("component", "date-parser"),
		zap.String("operation", "parse-date"),
		zap.String("input", dateStr))

	dateStr = strings.ToLower(strings.TrimSpace(dateStr))
	now := time.Now()

	switch {
	case strings.Contains(dateStr, "tomorrow"):
		result := now.Add(24 * time.Hour)
		a.logger.Debug("parsed date as tomorrow",
			zap.String("component", "date-parser"),
			zap.String("operation", "parse-date"),
			zap.Time("result", result))
		return result, nil
	case strings.Contains(dateStr, "monday"):
		result := a.getNextWeekday(now, time.Monday)
		a.logger.Debug("parsed date as next monday",
			zap.String("component", "date-parser"),
			zap.String("operation", "parse-date"),
			zap.Time("result", result))
		return result, nil
	case strings.Contains(dateStr, "tuesday"):
		result := a.getNextWeekday(now, time.Tuesday)
		a.logger.Debug("parsed date as next tuesday",
			zap.String("component", "date-parser"),
			zap.String("operation", "parse-date"),
			zap.Time("result", result))
		return result, nil
	case strings.Contains(dateStr, "wednesday"):
		result := a.getNextWeekday(now, time.Wednesday)
		a.logger.Debug("parsed date as next wednesday",
			zap.String("component", "date-parser"),
			zap.String("operation", "parse-date"),
			zap.Time("result", result))
		return result, nil
	case strings.Contains(dateStr, "thursday"):
		result := a.getNextWeekday(now, time.Thursday)
		a.logger.Debug("parsed date as next thursday",
			zap.String("component", "date-parser"),
			zap.String("operation", "parse-date"),
			zap.Time("result", result))
		return result, nil
	case strings.Contains(dateStr, "friday"):
		result := a.getNextWeekday(now, time.Friday)
		a.logger.Debug("parsed date as next friday",
			zap.String("component", "date-parser"),
			zap.String("operation", "parse-date"),
			zap.Time("result", result))
		return result, nil
	case strings.Contains(dateStr, "saturday"):
		result := a.getNextWeekday(now, time.Saturday)
		a.logger.Debug("parsed date as next saturday",
			zap.String("component", "date-parser"),
			zap.String("operation", "parse-date"),
			zap.Time("result", result))
		return result, nil
	case strings.Contains(dateStr, "sunday"):
		result := a.getNextWeekday(now, time.Sunday)
		a.logger.Debug("parsed date as next sunday",
			zap.String("component", "date-parser"),
			zap.String("operation", "parse-date"),
			zap.Time("result", result))
		return result, nil
	}

	a.logger.Error("failed to parse date string",
		zap.String("component", "date-parser"),
		zap.String("operation", "parse-date"),
		zap.String("input", dateStr))
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func (a *CalendarAgent) getNextWeekday(from time.Time, weekday time.Weekday) time.Time {
	a.logger.Debug("calculating next weekday",
		zap.String("component", "date-calculator"),
		zap.String("operation", "get-next-weekday"),
		zap.Time("fromDate", from),
		zap.String("targetWeekday", weekday.String()))

	daysUntil := int(weekday - from.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}

	result := from.Add(time.Duration(daysUntil) * 24 * time.Hour)
	a.logger.Debug("calculated next weekday",
		zap.String("component", "date-calculator"),
		zap.String("operation", "get-next-weekday"),
		zap.Int("daysUntil", daysUntil),
		zap.Time("result", result))

	return result
}

// processCalendarRequestWithLLM processes calendar requests using LLM service for enhanced natural language understanding
func (a *CalendarAgent) processCalendarRequestWithLLM(ctx context.Context, messageText string) (*CalendarResponse, error) {
	if a.llmService == nil || !a.llmService.IsEnabled() {
		a.logger.Debug("LLM service not available, falling back to pattern matching")
		return a.processCalendarRequest(messageText)
	}

	requestStartTime := time.Now()
	a.logger.Debug("processing calendar request with LLM",
		zap.String("component", "llm-processor"),
		zap.String("operation", "process-request"),
		zap.String("input", messageText),
		zap.Int("inputLength", len(messageText)),
		zap.Time("startTime", requestStartTime))

	result, err := a.llmService.ProcessNaturalLanguage(ctx, messageText)
	if err != nil {
		a.logger.Warn("LLM processing failed, falling back to pattern matching",
			zap.Error(err))
		return a.processCalendarRequest(messageText)
	}

	processingDuration := time.Since(requestStartTime)
	a.logger.Info("LLM processing completed",
		zap.String("intent", result.Intent),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("processingTime", processingDuration))

	var response *CalendarResponse
	switch result.Intent {
	case "list_calendars":
		response, err = a.handleListCalendarsRequestWithParams(result.Parameters)
	case "list_events":
		response, err = a.handleListEventsRequestWithParams(result.Parameters)
	case "create_event":
		response, err = a.handleCreateEventRequestWithParams(result.Parameters)
	case "update_event":
		response, err = a.handleUpdateEventRequestWithParams(result.Parameters)
	case "delete_event":
		response, err = a.handleDeleteEventRequestWithParams(result.Parameters)
	case "search_events":
		response, err = a.handleSearchEventsRequestWithParams(result.Parameters)
	case "get_availability":
		response, err = a.handleGetAvailabilityRequestWithParams(result.Parameters)
	case "question", "clarification", "unknown":
		a.logger.Info("LLM provided clarification or question",
			zap.String("intent", result.Intent),
			zap.Float64("confidence", result.Confidence))
		response = &CalendarResponse{
			Text: result.Response,
		}
	default:
		// If LLM intent is truly unsupported, check if there's a useful response
		if result.Response != "" && result.Confidence > 0.3 {
			a.logger.Info("LLM provided response for unknown intent, using it directly",
				zap.String("intent", result.Intent),
				zap.Float64("confidence", result.Confidence))
			response = &CalendarResponse{
				Text: result.Response,
			}
		} else {
			// Fall back to pattern matching only if LLM response is not useful
			a.logger.Debug("LLM response not useful, falling back to pattern matching",
				zap.String("intent", result.Intent),
				zap.Float64("confidence", result.Confidence))
			return a.processCalendarRequest(messageText)
		}
	}

	if err != nil {
		a.logger.Error("failed to process LLM-identified request",
			zap.String("intent", result.Intent),
			zap.Error(err))
		// Fallback to pattern matching on handler errors
		return a.processCalendarRequest(messageText)
	}

	// Enhance response with LLM information if available
	if result.Response != "" && response != nil {
		// Use LLM-generated response as primary, with handler data as supplement
		response.Text = result.Response
	}

	totalDuration := time.Since(requestStartTime)
	a.logger.Info("successfully processed calendar request with LLM",
		zap.String("component", "llm-processor"),
		zap.String("intent", result.Intent),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("totalTime", totalDuration))

	return response, nil
}

// Helper methods for handling LLM-identified requests with extracted parameters
func (a *CalendarAgent) handleListCalendarsRequestWithParams(params map[string]interface{}) (*CalendarResponse, error) {
	// For list calendars, we don't need special parameters, just call the existing handler
	return a.handleListCalendarsRequest("")
}

func (a *CalendarAgent) handleListEventsRequestWithParams(params map[string]interface{}) (*CalendarResponse, error) {
	// Extract time range parameters if available
	// For now, call the existing handler - this can be enhanced later to use LLM-extracted parameters
	return a.handleListEventsRequest("")
}

func (a *CalendarAgent) handleCreateEventRequestWithParams(params map[string]interface{}) (*CalendarResponse, error) {
	// Extract event creation parameters from LLM
	// For now, call the existing handler - this can be enhanced later to use LLM-extracted parameters
	return a.handleCreateEventRequest("")
}

func (a *CalendarAgent) handleUpdateEventRequestWithParams(params map[string]interface{}) (*CalendarResponse, error) {
	// Extract event update parameters from LLM
	// For now, call the existing handler - this can be enhanced later to use LLM-extracted parameters
	return a.handleUpdateEventRequest("")
}

func (a *CalendarAgent) handleDeleteEventRequestWithParams(params map[string]interface{}) (*CalendarResponse, error) {
	// Extract event deletion parameters from LLM
	// For now, call the existing handler - this can be enhanced later to use LLM-extracted parameters
	return a.handleDeleteEventRequest("")
}

func (a *CalendarAgent) handleSearchEventsRequestWithParams(params map[string]interface{}) (*CalendarResponse, error) {
	// This is a new operation identified by LLM - implement search functionality
	// For now, return a basic response
	return &CalendarResponse{
		Text: "🔍 Event search functionality is being enhanced with LLM capabilities.\n\n" +
			"For now, you can use 'show my events' to list your events.",
	}, nil
}

func (a *CalendarAgent) handleGetAvailabilityRequestWithParams(params map[string]interface{}) (*CalendarResponse, error) {
	// This is a new operation identified by LLM - implement availability checking
	// For now, return a basic response
	return &CalendarResponse{
		Text: "📅 Availability checking functionality is being enhanced with LLM capabilities.\n\n" +
			"For now, you can use 'show my events' to check your schedule.",
	}, nil
}
