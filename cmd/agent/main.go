package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	server "github.com/inference-gateway/a2a/adk/server"
	serverconfig "github.com/inference-gateway/a2a/adk/server/config"
	zap "go.uber.org/zap"

	a2a "github.com/inference-gateway/google-calendar-agent/a2a"
	config "github.com/inference-gateway/google-calendar-agent/config"
	logging "github.com/inference-gateway/google-calendar-agent/internal/logging"
	toolbox "github.com/inference-gateway/google-calendar-agent/toolbox"
)

var (
	commit = "unknown"
	date   = "unknown"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger, err := logging.NewLogger(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			_ = err
		}
	}()

	logger.Info("Starting Google Calendar A2A Agent",
		zap.String("version", server.BuildAgentVersion),
		zap.String("commit", commit),
		zap.String("date", date),
		zap.String("environment", cfg.App.Environment),
		zap.Bool("demo_mode", cfg.App.DemoMode),
		zap.String("agent_url", cfg.App.AgentURL),
		zap.String("port", cfg.Server.Port),
		zap.Bool("debug", cfg.IsDebugEnabled()),
		zap.String("log_level", cfg.Logging.Level),
	)

	toolBox := server.NewDefaultToolBox()

	calendarTools, err := toolbox.NewGoogleCalendarTools(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create Google Calendar tools", zap.Error(err))
	}

	calendarTools.RegisterTools(toolBox)

	serverCfg := serverconfig.Config{
		AgentURL: cfg.App.AgentURL,
		Port:     cfg.Server.Port,
		QueueConfig: serverconfig.QueueConfig{
			CleanupInterval: time.Minute * 5,
		},
	}

	if cfg.LLM.Enabled && cfg.LLM.GatewayURL != "" && !cfg.App.DemoMode {
		serverCfg.AgentConfig = serverconfig.AgentConfig{
			BaseURL:     cfg.LLM.GatewayURL,
			Provider:    cfg.LLM.Provider,
			APIKey:      "",
			Model:       cfg.LLM.Model,
			Temperature: cfg.LLM.Temperature,
			MaxTokens:   cfg.LLM.MaxTokens,
			CustomHeaders: map[string]string{
				"X-A2A-Bypass": "true",
			},
			MaxChatCompletionIterations: 20,
			MaxConversationHistory:      20,
			MaxRetries:                  10,
			Timeout:                     cfg.LLM.Timeout,
		}
		logger.Info("Configuring agent with LLM client",
			zap.String("base_url", cfg.LLM.GatewayURL),
			zap.String("provider", cfg.LLM.Provider),
			zap.String("model", cfg.LLM.Model),
			zap.Duration("timeout", cfg.LLM.Timeout))
	} else if cfg.App.DemoMode {
		serverCfg.AgentConfig = serverconfig.AgentConfig{
			Provider:                    "demo",
			Model:                       "demo-model",
			APIKey:                      "demo-key",
			Temperature:                 0.7,
			MaxTokens:                   4096,
			MaxChatCompletionIterations: 20,
			MaxConversationHistory:      20,
			MaxRetries:                  3,
			Timeout:                     time.Second * 30,
		}
		logger.Info("LLM configured in demo mode - agent will use mock responses")
	}

	if cfg.IsDebugEnabled() && !cfg.Server.DisableHealthLogs {
		serverCfg.Debug = true
	} else if cfg.IsDebugEnabled() && cfg.Server.DisableHealthLogs {
		logger.Debug("Debug mode enabled but health logs disabled - server debug mode remains off")
	}

	currentTime := time.Now().Format("Monday, January 2, 2006 at 15:04 MST")
	systemPrompt := fmt.Sprintf(`Today is %s. You are a Google Calendar assistant.

ALWAYS use tools - never provide responses without tool interactions.

Available tools:
- list_calendar_events - List events
- create_calendar_event - Create events (ALWAYS check conflicts first with check_conflicts)
- update_calendar_event - Update events
- delete_calendar_event - Delete events
- get_calendar_event - Get event details
- find_available_time - Find free time slots
- check_conflicts - Check scheduling conflicts

IMPORTANT: Before creating any event, MUST check for conflicts first. Always provide clear responses based on tool results.`, currentTime)

	agentCard := a2a.GetAgentCard(serverCfg)
	var a2aServer server.A2AServer
	if cfg.App.DemoMode {
		demoHandler := toolbox.NewDemoTaskHandler(toolBox, logger)

		a2aServer = server.NewA2AServerBuilder(serverCfg, logger).
			WithTaskHandler(demoHandler).
			WithAgentCard(agentCard).
			Build()
	} else {
		agentInstance, err := server.NewAgentBuilder(logger).
			WithConfig(&serverCfg.AgentConfig).
			WithSystemPrompt(systemPrompt).
			WithToolBox(toolBox).
			WithMaxConversationHistory(20).
			WithMaxChatCompletion(10).
			Build()
		if err != nil {
			logger.Fatal("Failed to create OpenAI-compatible agent", zap.Error(err))
		}

		a2aServer = server.NewA2AServerBuilder(serverCfg, logger).
			WithAgent(agentInstance).
			WithAgentCard(agentCard).
			Build()
	}

	logger.Info("Agent metadata",
		zap.String("agent_name", server.BuildAgentName),
		zap.String("agent_description", server.BuildAgentDescription),
		zap.String("agent_version", server.BuildAgentVersion))

	if cfg.LLM.Enabled && cfg.LLM.GatewayURL != "" && !cfg.App.DemoMode {
		logger.Info("✅ Google Calendar agent created with AI capabilities",
			zap.String("provider", cfg.LLM.Provider),
			zap.String("model", cfg.LLM.Model),
			zap.String("gateway_url", cfg.LLM.GatewayURL))
	} else if cfg.App.DemoMode {
		logger.Info("✅ Google Calendar agent created in demo mode (AI disabled)")
	} else {
		logger.Info("✅ Google Calendar agent created with default capabilities")
	}

	printStartupInfo(cfg, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := a2aServer.Start(ctx); err != nil {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := a2aServer.Stop(shutdownCtx); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}

	logger.Info("✅ Goodbye!")
}

