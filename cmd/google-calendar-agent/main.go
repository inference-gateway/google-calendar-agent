package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/option"

	"github.com/inference-gateway/google-calendar-agent/a2a"
	"github.com/inference-gateway/google-calendar-agent/google"
	google_mocks "github.com/inference-gateway/google-calendar-agent/google/mocks"
)

var (
	logger          *zap.Logger
	calendarService google.CalendarService

	// Version information - will be set by build flags
	version = "dev"
	commit  = "unknown"
	date    = "unknown"

	// Command line flags
	showVersion     = flag.Bool("version", false, "show version information and exit")
	showHelp        = flag.Bool("help", false, "show help information and exit")
	credentialsPath = flag.String("credentials", "", "path to Google credentials file (overrides GOOGLE_APPLICATION_CREDENTIALS env var)")
	logLevel        = flag.String("log-level", "debug", "log level (debug, info, warn, error)")
	calendarID      = flag.String("calendar-id", "", "Google calendar ID to use (overrides GOOGLE_CALENDAR_ID env var)")
	demoMode        = flag.Bool("demo", false, "run in demo mode with mock calendar service")
	ginMode         = flag.String("gin-mode", "", "Gin server mode (debug, release, test) - overrides GIN_MODE env var")
	enableTLS       = flag.Bool("enable-tls", false, "enable TLS/HTTPS server (overrides ENABLE_TLS env var)")
	tlsCertPath     = flag.String("tls-cert", "", "path to TLS certificate file (overrides TLS_CERT_PATH env var)")
	tlsKeyPath      = flag.String("tls-key", "", "path to TLS private key file (overrides TLS_KEY_PATH env var)")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("google-calendar-agent\n")
		fmt.Printf("  Version:    %s\n", version)
		fmt.Printf("  Commit:     %s\n", commit)
		fmt.Printf("  Build Date: %s\n", date)
		os.Exit(0)
	}

	if *showHelp {
		fmt.Printf("google-calendar-agent - A comprehensive Google Calendar agent using the A2A protocol\n\n")
		fmt.Printf("Usage:\n")
		fmt.Printf("  -calendar-id string           Google calendar ID to use (overrides GOOGLE_CALENDAR_ID env var)\n")
		fmt.Printf("  -credentials string           The path to Google credentials file (overrides GOOGLE_APPLICATION_CREDENTIALS env var)\n")
		fmt.Printf("  -demo                         Run in demo mode with mock calendar service\n")
		fmt.Printf("  -enable-tls                   Enable TLS/HTTPS server (overrides ENABLE_TLS env var)\n")
		fmt.Printf("  -gin-mode string              Gin server mode (debug, release, test) - overrides GIN_MODE env var\n")
		fmt.Printf("  -help                         Show help information and exit\n")
		fmt.Printf("  -log-level string             Log level (debug, info, warn, error) (default \"debug\")\n")
		fmt.Printf("  -tls-cert string              Path to TLS certificate file (overrides TLS_CERT_PATH env var)\n")
		fmt.Printf("  -tls-key string               Path to TLS private key file (overrides TLS_KEY_PATH env var)\n")
		fmt.Printf("  -version                      Show version information and exit\n")
		fmt.Printf("\nEnvironment Variables:\n")
		fmt.Printf("  GOOGLE_CALENDAR_SA_JSON       - The Google Service Account in a JSON format\n")
		fmt.Printf("  GOOGLE_CALENDAR_ID            - Google calendar ID to use (default: primary)\n")
		fmt.Printf("  GIN_MODE                      - Gin server mode (debug, release, test) - overridden by -gin-mode flag\n")
		fmt.Printf("  LOG_LEVEL                     - Log level (debug, info, warn, error) - overridden by -log-level flag\n")
		fmt.Printf("  ENABLE_TLS                    - Enable TLS/HTTPS server (true/false) - overridden by -enable-tls flag\n")
		fmt.Printf("  TLS_CERT_PATH                 - Path to TLS certificate file - overridden by -tls-cert flag\n")
		fmt.Printf("  TLS_KEY_PATH                  - Path to TLS private key file - overridden by -tls-key flag\n")
		os.Exit(0)
	}

	var err error

	finalLogLevel := *logLevel
	if finalLogLevel == "debug" {
		if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
			finalLogLevel = envLogLevel
		}
	}

	var logLevelZap zapcore.Level
	switch finalLogLevel {
	case "debug":
		logLevelZap = zap.DebugLevel
	case "info":
		logLevelZap = zap.InfoLevel
	case "warn":
		logLevelZap = zap.WarnLevel
	case "error":
		logLevelZap = zap.ErrorLevel
	default:
		fmt.Fprintf(os.Stderr, "Invalid log level: %s. Using debug.\n", finalLogLevel)
		logLevelZap = zap.DebugLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(logLevelZap)
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

	logger.Info("starting google-calendar-agent",
		zap.String("version", version),
		zap.String("commit", commit),
		zap.String("buildDate", date))

	finalGinMode := *ginMode
	if finalGinMode == "" {
		if envGinMode := os.Getenv("GIN_MODE"); envGinMode != "" {
			finalGinMode = envGinMode
		} else {
			finalGinMode = "release"
		}
	}

	switch finalGinMode {
	case "debug", "release", "test":
		gin.SetMode(finalGinMode)
		logger.Debug("gin mode configured", zap.String("mode", finalGinMode))
	default:
		logger.Warn("invalid gin mode, using debug mode", zap.String("invalidMode", finalGinMode))
		gin.SetMode(gin.DebugMode)
	}

	finalCredentialsPath := *credentialsPath
	if finalCredentialsPath == "" {
		finalCredentialsPath = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if finalCredentialsPath == "" {
			finalCredentialsPath = "credentials.json"
			logger.Debug("credentials path not specified, using default",
				zap.String("credentialsPath", finalCredentialsPath))
		} else {
			logger.Debug("using credentials path from environment",
				zap.String("credentialsPath", finalCredentialsPath))
		}
	} else {
		logger.Debug("using credentials path from flag",
			zap.String("credentialsPath", finalCredentialsPath))
	}

	err = google.CreateCredentialsFile(logger)
	if err != nil {
		logger.Fatal("failed to create google credentials file",
			zap.String("credentialsPath", finalCredentialsPath),
			zap.Error(err))
	}

	finalEnableTLS := *enableTLS
	if !finalEnableTLS {
		if envEnableTLS := os.Getenv("ENABLE_TLS"); envEnableTLS == "true" {
			finalEnableTLS = true
		}
	}

	finalPort := "8080"
	if finalEnableTLS {
		finalPort = "8443"
	}
	logger.Debug("using port", zap.String("port", finalPort), zap.Bool("tls", finalEnableTLS))

	var finalTLSCertPath, finalTLSKeyPath string
	if finalEnableTLS {
		finalTLSCertPath = *tlsCertPath
		if finalTLSCertPath == "" {
			finalTLSCertPath = os.Getenv("TLS_CERT_PATH")
		}

		finalTLSKeyPath = *tlsKeyPath
		if finalTLSKeyPath == "" {
			finalTLSKeyPath = os.Getenv("TLS_KEY_PATH")
		}

		if finalTLSCertPath == "" || finalTLSKeyPath == "" {
			logger.Fatal("TLS enabled but certificate or key path not provided",
				zap.Bool("enableTLS", finalEnableTLS),
				zap.String("certPath", finalTLSCertPath),
				zap.String("keyPath", finalTLSKeyPath))
		}

		logger.Info("TLS enabled",
			zap.String("certPath", finalTLSCertPath),
			zap.String("keyPath", finalTLSKeyPath))
	} else {
		logger.Debug("TLS disabled, running HTTP server")
	}

	finalCalendarID := *calendarID
	if finalCalendarID == "" {
		finalCalendarID = os.Getenv("GOOGLE_CALENDAR_ID")
		if finalCalendarID == "" {
			finalCalendarID = "primary"
		}
	}
	logger.Debug("using calendar ID", zap.String("calendarID", finalCalendarID))

	ctx := context.Background()
	logger.Info("initializing calendar service", zap.String("credentialsPath", finalCredentialsPath))

	if *demoMode {
		logger.Info("demo mode enabled, using mock calendar service")
		calendarService = &google_mocks.FakeCalendarService{}
	} else {
		googleService, err := google.NewCalendarService(ctx, logger, option.WithCredentialsFile(finalCredentialsPath))
		if err != nil {
			logger.Warn("failed to initialize calendar service, running in demo mode",
				zap.Error(err),
				zap.String("credentialsPath", finalCredentialsPath))
			calendarService = &google_mocks.FakeCalendarService{}
			logger.Warn("using mock calendar service - no real google calendar api calls will be made",
				zap.String("serviceType", "mock"))
		} else {
			calendarService = googleService
			logger.Info("calendar service initialized successfully",
				zap.String("serviceType", "google-api"))
		}
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

		protocol := "http"
		if finalEnableTLS {
			protocol = "https"
		}

		info := a2a.AgentCard{
			Name:        "google-calendar-agent",
			Description: "A comprehensive Google Calendar agent that can list, create, update, and delete calendar events using the A2A protocol",
			URL:         fmt.Sprintf("%s://localhost:%s", protocol, finalPort),
			Version:     version,
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

	if finalEnableTLS {
		logger.Info("starting HTTPS server",
			zap.String("port", finalPort),
			zap.String("certPath", finalTLSCertPath),
			zap.String("keyPath", finalTLSKeyPath))
		if err := r.RunTLS(":"+finalPort, finalTLSCertPath, finalTLSKeyPath); err != nil {
			logger.Fatal("failed to start HTTPS server", zap.Error(err))
		}
	} else {
		logger.Info("starting HTTP server", zap.String("port", finalPort))
		if err := r.Run(":" + finalPort); err != nil {
			logger.Fatal("failed to start HTTP server", zap.Error(err))
		}
	}
}
