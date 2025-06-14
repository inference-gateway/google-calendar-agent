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
	"github.com/inference-gateway/google-calendar-agent/agent"
	"github.com/inference-gateway/google-calendar-agent/config"
	"github.com/inference-gateway/google-calendar-agent/internal/logging"
	"go.uber.org/zap"
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
			// Ignore sync errors on stderr/stdout
			_ = err
		}
	}()

	logger.Info("Configuration loaded successfully",
		zap.String("environment", cfg.App.Environment),
		zap.Bool("debug", cfg.IsDebugEnabled()),
		zap.Bool("demo_mode", cfg.App.DemoMode),
		zap.String("log_level", cfg.Logging.Level))

	agentService, err := agent.NewGoogleCalendarAgent(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create Google Calendar agent service", zap.Error(err))
	}

	toolBox := server.NewDefaultToolBox()
	agentService.RegisterTools(toolBox)

	serverCfg := serverconfig.Config{
		AgentName:        "Google Calendar Agent",
		AgentDescription: "AI agent for Google Calendar operations including listing events, creating events, managing schedules, and finding available time slots",
		Port:             cfg.Server.Port,
	}

	var a2aServer server.A2AServer
	if cfg.LLM.Enabled && cfg.LLM.GatewayURL != "" {
		llmConfig := &serverconfig.LLMProviderClientConfig{
			Provider: cfg.LLM.Provider,
			Model:    cfg.LLM.Model,
			APIKey:   "",
		}

		a2aServer = server.SimpleA2AServerWithAgent(serverCfg, logger, llmConfig, toolBox)
		logger.Info("✅ Google Calendar agent created with AI capabilities",
			zap.String("provider", cfg.LLM.Provider),
			zap.String("model", cfg.LLM.Model))
	} else {
		agentInstance := server.NewDefaultOpenAICompatibleAgent(logger)
		agentInstance.SetSystemPrompt("You are a helpful Google Calendar assistant. You can help users manage their calendar events, create new events, find available time slots, and answer questions about their schedule. Use the available tools to interact with Google Calendar.")
		agentInstance.SetToolBox(toolBox)

		a2aServer = server.NewA2AServerBuilder(serverCfg, logger).
			WithAIPoweredAgent(agentInstance).
			Build()
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

	if err := agentService.Close(shutdownCtx); err != nil {
		logger.Error("Error during agent service shutdown", zap.Error(err))
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
        "content": "List my calendar events for today"
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