func printStartupInfo(cfg *config.Config, logger *zap.Logger) {
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	fmt.Printf("\n🌐 Google Calendar agent running on port %s\n", port)
	fmt.Printf("\n🎯 Available endpoints:\n")
	fmt.Printf("📋 Agent info: http://localhost:%s/.well-known/agent.json\n", port)
	fmt.Printf("💚 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("📡 A2A endpoint: http://localhost:%s/a2a\n", port)

	fmt.Println("\n📝 Example A2A request:")
	fmt.Printf(`curl -X POST http://localhost:%s/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "message/send",
    "params": {
      "message": {
        "role": "user",
        "parts": [
          {
            "kind": "text",
            "content": "List my calendar events for today"
          }
        ]
      }
    },
    "id": 1
  }'`, port)
	fmt.Println()

	fmt.Println("\n📦 Google Calendar Tools Available:")
	fmt.Println("• list_calendar_events - List upcoming events")
	fmt.Println("• create_calendar_event - Create new events")
	fmt.Println("• update_calendar_event - Update existing events")
	fmt.Println("• delete_calendar_event - Delete events")
	fmt.Println("• get_calendar_event - Get event details")
	fmt.Println("• find_available_time - Find free time slots")
	fmt.Println("• check_conflicts - Check for scheduling conflicts")

	if cfg.App.DemoMode {
		fmt.Println("\n⚠️  Running in DEMO MODE - using mock services (AI disabled)")
	} else if cfg.Google.ServiceAccountJSON == "" && cfg.Google.CredentialsPath == "" {
		fmt.Println("\n⚠️  Google credentials not configured - some features may be limited")
		fmt.Println("   Set GOOGLE_CALENDAR_SA_JSON or GOOGLE_APPLICATION_CREDENTIALS")
	}

	if !cfg.LLM.Enabled {
		fmt.Println("\n💡 LLM disabled - agent will have limited AI capabilities")
		fmt.Println("   Set LLM_ENABLED=true and configure LLM settings for full AI features")
	} else if cfg.App.DemoMode {
		fmt.Println("\n💡 LLM disabled in demo mode - agent will use pattern matching only")
	}

	if cfg.Server.DisableHealthLogs {
		fmt.Println("\n🔇 Health check logging disabled (quiet mode)")
		fmt.Println("   Set SERVER_DISABLE_HEALTH_LOGS=false to enable health check logs")
	}
}
