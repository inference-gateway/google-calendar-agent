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
	agent "github.com/inference-gateway/google-calendar-agent/agent"
	config "github.com/inference-gateway/google-calendar-agent/config"
	logging "github.com/inference-gateway/google-calendar-agent/internal/logging"
	zap "go.uber.org/zap"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	fmt.Printf("🗓️  Starting Google Calendar A2A Agent v%s (commit: %s, built: %s)\n", version, commit, date)

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

	logger.Info("Configuration loaded successfully",
		zap.String("environment", cfg.App.Environment),
		zap.Bool("debug", cfg.IsDebugEnabled()),
		zap.Bool("demo_mode", cfg.App.DemoMode),
		zap.String("log_level", cfg.Logging.Level))

	toolBox := server.NewDefaultToolBox()

	calendarTools, err := agent.NewGoogleCalendarTools(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create Google Calendar tools", zap.Error(err))
	}

	calendarTools.RegisterTools(toolBox)

	serverCfg := serverconfig.Config{
		AgentName:        "Google Calendar Agent",
		AgentDescription: "AI agent for Google Calendar operations including listing events, creating events, managing schedules, and finding available time slots",
		Port:             cfg.Server.Port,
		QueueConfig: &serverconfig.QueueConfig{
			CleanupInterval: time.Minute * 5,
		},
	}

	if cfg.LLM.Enabled && cfg.LLM.GatewayURL != "" {
		serverCfg.AgentConfig = &serverconfig.AgentConfig{
			BaseURL:     cfg.LLM.GatewayURL,
			Provider:    cfg.LLM.Provider,
			APIKey:      "",
			Model:       cfg.LLM.Model,
			Temperature: cfg.LLM.Temperature,
			MaxTokens:   cfg.LLM.MaxTokens,
			CustomHeaders: map[string]string{
				"X-A2A-Internal": "true",
			},
		}
		logger.Info("Configuring agent with LLM client",
			zap.String("base_url", cfg.LLM.GatewayURL),
			zap.String("provider", cfg.LLM.Provider),
			zap.String("model", cfg.LLM.Model))
	}
	if cfg.IsDebugEnabled() {
		serverCfg.Debug = true
	}

	agentInstance, err := server.NewOpenAICompatibleAgentWithConfig(logger, serverCfg.AgentConfig)
	if err != nil {
		logger.Fatal("Failed to create OpenAI-compatible agent", zap.Error(err))
	}

	currentTime := time.Now().Format("Monday, January 2, 2006 at 15:04 MST")
	systemPrompt := fmt.Sprintf(`Today is %s. You are a helpful Google Calendar assistant.

ALWAYS use the available tools to interact with Google Calendar - never provide generic responses without using tools.

Tool Usage:
- For listing events: use list_calendar_events
- For creating events: use create_calendar_event  
- For finding free time: use find_available_time

IMPORTANT: After using any tool, you MUST provide a clear, helpful response to the user based on the tool results. Never leave your response empty.`, currentTime)
	agentInstance.SetSystemPrompt(systemPrompt)
	agentInstance.SetToolBox(toolBox)

	a2aServer := server.NewA2AServerBuilder(serverCfg, logger).
		WithAgent(agentInstance).
		Build()

	if cfg.LLM.Enabled && cfg.LLM.GatewayURL != "" {
		logger.Info("✅ Google Calendar agent created with AI capabilities",
			zap.String("provider", cfg.LLM.Provider),
			zap.String("model", cfg.LLM.Model),
			zap.String("gateway_url", cfg.LLM.GatewayURL))
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
		fmt.Println("\n⚠️  Running in DEMO MODE - using mock services")
	} else if cfg.Google.ServiceAccountJSON == "" && cfg.Google.CredentialsPath == "" {
		fmt.Println("\n⚠️  Google credentials not configured - some features may be limited")
		fmt.Println("   Set GOOGLE_CALENDAR_SA_JSON or GOOGLE_APPLICATION_CREDENTIALS")
	}

	if !cfg.LLM.Enabled {
		fmt.Println("\n💡 LLM disabled - agent will have limited AI capabilities")
		fmt.Println("   Set LLM_ENABLED=true and configure LLM settings for full AI features")
	}
}
