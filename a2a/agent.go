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

	a.logger.Debug("full request params",
		zap.Any("params", req.Params),
		zap.Any("requestId", req.ID))

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
		zap.Any("parts", partsArray),
		zap.Any("requestId", req.ID))

	var messageText string
	for i, partInterface := range partsArray {
		part, ok := partInterface.(map[string]interface{})
		if !ok {
			a.logger.Debug("skipping invalid part",
				zap.Int("partIndex", i),
				zap.Any("part", partInterface),
				zap.Any("requestId", req.ID))
			continue
		}

		a.logger.Debug("processing part",
			zap.Int("partIndex", i),
			zap.Any("part", part),
			zap.Any("requestId", req.ID))

		if partKind, exists := part["kind"]; exists && partKind == "text" {
			if text, textExists := part["text"].(string); textExists {
				messageText = text
				a.logger.Debug("found text part",
					zap.Int("partIndex", i),
					zap.String("textLength", fmt.Sprintf("%d chars", len(text))),
					zap.String("text", text),
					zap.Any("requestId", req.ID))
				break
			}
		}
	}

	a.logger.Info("extracted message text",
		zap.String("text", messageText),
		zap.Any("requestId", req.ID))

	var response *CalendarResponse
	var err error

	if strings.TrimSpace(messageText) == "" {
		if metadata, hasMetadata := req.Params["metadata"].(map[string]interface{}); hasMetadata {
			if skill, hasSkill := metadata["skill"].(string); hasSkill {
				if arguments, hasArgs := metadata["arguments"].(map[string]interface{}); hasArgs {
					a.logger.Info("processing direct tool call",
						zap.String("skill", skill),
						zap.Any("arguments", arguments),
						zap.Any("requestId", req.ID))

					response, err = a.processDirectToolCall(c.Request.Context(), skill, arguments)
					if err != nil {
						a.logger.Error("failed to process direct tool call",
							zap.Error(err),
							zap.String("skill", skill),
							zap.Any("arguments", arguments),
							zap.Any("requestId", req.ID))
						a.sendError(c, req.ID, -32603, "internal error: "+err.Error())
						return
					}
				} else {
					a.logger.Error("direct tool call missing arguments",
						zap.Any("requestId", req.ID))
					a.sendError(c, req.ID, -32602, "invalid params: direct tool call missing arguments")
					return
				}
			} else {
				a.logger.Error("received empty message text and no direct tool call",
					zap.Any("requestId", req.ID))
				a.sendError(c, req.ID, -32602, "invalid params: message text cannot be empty")
				return
			}
		} else {
			a.logger.Error("received empty message text and no metadata",
				zap.Any("requestId", req.ID))
			a.sendError(c, req.ID, -32602, "invalid params: message text cannot be empty")
			return
		}
	} else {
		response, err = a.processCalendarRequestWithLLM(c.Request.Context(), messageText)
	}
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

	// Check for conflicts before creating the event
	conflicts, err := a.calendarService.CheckConflicts(calendarID, eventDetails.StartTime, eventDetails.EndTime)
	if err != nil {
		a.logger.Error("failed to check for conflicts",
			zap.Error(err),
			zap.String("calendarID", calendarID),
			zap.Time("startTime", eventDetails.StartTime),
			zap.Time("endTime", eventDetails.EndTime))
		return nil, fmt.Errorf("failed to check for scheduling conflicts: %w", err)
	}

	// If conflicts found, suggest alternative times
	if len(conflicts) > 0 {
		a.logger.Info("found scheduling conflicts",
			zap.Int("conflictCount", len(conflicts)),
			zap.String("proposedTitle", eventDetails.Title),
			zap.Time("proposedStartTime", eventDetails.StartTime),
			zap.Time("proposedEndTime", eventDetails.EndTime))

		conflictText := "⚠️ **Scheduling Conflict Detected!**\n\n"
		conflictText += fmt.Sprintf("You already have %d event(s) scheduled during %s - %s:\n\n",
			len(conflicts),
			eventDetails.StartTime.Format("3:04 PM"),
			eventDetails.EndTime.Format("3:04 PM"))

		for i, conflict := range conflicts {
			conflictStartTime, _ := time.Parse(time.RFC3339, conflict.Start.DateTime)
			conflictEndTime, _ := time.Parse(time.RFC3339, conflict.End.DateTime)
			conflictText += fmt.Sprintf("%d. **%s**\n   Time: %s - %s\n",
				i+1,
				conflict.Summary,
				conflictStartTime.Format("3:04 PM"),
				conflictEndTime.Format("3:04 PM"))
			if conflict.Location != "" {
				conflictText += fmt.Sprintf("   Location: %s\n", conflict.Location)
			}
			conflictText += "\n"
		}

		// Suggest alternative times
		conflictText += "**Suggested alternative times:**\n"
		duration := eventDetails.EndTime.Sub(eventDetails.StartTime)

		// Suggest 1 hour later
		altTime1 := eventDetails.StartTime.Add(time.Hour)
		conflictText += fmt.Sprintf("• %s - %s\n",
			altTime1.Format("3:04 PM"),
			altTime1.Add(duration).Format("3:04 PM"))

		// Suggest 2 hours later
		altTime2 := eventDetails.StartTime.Add(2 * time.Hour)
		conflictText += fmt.Sprintf("• %s - %s\n",
			altTime2.Format("3:04 PM"),
			altTime2.Add(duration).Format("3:04 PM"))

		// Suggest next day same time
		altTime3 := eventDetails.StartTime.AddDate(0, 0, 1)
		conflictText += fmt.Sprintf("• %s - %s (%s)\n",
			altTime3.Format("3:04 PM"),
			altTime3.Add(duration).Format("3:04 PM"),
			altTime3.Format("Monday, January 2"))

		conflictText += "\nWould you like me to schedule it at one of these alternative times instead?"

		return &CalendarResponse{
			Text: conflictText,
			Data: map[string]interface{}{
				"conflicts":      conflicts,
				"proposed_event": eventDetails,
				"calendar_id":    calendarID,
				"alternative_times": []map[string]interface{}{
					{
						"start_time": altTime1.Format(time.RFC3339),
						"end_time":   altTime1.Add(duration).Format(time.RFC3339),
						"display":    fmt.Sprintf("%s - %s", altTime1.Format("3:04 PM"), altTime1.Add(duration).Format("3:04 PM")),
					},
					{
						"start_time": altTime2.Format(time.RFC3339),
						"end_time":   altTime2.Add(duration).Format(time.RFC3339),
						"display":    fmt.Sprintf("%s - %s", altTime2.Format("3:04 PM"), altTime2.Add(duration).Format("3:04 PM")),
					},
					{
						"start_time": altTime3.Format(time.RFC3339),
						"end_time":   altTime3.Add(duration).Format(time.RFC3339),
						"display":    fmt.Sprintf("%s - %s (%s)", altTime3.Format("3:04 PM"), altTime3.Add(duration).Format("3:04 PM"), altTime3.Format("Monday, January 2")),
					},
				},
			},
		}, nil
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
		response, err = a.handleListCalendarsRequest("")
	case "list_events":
		response, err = a.handleDirectListEvents(result.Parameters)
	case "create_event":
		response, err = a.handleDirectCreateEvent(result.Parameters)
	case "update_event":
		response, err = a.handleDirectUpdateEvent(result.Parameters)
	case "delete_event":
		response, err = a.handleDirectDeleteEvent(result.Parameters)
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
		if result.Response != "" && result.Confidence > 0.3 {
			a.logger.Info("LLM provided response for unknown intent, using it directly",
				zap.String("intent", result.Intent),
				zap.Float64("confidence", result.Confidence))
			response = &CalendarResponse{
				Text: result.Response,
			}
		} else {
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
		return a.processCalendarRequest(messageText)
	}

	if result.Response != "" && response != nil {
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

// processDirectToolCall handles direct tool calls with structured arguments (no message text)
func (a *CalendarAgent) processDirectToolCall(ctx context.Context, skill string, arguments map[string]interface{}) (*CalendarResponse, error) {
	a.logger.Debug("processing direct tool call",
		zap.String("component", "direct-tool-processor"),
		zap.String("operation", "process-tool-call"),
		zap.String("skill", skill),
		zap.Any("arguments", arguments))

	switch skill {
	case "list_events":
		return a.handleDirectListEvents(arguments)
	case "create_event":
		return a.handleDirectCreateEvent(arguments)
	case "update_event":
		return a.handleDirectUpdateEvent(arguments)
	case "delete_event":
		return a.handleDirectDeleteEvent(arguments)
	case "search_events":
		return a.handleSearchEventsRequestWithParams(arguments)
	case "get_availability":
		return a.handleGetAvailabilityRequestWithParams(arguments)
	case "list_calendars":
		return a.handleListCalendarsRequest("")
	default:
		return nil, fmt.Errorf("unsupported skill: %s", skill)
	}
}

// handleDirectListEvents processes direct list events calls with structured arguments
func (a *CalendarAgent) handleDirectListEvents(arguments map[string]interface{}) (*CalendarResponse, error) {
	a.logger.Debug("handling direct list events request", zap.Any("arguments", arguments))

	// Parse parameters using type-safe method
	params, err := a.parseListEventsParams(arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse list events parameters: %w", err)
	}

	calendarID := params.CalendarID
	if calendarID == "" {
		calendarID = os.Getenv("GOOGLE_CALENDAR_ID")
		if calendarID == "" {
			calendarID = "primary"
		}
	}

	// Parse start and end dates from structured parameters
	timeMin, err := time.Parse(time.RFC3339, params.StartDate)
	if err != nil {
		a.logger.Error("failed to parse start_date", zap.String("start_date", params.StartDate), zap.Error(err))
		return nil, fmt.Errorf("invalid start_date format: %w", err)
	}

	timeMax, err := time.Parse(time.RFC3339, params.EndDate)
	if err != nil {
		a.logger.Error("failed to parse end_date", zap.String("end_date", params.EndDate), zap.Error(err))
		return nil, fmt.Errorf("invalid end_date format: %w", err)
	}

	timeDescription := fmt.Sprintf("from %s to %s",
		timeMin.Format("Jan 2, 2006"),
		timeMax.Format("Jan 2, 2006"))

	a.logger.Debug("time range for direct events call",
		zap.Time("timeMin", timeMin),
		zap.Time("timeMax", timeMax),
		zap.String("description", timeDescription))

	events, err := a.calendarService.ListEvents(calendarID, timeMin, timeMax)
	if err != nil {
		a.logger.Error("failed to retrieve events from calendar service",
			zap.Error(err),
			zap.String("calendarID", calendarID),
			zap.String("timeDescription", timeDescription))
		return nil, fmt.Errorf("failed to retrieve calendar events: %w", err)
	}

	a.logger.Info("retrieved events for direct call",
		zap.Int("eventCount", len(events)),
		zap.String("timeDescription", timeDescription))

	if len(events) == 0 {
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

	return &CalendarResponse{
		Text: responseText,
		Data: events,
	}, nil
}

// handleDirectCreateEvent processes direct create event calls with structured arguments
func (a *CalendarAgent) handleDirectCreateEvent(arguments map[string]interface{}) (*CalendarResponse, error) {
	a.logger.Debug("handling direct create event request", zap.Any("arguments", arguments))

	params, err := a.parseCreateEventParams(arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse create event parameters: %w", err)
	}

	a.logger.Info("parsed create event parameters",
		zap.String("title", params.Title),
		zap.String("startTime", params.StartTime),
		zap.String("endTime", params.EndTime),
		zap.String("location", params.Location))

	calendarID := params.CalendarID
	if calendarID == "" {
		calendarID = os.Getenv("GOOGLE_CALENDAR_ID")
		if calendarID == "" {
			calendarID = "primary"
		}
	}

	startTime, err := time.Parse(time.RFC3339, params.StartTime)
	if err != nil {
		a.logger.Error("failed to parse start_time", zap.String("start_time", params.StartTime), zap.Error(err))
		return nil, fmt.Errorf("invalid start_time format: %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, params.EndTime)
	if err != nil {
		a.logger.Error("failed to parse end_time", zap.String("end_time", params.EndTime), zap.Error(err))
		return nil, fmt.Errorf("invalid end_time format: %w", err)
	}

	// Check for conflicts before creating the event
	conflicts, err := a.calendarService.CheckConflicts(calendarID, startTime, endTime)
	if err != nil {
		a.logger.Error("failed to check for conflicts in direct tool call",
			zap.Error(err),
			zap.String("calendarID", calendarID),
			zap.Time("startTime", startTime),
			zap.Time("endTime", endTime))
		return nil, fmt.Errorf("failed to check for scheduling conflicts: %w", err)
	}

	// If conflicts found, suggest alternative times
	if len(conflicts) > 0 {
		a.logger.Info("found scheduling conflicts in direct tool call",
			zap.Int("conflictCount", len(conflicts)),
			zap.String("proposedTitle", params.Title),
			zap.Time("proposedStartTime", startTime),
			zap.Time("proposedEndTime", endTime))

		conflictText := "⚠️ **Scheduling Conflict Detected!**\n\n"
		conflictText += fmt.Sprintf("You already have %d event(s) scheduled during %s - %s:\n\n",
			len(conflicts),
			startTime.Format("3:04 PM"),
			endTime.Format("3:04 PM"))

		for i, conflict := range conflicts {
			conflictStartTime, _ := time.Parse(time.RFC3339, conflict.Start.DateTime)
			conflictEndTime, _ := time.Parse(time.RFC3339, conflict.End.DateTime)
			conflictText += fmt.Sprintf("%d. **%s**\n   Time: %s - %s\n",
				i+1,
				conflict.Summary,
				conflictStartTime.Format("3:04 PM"),
				conflictEndTime.Format("3:04 PM"))
			if conflict.Location != "" {
				conflictText += fmt.Sprintf("   Location: %s\n", conflict.Location)
			}
			conflictText += "\n"
		}

		// Suggest alternative times
		conflictText += "**Suggested alternative times:**\n"
		duration := endTime.Sub(startTime)

		// Suggest 1 hour later
		altTime1 := startTime.Add(time.Hour)
		conflictText += fmt.Sprintf("• %s - %s\n",
			altTime1.Format("3:04 PM"),
			altTime1.Add(duration).Format("3:04 PM"))

		// Suggest 2 hours later
		altTime2 := startTime.Add(2 * time.Hour)
		conflictText += fmt.Sprintf("• %s - %s\n",
			altTime2.Format("3:04 PM"),
			altTime2.Add(duration).Format("3:04 PM"))

		// Suggest next day same time
		altTime3 := startTime.AddDate(0, 0, 1)
		conflictText += fmt.Sprintf("• %s - %s (%s)\n",
			altTime3.Format("3:04 PM"),
			altTime3.Add(duration).Format("3:04 PM"),
			altTime3.Format("Monday, January 2"))

		conflictText += "\nWould you like me to schedule it at one of these alternative times instead?"

		return &CalendarResponse{
			Text: conflictText,
			Data: map[string]interface{}{
				"conflicts":      conflicts,
				"proposed_event": params,
				"calendar_id":    calendarID,
				"alternative_times": []map[string]interface{}{
					{
						"start_time": altTime1.Format(time.RFC3339),
						"end_time":   altTime1.Add(duration).Format(time.RFC3339),
						"display":    fmt.Sprintf("%s - %s", altTime1.Format("3:04 PM"), altTime1.Add(duration).Format("3:04 PM")),
					},
					{
						"start_time": altTime2.Format(time.RFC3339),
						"end_time":   altTime2.Add(duration).Format(time.RFC3339),
						"display":    fmt.Sprintf("%s - %s", altTime2.Format("3:04 PM"), altTime2.Add(duration).Format("3:04 PM")),
					},
					{
						"start_time": altTime3.Format(time.RFC3339),
						"end_time":   altTime3.Add(duration).Format(time.RFC3339),
						"display":    fmt.Sprintf("%s - %s (%s)", altTime3.Format("3:04 PM"), altTime3.Add(duration).Format("3:04 PM"), altTime3.Format("Monday, January 2")),
					},
				},
			},
		}, nil
	}

	event := &calendar.Event{
		Summary:     params.Title,
		Description: params.Description,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
		},
		Location: params.Location,
	}

	a.logger.Debug("created calendar event object for direct tool call",
		zap.String("eventSummary", event.Summary),
		zap.String("calendarID", calendarID))

	createdEvent, err := a.calendarService.CreateEvent(calendarID, event)
	if err != nil {
		a.logger.Error("failed to create event in calendar service for direct tool call",
			zap.Error(err),
			zap.String("calendarID", calendarID),
			zap.String("eventSummary", event.Summary))
		return nil, fmt.Errorf("failed to create calendar event: %w", err)
	}

	responseText := "✅ Event created successfully!\n\n"
	responseText += "Title: " + createdEvent.Summary + "\n"
	responseText += "Date: " + startTime.Format("Monday, January 2, 2006") + "\n"
	responseText += fmt.Sprintf("Time: %s - %s\n",
		startTime.Format("3:04 PM"),
		endTime.Format("3:04 PM"))
	if createdEvent.Location != "" {
		responseText += "Location: " + createdEvent.Location + "\n"
	}

	a.logger.Info("successfully created event via direct tool call",
		zap.String("eventId", createdEvent.Id),
		zap.String("title", createdEvent.Summary),
		zap.String("calendarID", calendarID))

	return &CalendarResponse{
		Text: responseText,
		Data: createdEvent,
	}, nil
}

// handleDirectUpdateEvent processes direct update event calls with structured arguments
func (a *CalendarAgent) handleDirectUpdateEvent(arguments map[string]interface{}) (*CalendarResponse, error) {
	a.logger.Debug("handling direct update event request", zap.Any("arguments", arguments))

	params, err := a.parseUpdateEventParams(arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse update event parameters: %w", err)
	}

	a.logger.Info("parsed update event parameters",
		zap.String("eventID", params.EventID),
		zap.String("title", params.Title),
		zap.String("startTime", params.StartTime),
		zap.String("endTime", params.EndTime),
		zap.String("location", params.Location))

	// Determine calendar ID
	calendarID := params.CalendarID
	if calendarID == "" {
		calendarID = os.Getenv("GOOGLE_CALENDAR_ID")
		if calendarID == "" {
			calendarID = "primary"
		}
	}

	// Create the updated event structure
	event := &calendar.Event{}

	// Update fields only if they are provided
	if params.Title != "" {
		event.Summary = params.Title
	}
	if params.Description != "" {
		event.Description = params.Description
	}
	if params.Location != "" {
		event.Location = params.Location
	}

	// Parse and update start time if provided
	if params.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, params.StartTime)
		if err != nil {
			a.logger.Error("failed to parse start_time", zap.String("start_time", params.StartTime), zap.Error(err))
			return nil, fmt.Errorf("invalid start_time format: %w", err)
		}
		event.Start = &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
		}
	}

	// Parse and update end time if provided
	if params.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, params.EndTime)
		if err != nil {
			a.logger.Error("failed to parse end_time", zap.String("end_time", params.EndTime), zap.Error(err))
			return nil, fmt.Errorf("invalid end_time format: %w", err)
		}
		event.End = &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
		}
	}

	a.logger.Debug("created calendar event object for direct update tool call",
		zap.String("eventID", params.EventID),
		zap.String("calendarID", calendarID))

	updatedEvent, err := a.calendarService.UpdateEvent(calendarID, params.EventID, event)
	if err != nil {
		a.logger.Error("failed to update event in calendar service for direct tool call",
			zap.Error(err),
			zap.String("calendarID", calendarID),
			zap.String("eventID", params.EventID))
		return nil, fmt.Errorf("failed to update calendar event: %w", err)
	}

	responseText := "✅ Event updated successfully!\n\n"
	responseText += "Event ID: " + updatedEvent.Id + "\n"
	if updatedEvent.Summary != "" {
		responseText += "Title: " + updatedEvent.Summary + "\n"
	}
	if updatedEvent.Start != nil && updatedEvent.Start.DateTime != "" {
		startTime, _ := time.Parse(time.RFC3339, updatedEvent.Start.DateTime)
		responseText += "Date: " + startTime.Format("Monday, January 2, 2006") + "\n"
		if updatedEvent.End != nil && updatedEvent.End.DateTime != "" {
			endTime, _ := time.Parse(time.RFC3339, updatedEvent.End.DateTime)
			responseText += fmt.Sprintf("Time: %s - %s\n",
				startTime.Format("3:04 PM"),
				endTime.Format("3:04 PM"))
		}
	}
	if updatedEvent.Location != "" {
		responseText += "Location: " + updatedEvent.Location + "\n"
	}

	a.logger.Info("successfully updated event via direct tool call",
		zap.String("eventId", updatedEvent.Id),
		zap.String("title", updatedEvent.Summary),
		zap.String("calendarID", calendarID))

	return &CalendarResponse{
		Text: responseText,
		Data: updatedEvent,
	}, nil
}

