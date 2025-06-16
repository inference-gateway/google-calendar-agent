package toolbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	a2a "github.com/inference-gateway/google-calendar-agent/a2a"
	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"
)

// handleListEvents handles the list events tool call with A2A structured response
func (g *GoogleCalendarTools) handleListEvents(ctx context.Context, args map[string]interface{}) (string, error) {
	g.logger.Info("🔧 Tool called: list_calendar_events", zap.Any("args", args))

	if ctx != nil {
		g.logger.Debug("checking context for A2A information", zap.Any("context_type", fmt.Sprintf("%T", ctx)))
	}

	if g.isMockMode {
		g.logger.Debug("returning mock events")
		return g.getMockEvents(), nil
	}

	g.logger.Debug("processing list events request in non-mock mode")

	timeMin := time.Now()
	if val, ok := args["timeMin"].(string); ok && val != "" {
		if parsedTime, err := time.Parse(time.RFC3339, val); err == nil {
			timeMin = parsedTime
		}
	}

	timeMax := timeMin.Add(24 * time.Hour)
	if val, ok := args["timeMax"].(string); ok && val != "" {
		if parsedTime, err := time.Parse(time.RFC3339, val); err == nil {
			timeMax = parsedTime
		}
	}

	events, err := g.calSvc.ListEvents(g.config.Google.CalendarID, timeMin, timeMax)
	if err != nil {
		return "", fmt.Errorf("failed to list events: %w", err)
	}

	response := a2a.CalendarEventResponse{
		Events:  events,
		Message: fmt.Sprintf("Found %d events between %s and %s", len(events), timeMin.Format("2006-01-02 15:04"), timeMax.Format("2006-01-02 15:04")),
		Success: true,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonResponse), nil
}

// handleCreateEvent handles the create event tool call
func (g *GoogleCalendarTools) handleCreateEvent(ctx context.Context, args map[string]interface{}) (string, error) {
	g.logger.Debug("handleCreateEvent called with args", zap.Any("args", args))

	if g.isMockMode {
		g.logger.Debug("returning mock create event response")
		return g.getMockCreateEvent(args), nil
	}

	g.logger.Debug("processing create event request in non-mock mode")

	summary, ok := args["summary"].(string)
	if !ok || summary == "" {
		return "", fmt.Errorf("summary is required")
	}

	startTimeStr, ok := args["startTime"].(string)
	if !ok || startTimeStr == "" {
		return "", fmt.Errorf("startTime is required")
	}

	endTimeStr, ok := args["endTime"].(string)
	if !ok || endTimeStr == "" {
		return "", fmt.Errorf("endTime is required")
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return "", fmt.Errorf("invalid startTime format: %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return "", fmt.Errorf("invalid endTime format: %w", err)
	}

	if endTime.Before(startTime) {
		return "", fmt.Errorf("endTime must be after startTime")
	}

	event := &calendar.Event{
		Summary: summary,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
		},
	}

	if desc, ok := args["description"].(string); ok && desc != "" {
		event.Description = desc
	}

	if loc, ok := args["location"].(string); ok && loc != "" {
		event.Location = loc
	}

	if attendeesRaw, ok := args["attendees"]; ok {
		if attendeesList, ok := attendeesRaw.([]interface{}); ok {
			var attendees []*calendar.EventAttendee
			for _, attendeeRaw := range attendeesList {
				if email, ok := attendeeRaw.(string); ok {
					attendees = append(attendees, &calendar.EventAttendee{Email: email})
				}
			}
			event.Attendees = attendees
		}
	}

	createdEvent, err := g.calSvc.CreateEvent(g.config.Google.CalendarID, event)
	if err != nil {
		return "", fmt.Errorf("failed to create event: %w", err)
	}

	response := a2a.CalendarEventResponse{
		Event:   createdEvent,
		Message: fmt.Sprintf("Event '%s' created successfully", createdEvent.Summary),
		Success: true,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonResponse), nil
}

