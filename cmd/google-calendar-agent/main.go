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
	"github.com/inference-gateway/google-calendar-agent/config"
	"github.com/inference-gateway/google-calendar-agent/google"
	google_mocks "github.com/inference-gateway/google-calendar-agent/google/mocks"
	"github.com/inference-gateway/google-calendar-agent/llm"
	llm_mocks "github.com/inference-gateway/google-calendar-agent/llm/mocks"
)

var (
	logger          *zap.Logger
	calendarService google.CalendarService
	llmService      llm.Service

	// Version information - will be set by build flags
	version = "dev"
	commit  = "unknown"
	date    = "unknown"

	// Command line flags
	showVersion = flag.Bool("version", false, "show version information and exit")
	showHelp    = flag.Bool("help", false, "show help information and exit")
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
		fmt.Printf("google-calendar-agent - A Google Calendar agent using the A2A protocol\n\n")
		fmt.Printf("Usage:\n")
		fmt.Printf("  -help                         Show help information and exit\n")
		fmt.Printf("  -version                      Show version information and exit\n")
		fmt.Printf("\nConfiguration is managed through environment variables and config files.\n")
		fmt.Printf("See the project documentation for configuration details.\n")
		os.Exit(0)
	}

	ctx := context.Background()
	cfg, err := config.Load(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	logLevelStr := cfg.GetLogLevel()

	var logLevel zapcore.Level
	switch logLevelStr {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "warn":
		logLevel = zap.WarnLevel
	case "error":
		logLevel = zap.ErrorLevel
	default:
		logLevel = zap.InfoLevel
	}

	logConfig := zap.NewProductionConfig()
	logConfig.Level = zap.NewAtomicLevelAt(logLevel)
	logConfig.OutputPaths = []string{"stdout"}
	logConfig.ErrorOutputPaths = []string{"stderr"}
	logConfig.Encoding = "json"
	logConfig.EncoderConfig.TimeKey = "timestamp"
	logConfig.EncoderConfig.LevelKey = "level"
	logConfig.EncoderConfig.MessageKey = "message"
	logConfig.EncoderConfig.CallerKey = "caller"
	logConfig.EncoderConfig.StacktraceKey = "stacktrace"
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logConfig.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	logConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	logger, err = logConfig.Build()
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

	gin.SetMode(cfg.Server.Mode)
	logger.Info("gin mode configured", zap.String("mode", cfg.Server.Mode))

	err = google.CreateCredentialsFile(logger, cfg)
	if err != nil {
		logger.Fatal("failed to create google credentials file", zap.Error(err))
	}

	_, err = cfg.GetTLSConfig()
	if err != nil {
		logger.Fatal("failed to get TLS config", zap.Error(err))
	}

	port := cfg.GetPort()
	logger.Debug("using port", zap.String("port", port), zap.Bool("tls", cfg.Server.EnableTLS))

	if cfg.Server.EnableTLS {
		if cfg.TLS.CertPath == "" || cfg.TLS.KeyPath == "" {
			logger.Fatal("TLS enabled but certificate or key path not provided",
				zap.Bool("enableTLS", cfg.Server.EnableTLS),
				zap.String("certPath", cfg.TLS.CertPath),
				zap.String("keyPath", cfg.TLS.KeyPath))
		}

		logger.Info("TLS enabled",
			zap.String("certPath", cfg.TLS.CertPath),
			zap.String("keyPath", cfg.TLS.KeyPath))
	} else {
		logger.Debug("TLS disabled, running HTTP server")
	}

	calendarID := cfg.Google.CalendarID
	if calendarID == "" {
		calendarID = "primary"
	}
	logger.Debug("using calendar ID", zap.String("calendarID", calendarID))

	logger.Info("initializing calendar service")

	if cfg.ShouldUseMockService() {
		logger.Info("demo mode enabled, using mock calendar service")
		calendarService = &google_mocks.FakeCalendarService{}
	} else {
		credType, credValue, optErr := cfg.GetGoogleCredentialsOption()
		if optErr != nil {
			logger.Warn("failed to get google credentials option, running in demo mode",
				zap.Error(optErr))
			calendarService = &google_mocks.FakeCalendarService{}
		} else {
			var googleService google.CalendarService
			var err error

			switch credType {
			case "json":
				googleService, err = google.NewCalendarService(ctx, cfg, logger, option.WithCredentialsJSON([]byte(credValue)))
			case "file":
				googleService, err = google.NewCalendarService(ctx, cfg, logger, option.WithCredentialsFile(credValue))
			default:
				logger.Warn("no credentials available, running in demo mode")
				calendarService = &google_mocks.FakeCalendarService{}
				googleService = nil
			}

			if err != nil {
				logger.Warn("failed to initialize calendar service, running in demo mode",
					zap.Error(err))
				calendarService = &google_mocks.FakeCalendarService{}
				logger.Warn("using mock calendar service - no real google calendar api calls will be made",
					zap.String("serviceType", "mock"))
			} else if googleService != nil {
				calendarService = googleService
				logger.Info("calendar service initialized successfully",
					zap.String("serviceType", "google-api"))
			}
		}
	}

	// Initialize LLM service
	logger.Info("initializing LLM service")
	llmService, err = llm.NewInferenceGatewayService(cfg, logger)
	if err != nil {
		logger.Warn("failed to initialize LLM service, using disabled mock",
			zap.Error(err))
		// Create a disabled mock service
		mockService := &llm_mocks.FakeService{}
		mockService.IsEnabledReturns(false)
		mockService.GetProviderReturns("")
		mockService.GetModelReturns("")
		llmService = mockService
	} else if llmService.IsEnabled() {
		logger.Info("LLM service initialized successfully",
			zap.String("provider", llmService.GetProvider()),
			zap.String("model", llmService.GetModel()))
	} else {
		logger.Info("LLM service is disabled")
	}

	agent := a2a.NewCalendarAgentWithLLM(calendarService, logger, cfg, llmService)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/a2a" && c.Request.Method != "POST" {
			logger.Debug("unsupported method on /a2a endpoint",
				zap.String("method", c.Request.Method),
				zap.String("clientIP", c.ClientIP()),
				zap.String("userAgent", c.GetHeader("User-Agent")))
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"error":           "Method Not Allowed",
				"message":         "Only POST requests are supported on this endpoint",
				"allowed_methods": []string{"POST"},
				"endpoint":        "/a2a",
			})
			c.Abort()
			return
		}
		c.Next()
	})

	r.NoRoute(func(c *gin.Context) {
		logger.Debug("route not found",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("clientIP", c.ClientIP()),
			zap.String("userAgent", c.GetHeader("User-Agent")))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "The requested endpoint does not exist",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
			"available_endpoints": []string{
				"GET /health",
				"POST /a2a",
				"GET /.well-known/agent.json",
			},
		})
	})

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

		baseURL := cfg.GetBaseURL()

		info := a2a.AgentCard{
			Name:        "google-calendar-agent",
			Description: "A Google Calendar agent that can list, create, update, and delete calendar events using the A2A protocol",
			URL:         baseURL,
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

	if cfg.Server.EnableTLS {
		logger.Info("starting HTTPS server",
			zap.String("port", port),
			zap.String("certPath", cfg.TLS.CertPath),
			zap.String("keyPath", cfg.TLS.KeyPath))
		if err := r.RunTLS(":"+port, cfg.TLS.CertPath, cfg.TLS.KeyPath); err != nil {
			logger.Fatal("failed to start HTTPS server", zap.Error(err))
		}
	} else {
		logger.Info("starting HTTP server", zap.String("port", port))
		if err := r.Run(":" + port); err != nil {
			logger.Fatal("failed to start HTTP server", zap.Error(err))
		}
	}
}