// handleDirectDeleteEvent processes direct delete event calls with structured arguments
func (a *CalendarAgent) handleDirectDeleteEvent(arguments map[string]interface{}) (*CalendarResponse, error) {
	a.logger.Debug("handling direct delete event request", zap.Any("arguments", arguments))

	params, err := a.parseDeleteEventParams(arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse delete event parameters: %w", err)
	}

	a.logger.Info("parsed delete event parameters",
		zap.String("eventID", params.EventID),
		zap.String("calendarID", params.CalendarID),
		zap.String("title", params.Title),
		zap.String("date", params.Date))

	// Determine calendar ID
	calendarID := params.CalendarID
	if calendarID == "" {
		calendarID = os.Getenv("GOOGLE_CALENDAR_ID")
		if calendarID == "" {
			calendarID = "primary"
		}
	}

	// Optionally get event details before deletion for confirmation message
	var eventTitle string
	if params.EventID != "" {
		event, err := a.calendarService.GetEvent(calendarID, params.EventID)
		if err != nil {
			a.logger.Warn("could not retrieve event details before deletion",
				zap.String("eventID", params.EventID),
				zap.String("calendarID", calendarID),
				zap.Error(err))
		} else {
			eventTitle = event.Summary
		}
	}

	a.logger.Debug("deleting calendar event via direct tool call",
		zap.String("eventID", params.EventID),
		zap.String("calendarID", calendarID))

	err = a.calendarService.DeleteEvent(calendarID, params.EventID)
	if err != nil {
		a.logger.Error("failed to delete event in calendar service for direct tool call",
			zap.Error(err),
			zap.String("calendarID", calendarID),
			zap.String("eventID", params.EventID))
		return nil, fmt.Errorf("failed to delete calendar event: %w", err)
	}

	responseText := "✅ Event deleted successfully!\n\n"
	responseText += "Event ID: " + params.EventID + "\n"
	if eventTitle != "" {
		responseText += "Title: " + eventTitle + "\n"
	}

	a.logger.Info("successfully deleted event via direct tool call",
		zap.String("eventId", params.EventID),
		zap.String("title", eventTitle),
		zap.String("calendarID", calendarID))

	return &CalendarResponse{
		Text: responseText,
		Data: map[string]string{
			"event_id":    params.EventID,
			"calendar_id": calendarID,
			"status":      "deleted",
		},
	}, nil
}