// handleUpdateEvent handles the update event tool call
func (g *GoogleCalendarTools) handleUpdateEvent(ctx context.Context, args map[string]interface{}) (string, error) {
	if g.isMockMode {
		return g.getMockUpdateEvent(args), nil
	}

	eventId, ok := args["eventId"].(string)
	if !ok || eventId == "" {
		return "", fmt.Errorf("eventId is required")
	}

	existingEvent, err := g.calSvc.GetEvent(g.config.Google.CalendarID, eventId)
	if err != nil {
		return "", fmt.Errorf("failed to get existing event: %w", err)
	}

	if summary, ok := args["summary"].(string); ok && summary != "" {
		existingEvent.Summary = summary
	}

	if desc, ok := args["description"].(string); ok && desc != "" {
		existingEvent.Description = desc
	}

	if loc, ok := args["location"].(string); ok && loc != "" {
		existingEvent.Location = loc
	}

	if startTimeStr, ok := args["startTime"].(string); ok && startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			existingEvent.Start.DateTime = startTime.Format(time.RFC3339)
		}
	}

	if endTimeStr, ok := args["endTime"].(string); ok && endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			existingEvent.End.DateTime = endTime.Format(time.RFC3339)
		}
	}

	updatedEvent, err := g.calSvc.UpdateEvent(g.config.Google.CalendarID, eventId, existingEvent)
	if err != nil {
		return "", fmt.Errorf("failed to update event: %w", err)
	}

	result := map[string]interface{}{
		"success": true,
		"eventId": updatedEvent.Id,
		"message": "Event updated successfully",
		"event":   updatedEvent,
		"mock":    false,
	}

	response, _ := json.Marshal(result)
	return string(response), nil
}

// handleDeleteEvent handles the delete event tool call
func (g *GoogleCalendarTools) handleDeleteEvent(ctx context.Context, args map[string]interface{}) (string, error) {
	if g.isMockMode {
		return g.getMockDeleteEvent(args), nil
	}

	eventId, ok := args["eventId"].(string)
	if !ok || eventId == "" {
		return "", fmt.Errorf("eventId is required")
	}

	err := g.calSvc.DeleteEvent(g.config.Google.CalendarID, eventId)
	if err != nil {
		return "", fmt.Errorf("failed to delete event: %w", err)
	}

	result := map[string]interface{}{
		"success": true,
		"eventId": eventId,
		"message": "Event deleted successfully",
		"mock":    false,
	}

	response, _ := json.Marshal(result)
	return string(response), nil
}

// handleGetEvent handles the get event tool call
func (g *GoogleCalendarTools) handleGetEvent(ctx context.Context, args map[string]interface{}) (string, error) {
	if g.isMockMode {
		return g.getMockGetEvent(args), nil
	}

	eventId, ok := args["eventId"].(string)
	if !ok || eventId == "" {
		return "", fmt.Errorf("eventId is required")
	}

	event, err := g.calSvc.GetEvent(g.config.Google.CalendarID, eventId)
	if err != nil {
		return "", fmt.Errorf("failed to get event: %w", err)
	}

	result := map[string]interface{}{
		"success": true,
		"event":   event,
		"mock":    false,
	}

	response, _ := json.Marshal(result)
	return string(response), nil
}

// handleFindAvailableTime handles the find available time tool call
func (g *GoogleCalendarTools) handleFindAvailableTime(ctx context.Context, args map[string]interface{}) (string, error) {
	if g.isMockMode {
		return g.getMockAvailableTime(args), nil
	}

	// TODO: Implement real availability search
	// For now, return mock response
	return g.getMockAvailableTime(args), nil
}

// handleCheckConflicts handles the check conflicts tool call
func (g *GoogleCalendarTools) handleCheckConflicts(ctx context.Context, args map[string]interface{}) (string, error) {
	if g.isMockMode {
		return g.getMockConflicts(args), nil
	}

	startTimeStr, ok := args["startTime"].(string)
	if !ok || startTimeStr == "" {
		return "", fmt.Errorf("startTime is required")
	}

	endTimeStr, ok := args["endTime"].(string)
	if !ok || endTimeStr == "" {
		return "", fmt.Errorf("endTime is required")
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return "", fmt.Errorf("invalid startTime format: %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return "", fmt.Errorf("invalid endTime format: %w", err)
	}

	conflicts, err := g.calSvc.CheckConflicts(g.config.Google.CalendarID, startTime, endTime)
	if err != nil {
		return "", fmt.Errorf("failed to check conflicts: %w", err)
	}

	result := map[string]interface{}{
		"hasConflicts":   len(conflicts) > 0,
		"conflictCount":  len(conflicts),
		"conflictEvents": conflicts,
		"timeRange": map[string]string{
			"start": startTimeStr,
			"end":   endTimeStr,
		},
		"mock": false,
	}

	response, _ := json.Marshal(result)
	return string(response), nil
}
