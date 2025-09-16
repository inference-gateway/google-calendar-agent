package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	server "github.com/inference-gateway/adk/server"
	google "github.com/inference-gateway/google-calendar-agent/internal/google"
	zap "go.uber.org/zap"
	calendar "google.golang.org/api/calendar/v3"
)

// FindAvailableTimeSkill struct holds the skill with dependencies
type FindAvailableTimeSkill struct {
	logger *zap.Logger
	google google.CalendarService
}

// NewFindAvailableTimeSkill creates a new find_available_time skill
func NewFindAvailableTimeSkill(logger *zap.Logger, google google.CalendarService) server.Tool {
	skill := &FindAvailableTimeSkill{
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
		skill.FindAvailableTimeHandler,
	)
}

// FindAvailableTimeHandler handles the find_available_time skill execution
func (s *FindAvailableTimeSkill) FindAvailableTimeHandler(ctx context.Context, args map[string]any) (string, error) {
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
	existingEvents, err := s.google.ListEvents(calendarID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to list events for availability check", zap.Error(err))
		return "", fmt.Errorf("failed to list events for availability check: %w", err)
	}

	availableSlots := s.findAvailableSlots(startDate, endDate, time.Duration(duration)*time.Minute, existingEvents)

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

// findAvailableSlots finds available time slots between existing events
func (s *FindAvailableTimeSkill) findAvailableSlots(startDate, endDate time.Time, duration time.Duration, events []*calendar.Event) []timeSlot {
	// Create a list of busy periods
	var busyPeriods []timeSlot
	for _, event := range events {
		if event.Start != nil && event.End != nil {
			eventStart, err1 := time.Parse(time.RFC3339, event.Start.DateTime)
			eventEnd, err2 := time.Parse(time.RFC3339, event.End.DateTime)

			if err1 == nil && err2 == nil {
				busyPeriods = append(busyPeriods, timeSlot{
					startTime: eventStart,
					endTime:   eventEnd,
					duration:  eventEnd.Sub(eventStart),
				})
			}
		}
	}

	// Sort busy periods by start time
	for i := 0; i < len(busyPeriods)-1; i++ {
		for j := i + 1; j < len(busyPeriods); j++ {
			if busyPeriods[i].startTime.After(busyPeriods[j].startTime) {
				busyPeriods[i], busyPeriods[j] = busyPeriods[j], busyPeriods[i]
			}
		}
	}

	// Find gaps between busy periods that can accommodate the requested duration
	var availableSlots []timeSlot

	// Check for availability from start date to first event
	if len(busyPeriods) > 0 {
		if busyPeriods[0].startTime.Sub(startDate) >= duration {
			availableSlots = append(availableSlots, timeSlot{
				startTime: startDate,
				endTime:   startDate.Add(duration),
				duration:  duration,
			})
		}
	} else {
		// No events, entire period is available
		availableSlots = append(availableSlots, timeSlot{
			startTime: startDate,
			endTime:   startDate.Add(duration),
			duration:  duration,
		})
		return availableSlots
	}

	// Check gaps between events
	for i := 0; i < len(busyPeriods)-1; i++ {
		gapStart := busyPeriods[i].endTime
		gapEnd := busyPeriods[i+1].startTime
		gapDuration := gapEnd.Sub(gapStart)

		if gapDuration >= duration {
			availableSlots = append(availableSlots, timeSlot{
				startTime: gapStart,
				endTime:   gapStart.Add(duration),
				duration:  duration,
			})
		}
	}

	// Check for availability from last event to end date
	if len(busyPeriods) > 0 {
		lastEventEnd := busyPeriods[len(busyPeriods)-1].endTime
		if endDate.Sub(lastEventEnd) >= duration {
			availableSlots = append(availableSlots, timeSlot{
				startTime: lastEventEnd,
				endTime:   lastEventEnd.Add(duration),
				duration:  duration,
			})
		}
	}

	return availableSlots
}
