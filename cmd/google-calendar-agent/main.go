package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/option"

	"github.com/inference-gateway/google-calendar-agent/a2a"
	"github.com/inference-gateway/google-calendar-agent/google"
	google_mocks "github.com/inference-gateway/google-calendar-agent/google/mocks"
	"github.com/inference-gateway/google-calendar-agent/utils"
)

var (
	logger          *zap.Logger
	calendarService google.CalendarService
)

func main() {
	var err error

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.Encoding = "json"
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	logger, err = config.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer func() {
		_ = logger.Sync()
	}()

	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		credentialsPath = "credentials.json"
		logger.Debug("credentials path not specified in environment, using default",
			zap.String("credentialsPath", credentialsPath))
	} else {
		logger.Debug("using credentials path from environment",
			zap.String("credentialsPath", credentialsPath))
	}

	err = utils.CreateGoogleCredentialsFile(logger)
	if err != nil {
		logger.Fatal("failed to create google credentials file",
			zap.String("credentialsPath", credentialsPath),
			zap.Error(err))
	}

	logger.Info("starting google-calendar-agent")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
		logger.Debug("port not specified in environment, using default", zap.String("port", port))
	} else {
		logger.Debug("using port from environment", zap.String("port", port))
	}

	ctx := context.Background()
	logger.Info("initializing calendar service", zap.String("credentialsPath", credentialsPath))

	googleService, err := google.NewCalendarService(ctx, logger, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		logger.Warn("failed to initialize calendar service, running in demo mode",
			zap.Error(err),
			zap.String("credentialsPath", credentialsPath))
		calendarService = &google_mocks.FakeCalendarService{}
		logger.Warn("using mock calendar service - no real google calendar api calls will be made",
			zap.String("serviceType", "mock"))
	} else {
		calendarService = googleService
		logger.Info("calendar service initialized successfully",
			zap.String("serviceType", "google-api"))
	}

	agent := a2a.NewCalendarAgent(calendarService, logger)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		logger.Debug("health check requested",
			zap.String("clientIP", c.ClientIP()),
			zap.String("userAgent", c.GetHeader("User-Agent")))
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	r.POST("/a2a", func(c *gin.Context) {
		agent.HandleA2ARequest(c)
	})

	r.GET("/.well-known/agent.json", func(c *gin.Context) {
		logger.Info("agent info requested",
			zap.String("clientIP", c.ClientIP()),
			zap.String("userAgent", c.GetHeader("User-Agent")))
		info := a2a.AgentCard{
			Name:        "google-calendar-agent",
			Description: "A comprehensive Google Calendar agent that can list, create, update, and delete calendar events using the A2A protocol",
			URL:         "http://localhost:8084",
			Version:     "1.0.0",
			Capabilities: a2a.AgentCapabilities{
				Streaming:              false,
				PushNotifications:      false,
				StateTransitionHistory: false,
			},
			DefaultInputModes:  []string{"text/plain"},
			DefaultOutputModes: []string{"text/plain", "application/json"},
			Skills: []a2a.AgentSkill{
				{
					ID:          "list-calendars",
					Name:        "List Available Calendars",
					Description: "Discover and list all available Google Calendars with their IDs",
					InputModes:  []string{"text/plain"},
					OutputModes: []string{"text/plain", "application/json"},
					Examples:    []string{"List my calendars", "Show available calendars", "What calendars do I have?", "Find my calendar ID"},
				},
				{
					ID:          "list-events",
					Name:        "List Calendar Events",
					Description: "List upcoming events from your Google Calendar",
					InputModes:  []string{"text/plain"},
					OutputModes: []string{"text/plain", "application/json"},
					Examples:    []string{"Show me my events today", "What's on my calendar this week?", "List my meetings tomorrow"},
				},
				{
					ID:          "create-event",
					Name:        "Create Calendar Event",
					Description: "Create a new event in your Google Calendar",
					InputModes:  []string{"text/plain"},
					OutputModes: []string{"text/plain", "application/json"},
					Examples:    []string{"Schedule a meeting with John at 2pm tomorrow", "Create a dentist appointment on Friday at 10am", "Book lunch with Sarah next Tuesday at 12:30pm"},
				},
				{
					ID:          "update-event",
					Name:        "Update Calendar Event",
					Description: "Modify existing events in your Google Calendar",
					InputModes:  []string{"text/plain"},
					OutputModes: []string{"text/plain", "application/json"},
					Examples:    []string{"Move my 2pm meeting to 3pm", "Change the title of my appointment", "Update my meeting location"},
				},
				{
					ID:          "delete-event",
					Name:        "Delete Calendar Event",
					Description: "Remove events from your Google Calendar",
					InputModes:  []string{"text/plain"},
					OutputModes: []string{"text/plain", "application/json"},
					Examples:    []string{"Cancel my 2pm meeting", "Delete my dentist appointment", "Remove the lunch with Sarah"},
				},
			},
		}
		c.JSON(http.StatusOK, info)
	})

	logger.Info("server starting", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
