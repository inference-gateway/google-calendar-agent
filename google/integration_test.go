package google_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/inference-gateway/google-calendar-agent/config"
	"github.com/inference-gateway/google-calendar-agent/google"
)

func TestCalendarServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Skip("Integration tests require Google Calendar API credentials")

	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	cfg := &config.Config{
		Google: config.GoogleConfig{
			CalendarID: "primary",
			ReadOnly:   false,
		},
	}

	service, err := google.NewCalendarService(ctx, cfg, logger, option.WithCredentialsFile("testdata/service-account.json"))
	require.NoError(t, err)

	t.Run("ListCalendars", func(t *testing.T) {
		calendars, err := service.ListCalendars()
		assert.NoError(t, err)
		assert.NotNil(t, calendars)
	})

	t.Run("ListEvents", func(t *testing.T) {
		timeMin := time.Now().Add(-24 * time.Hour)
		timeMax := time.Now().Add(24 * time.Hour)

		events, err := service.ListEvents("primary", timeMin, timeMax)
		assert.NoError(t, err)
		assert.NotNil(t, events)
	})

	t.Run("CreateAndDeleteEvent", func(t *testing.T) {
		event := &calendar.Event{
			Summary:     "Test Event",
			Description: "Created by integration test",
			Start: &calendar.EventDateTime{
				DateTime: time.Now().Add(time.Hour).Format(time.RFC3339),
				TimeZone: "UTC",
			},
			End: &calendar.EventDateTime{
				DateTime: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
				TimeZone: "UTC",
			},
		}

		createdEvent, err := service.CreateEvent("primary", event)
		require.NoError(t, err)
		require.NotNil(t, createdEvent)
		require.NotEmpty(t, createdEvent.Id)

		defer func() {
			err := service.DeleteEvent("primary", createdEvent.Id)
			assert.NoError(t, err)
		}()

		retrievedEvent, err := service.GetEvent("primary", createdEvent.Id)
		assert.NoError(t, err)
		assert.Equal(t, createdEvent.Summary, retrievedEvent.Summary)

		updatedEvent := &calendar.Event{
			Summary:     "Updated Test Event",
			Description: "Updated by integration test",
		}

		resultEvent, err := service.UpdateEvent("primary", createdEvent.Id, updatedEvent)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Test Event", resultEvent.Summary)
	})
}

func TestNewCalendarService(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	cfg := &config.Config{
		Google: config.GoogleConfig{
			CalendarID: "primary",
			ReadOnly:   false,
		},
	}

	t.Run("SuccessWithValidOptions", func(t *testing.T) {
		service, err := google.NewCalendarService(ctx, cfg, logger)

		if err == nil {
			assert.NotNil(t, service)
			assert.Implements(t, (*google.CalendarService)(nil), service)
		} else {
			assert.Error(t, err)
			assert.Nil(t, service)
		}
	})

	t.Run("WithCustomOptions", func(t *testing.T) {
		service, err := google.NewCalendarService(ctx, cfg, logger, option.WithCredentialsFile("nonexistent.json"))

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "unable to create calendar service")
	})

	t.Run("ReadOnlyConfiguration", func(t *testing.T) {
		readOnlyCfg := &config.Config{
			Google: config.GoogleConfig{
				CalendarID: "primary",
				ReadOnly:   true,
			},
		}

		service, err := google.NewCalendarService(ctx, readOnlyCfg, logger)

		if err == nil {
			assert.NotNil(t, service)
			assert.Implements(t, (*google.CalendarService)(nil), service)
		} else {
			assert.Error(t, err)
			assert.Nil(t, service)
		}
	})
}
