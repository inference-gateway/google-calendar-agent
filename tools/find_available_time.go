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
func (s *FindAvailableTimeTool) findAvailableSlots(startDate, endDate time.Time, duration time.Duration, events []*calendar.Event) []timeSlot {
	loc, _, _ := resolveTimezone()

	var busyPeriods []timeSlot
	for _, event := range events {
		if event.Start == nil || event.End == nil {
			continue
		}

		if event.Start.DateTime != "" {
			eventStart, err1 := time.Parse(time.RFC3339, event.Start.DateTime)
			eventEnd, err2 := time.Parse(time.RFC3339, event.End.DateTime)
			if err1 == nil && err2 == nil {
				busyPeriods = append(busyPeriods, timeSlot{
					startTime: eventStart,
					endTime:   eventEnd,
					duration:  eventEnd.Sub(eventStart),
				})
			}
			continue
		}

		if event.Start.Date != "" && event.End.Date != "" {
			startDay, err1 := time.ParseInLocation("2006-01-02", event.Start.Date, loc)
			endDay, err2 := time.ParseInLocation("2006-01-02", event.End.Date, loc)
			if err1 == nil && err2 == nil {
				busyPeriods = append(busyPeriods, timeSlot{
					startTime: startDay,
					endTime:   endDay,
					duration:  endDay.Sub(startDay),
				})
			}
		}
	}

	sort.Slice(busyPeriods, func(i, j int) bool {
		return busyPeriods[i].startTime.Before(busyPeriods[j].startTime)
	})

	var availableSlots []timeSlot

	if len(busyPeriods) > 0 {
		if busyPeriods[0].startTime.Sub(startDate) >= duration {
			availableSlots = append(availableSlots, timeSlot{
				startTime: startDate,
				endTime:   startDate.Add(duration),
				duration:  duration,
			})
		}
	} else {
		availableSlots = append(availableSlots, timeSlot{
			startTime: startDate,
			endTime:   startDate.Add(duration),
			duration:  duration,
		})
		return availableSlots
	}

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
