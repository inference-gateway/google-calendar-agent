package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"

	server "github.com/inference-gateway/adk/server"

	google "github.com/inference-gateway/google-calendar-agent/internal/google"
)

// FindAvailableTimeTool struct holds the tool with dependencies
type FindAvailableTimeTool struct {
	logger *zap.Logger
	google google.CalendarService
}

// NewFindAvailableTimeTool creates a new find_available_time tool
func NewFindAvailableTimeTool(logger *zap.Logger, google google.CalendarService) server.Tool {
	tool := &FindAvailableTimeTool{
		logger: logger,
		google: google,
	}
	return server.NewBasicTool(
		"find_available_time",
		"Find available time slots in the calendar",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"duration": map[string]any{
					"description": "Duration in minutes for the desired time slot (default: 60)",
					"maximum":     480,
					"minimum":     15,
					"type":        "integer",
				},
				"endDate": map[string]any{
					"description": "End date for search (RFC3339 format, e.g., 2024-01-01T23:59:59Z)",
					"type":        "string",
				},
				"startDate": map[string]any{
					"description": "Start date for search (RFC3339 format, e.g., 2024-01-01T00:00:00Z)",
					"type":        "string",
				},
			},
			"required": []string{"startDate", "endDate"},
		},
		tool.FindAvailableTimeHandler,
	)
}

// FindAvailableTimeHandler handles the find_available_time tool execution
func (s *FindAvailableTimeTool) FindAvailableTimeHandler(ctx context.Context, args map[string]any) (string, error) {
	span := startToolSpan(ctx, "find_available_time")
	defer span.End()
	s.logger.Debug("finding available time", zap.Any("args", args))

	startDateStr, ok := args["startDate"].(string)
	if !ok || startDateStr == "" {
		return "", fmt.Errorf("startDate is required")
	}

	endDateStr, ok := args["endDate"].(string)
	if !ok || endDateStr == "" {
		return "", fmt.Errorf("endDate is required")
	}

	duration := 60
	if d, exists := args["duration"]; exists && d != nil {
		if dFloat, ok := d.(float64); ok {
			duration = int(dFloat)
		}
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return "", fmt.Errorf("invalid startDate format: %w", err)
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return "", fmt.Errorf("invalid endDate format: %w", err)
	}

	calendarID := s.google.GetCalendarID()
	existingEvents, calendarTimeZone, err := s.google.ListEvents(calendarID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to list events for availability check", zap.Error(err))
		return "", fmt.Errorf("failed to list events for availability check: %w", err)
	}

	userLocation, _, _ := resolveTimezone()
	loc := google.LocationOrDefault(calendarTimeZone, userLocation)
	availableSlots, truncated := s.findAvailableSlots(startDate, endDate, time.Duration(duration)*time.Minute, existingEvents, loc)

	s.logger.Info("available time slots found", zap.Int("slotCount", len(availableSlots)))

	var slots []map[string]any
	for _, slot := range availableSlots {
		slots = append(slots, map[string]any{
			"startTime": slot.startTime.Format(time.RFC3339),
			"endTime":   slot.endTime.Format(time.RFC3339),
			"duration":  int(slot.duration.Minutes()),
		})
	}

	result := map[string]any{
		"success":           true,
		"availableSlots":    slots,
		"slotCount":         len(slots),
		"truncated":         truncated,
		"requestedDuration": duration,
		"searchRange": map[string]string{
			"startDate": startDateStr,
			"endDate":   endDateStr,
		},
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}

// timeSlot represents an available time slot
type timeSlot struct {
	startTime time.Time
	endTime   time.Time
	duration  time.Duration
}

// maxAvailableSlots caps how many slots one call returns so a wide search window
// does not flood the model. The response flags truncation so it can narrow the window.
const maxAvailableSlots = 50

// findAvailableSlots walks the gaps between busy periods inside [startDate, endDate)
// and emits every consecutive slot of the requested duration, placing all-day events in loc.
// The second result reports whether the cap cut the list short.
func (s *FindAvailableTimeTool) findAvailableSlots(startDate, endDate time.Time, duration time.Duration, events []*calendar.Event, loc *time.Location) ([]timeSlot, bool) {
	busyPeriods := mergeBusyPeriods(collectBusyPeriods(events, loc))
	searchEnd := timeSlot{startTime: endDate, endTime: endDate}

	var slots []timeSlot
	cursor := startDate
	for _, period := range append(busyPeriods, searchEnd) {
		for !cursor.Add(duration).After(period.startTime) {
			if len(slots) == maxAvailableSlots {
				return slots, true
			}
			slots = append(slots, timeSlot{startTime: cursor, endTime: cursor.Add(duration), duration: duration})
			cursor = cursor.Add(duration)
		}
		if period.endTime.After(cursor) {
			cursor = period.endTime
		}
	}
	return slots, false
}

// collectBusyPeriods converts events into busy intervals, placing all-day events in loc.
func collectBusyPeriods(events []*calendar.Event, loc *time.Location) []timeSlot {
	var busyPeriods []timeSlot
	for _, event := range events {
		eventStart, eventEnd, ok := google.EventTimeRange(event, loc)
		if !ok {
			continue
		}
		busyPeriods = append(busyPeriods, timeSlot{
			startTime: eventStart,
			endTime:   eventEnd,
			duration:  eventEnd.Sub(eventStart),
		})
	}
	return busyPeriods
}

// mergeBusyPeriods sorts busy intervals and folds overlapping or touching ones together
// so that a gap is only reported when no event covers it.
func mergeBusyPeriods(busyPeriods []timeSlot) []timeSlot {
	sort.Slice(busyPeriods, func(i, j int) bool {
		return busyPeriods[i].startTime.Before(busyPeriods[j].startTime)
	})

	var merged []timeSlot
	for _, period := range busyPeriods {
		last := len(merged) - 1
		if last >= 0 && !period.startTime.After(merged[last].endTime) {
			if period.endTime.After(merged[last].endTime) {
				merged[last].endTime = period.endTime
				merged[last].duration = merged[last].endTime.Sub(merged[last].startTime)
			}
			continue
		}
		merged = append(merged, period)
	}
	return merged
}