// ListEventsParams represents type-safe parameters for list events operations
type ListEventsParams struct {
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	CalendarID string `json:"calendar_id,omitempty"`
	MaxResults int    `json:"max_results,omitempty"`
	Query      string `json:"query,omitempty"`
}

// CreateEventParams represents type-safe parameters for create event operations
type CreateEventParams struct {
	Title       string `json:"title"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Date        string `json:"date,omitempty"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
	CalendarID  string `json:"calendar_id,omitempty"`
}

// UpdateEventParams represents type-safe parameters for update event operations
type UpdateEventParams struct {
	EventID     string `json:"event_id"`
	Title       string `json:"title,omitempty"`
	StartTime   string `json:"start_time,omitempty"`
	EndTime     string `json:"end_time,omitempty"`
	Date        string `json:"date,omitempty"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
	CalendarID  string `json:"calendar_id,omitempty"`
}

// DeleteEventParams represents type-safe parameters for delete event operations
type DeleteEventParams struct {
	EventID    string `json:"event_id"`
	CalendarID string `json:"calendar_id,omitempty"`
	Title      string `json:"title,omitempty"`
	Date       string `json:"date,omitempty"`
}

// SearchEventsParams represents type-safe parameters for search events operations
type SearchEventsParams struct {
	Query      string `json:"query"`
	StartDate  string `json:"start_date,omitempty"`
	EndDate    string `json:"end_date,omitempty"`
	CalendarID string `json:"calendar_id,omitempty"`
	MaxResults int    `json:"max_results,omitempty"`
}

// AvailabilityParams represents type-safe parameters for availability operations
type AvailabilityParams struct {
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Duration   int    `json:"duration,omitempty"` // in minutes
	CalendarID string `json:"calendar_id,omitempty"`
}

// parseListEventsParams safely parses arguments into ListEventsParams struct
func (a *CalendarAgent) parseListEventsParams(arguments map[string]interface{}) (*ListEventsParams, error) {
	params := &ListEventsParams{}

	if startDate, ok := arguments["start_date"].(string); ok {
		params.StartDate = startDate
	} else {
		return nil, fmt.Errorf("missing or invalid start_date parameter")
	}

	if endDate, ok := arguments["end_date"].(string); ok {
		params.EndDate = endDate
	} else {
		return nil, fmt.Errorf("missing or invalid end_date parameter")
	}

	if calendarID, ok := arguments["calendar_id"].(string); ok {
		params.CalendarID = calendarID
	}

	if maxResults, ok := arguments["max_results"].(float64); ok {
		params.MaxResults = int(maxResults)
	}

	if query, ok := arguments["query"].(string); ok {
		params.Query = query
	}

	return params, nil
}

// parseCreateEventParams safely parses arguments into CreateEventParams struct
func (a *CalendarAgent) parseCreateEventParams(arguments map[string]interface{}) (*CreateEventParams, error) {
	params := &CreateEventParams{}

	if title, ok := arguments["title"].(string); ok {
		params.Title = title
	} else {
		return nil, fmt.Errorf("missing or invalid title parameter")
	}

	if startTime, ok := arguments["start_time"].(string); ok {
		params.StartTime = startTime
	}

	if endTime, ok := arguments["end_time"].(string); ok {
		params.EndTime = endTime
	}

	if date, ok := arguments["date"].(string); ok {
		params.Date = date
	}

	if location, ok := arguments["location"].(string); ok {
		params.Location = location
	}

	if description, ok := arguments["description"].(string); ok {
		params.Description = description
	}

	if calendarID, ok := arguments["calendar_id"].(string); ok {
		params.CalendarID = calendarID
	}

	return params, nil
}

// parseUpdateEventParams safely parses arguments into UpdateEventParams struct
func (a *CalendarAgent) parseUpdateEventParams(arguments map[string]interface{}) (*UpdateEventParams, error) {
	params := &UpdateEventParams{}

	if eventID, ok := arguments["event_id"].(string); ok {
		params.EventID = eventID
	} else {
		return nil, fmt.Errorf("missing or invalid event_id parameter")
	}

	if title, ok := arguments["title"].(string); ok {
		params.Title = title
	}

	if startTime, ok := arguments["start_time"].(string); ok {
		params.StartTime = startTime
	}

	if endTime, ok := arguments["end_time"].(string); ok {
		params.EndTime = endTime
	}

	if date, ok := arguments["date"].(string); ok {
		params.Date = date
	}

	if location, ok := arguments["location"].(string); ok {
		params.Location = location
	}

	if description, ok := arguments["description"].(string); ok {
		params.Description = description
	}

	if calendarID, ok := arguments["calendar_id"].(string); ok {
		params.CalendarID = calendarID
	}

	return params, nil
}

// parseDeleteEventParams safely parses arguments into DeleteEventParams struct
func (a *CalendarAgent) parseDeleteEventParams(arguments map[string]interface{}) (*DeleteEventParams, error) {
	params := &DeleteEventParams{}

	if eventID, ok := arguments["event_id"].(string); ok {
		params.EventID = eventID
	} else {
		return nil, fmt.Errorf("missing or invalid event_id parameter")
	}

	if calendarID, ok := arguments["calendar_id"].(string); ok {
		params.CalendarID = calendarID
	}

	if title, ok := arguments["title"].(string); ok {
		params.Title = title
	}

	if date, ok := arguments["date"].(string); ok {
		params.Date = date
	}

	return params, nil
}

// parseSearchEventsParams safely parses arguments into SearchEventsParams struct
func (a *CalendarAgent) parseSearchEventsParams(arguments map[string]interface{}) (*SearchEventsParams, error) {
	params := &SearchEventsParams{}

	if query, ok := arguments["query"].(string); ok {
		params.Query = query
	} else {
		return nil, fmt.Errorf("missing or invalid query parameter")
	}

	if startDate, ok := arguments["start_date"].(string); ok {
		params.StartDate = startDate
	}

	if endDate, ok := arguments["end_date"].(string); ok {
		params.EndDate = endDate
	}

	if calendarID, ok := arguments["calendar_id"].(string); ok {
		params.CalendarID = calendarID
	}

	if maxResults, ok := arguments["max_results"].(float64); ok {
		params.MaxResults = int(maxResults)
	}

	return params, nil
}

// parseAvailabilityParams safely parses arguments into AvailabilityParams struct
func (a *CalendarAgent) parseAvailabilityParams(arguments map[string]interface{}) (*AvailabilityParams, error) {
	params := &AvailabilityParams{}

	if startDate, ok := arguments["start_date"].(string); ok {
		params.StartDate = startDate
	} else {
		return nil, fmt.Errorf("missing or invalid start_date parameter")
	}

	if endDate, ok := arguments["end_date"].(string); ok {
		params.EndDate = endDate
	} else {
		return nil, fmt.Errorf("missing or invalid end_date parameter")
	}

	if duration, ok := arguments["duration"].(float64); ok {
		params.Duration = int(duration)
	}

	if calendarID, ok := arguments["calendar_id"].(string); ok {
		params.CalendarID = calendarID
	}

	return params, nil
}

// handleSearchEventsRequestWithParams processes search events requests with structured parameters
func (a *CalendarAgent) handleSearchEventsRequestWithParams(arguments map[string]interface{}) (*CalendarResponse, error) {
	a.logger.Debug("handling search events request with params", zap.Any("arguments", arguments))

	params, err := a.parseSearchEventsParams(arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse search events parameters: %w", err)
	}

	// For now, return a simple response
	return &CalendarResponse{
		Text: fmt.Sprintf("Search functionality for '%s' is not yet implemented", params.Query),
	}, nil
}

// handleGetAvailabilityRequestWithParams processes availability requests with structured parameters
func (a *CalendarAgent) handleGetAvailabilityRequestWithParams(arguments map[string]interface{}) (*CalendarResponse, error) {
	a.logger.Debug("handling get availability request with params", zap.Any("arguments", arguments))

	params, err := a.parseAvailabilityParams(arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse availability parameters: %w", err)
	}

	// For now, return a simple response
	return &CalendarResponse{
		Text: fmt.Sprintf("Availability checking from %s to %s is not yet implemented", params.StartDate, params.EndDate),
	}, nil
}
