package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	server "github.com/inference-gateway/adk/server"
	zap "go.uber.org/zap"

	config "github.com/inference-gateway/google-calendar-agent/config"
	logging "github.com/inference-gateway/google-calendar-agent/internal/logging"
	toolbox "github.com/inference-gateway/google-calendar-agent/toolbox"
)

var (
	Version          = "unknown"
	Commit           = "unknown"
	Date             = "unknown"
	AgentName        = "Google Calendar Agent"
	AgentDescription = "AI agent for Google Calendar operations including listing events, creating events, managing schedules, and finding available time slots."
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	cfg.A2A.AgentName = AgentName
	cfg.A2A.AgentDescription = AgentDescription
	cfg.A2A.AgentVersion = Version

	logger, err := logging.NewLogger(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	logger.Info("Starting Google Calendar A2A Agent",
		zap.String("version", Version),
		zap.String("commit", Commit),
		zap.String("date", Date),
		zap.String("environment", cfg.Environment),
		zap.Bool("demo_mode", cfg.DemoMode),
		zap.Bool("debug", cfg.IsDebugEnabled()),
	)

	toolBox := server.NewDefaultToolBox()
	calendarTools, err := toolbox.NewGoogleCalendarTools(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create Google Calendar tools", zap.Error(err))
	}
	calendarTools.RegisterTools(toolBox)

	serverCfg := cfg.A2A
	if cfg.IsDebugEnabled() {
		serverCfg.Debug = true
	}

	if serverCfg.AgentURL == "" {
		logger.Fatal("Agent URL is not configured. Please set A2A_AGENT_URL.")
	}

	var a2aServer server.A2AServer

	if cfg.DemoMode {
		logger.Info("✅ Google Calendar agent created in demo mode (AI disabled)")
		demoHandler := toolbox.NewDemoTaskHandler(toolBox, logger)
		a2aServer, err = server.NewA2AServerBuilder(serverCfg, logger).
			WithBackgroundTaskHandler(demoHandler).
			WithDefaultStreamingTaskHandler().
			WithAgentCardFromFile(".well-known/agent.json", map[string]interface{}{
				"name":        AgentName,
				"description": AgentDescription,
				"version":     Version,
				"url":         serverCfg.AgentURL,
			}).
			Build()
		if err != nil {
			logger.Fatal("Failed to create demo server", zap.Error(err))
		}
	} else {
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

IMPORTANT: Before creating any event, MUST check for conflicts first. Always provide clear responses based on tool results.`,
			time.Now().Format("Monday, January 2, 2006 at 15:04 MST"))

		agentInstance, err := server.NewAgentBuilder(logger).
			WithConfig(&serverCfg.AgentConfig).
			WithSystemPrompt(systemPrompt).
			WithToolBox(toolBox).
			WithMaxConversationHistory(20).
			WithMaxChatCompletion(10).
			Build()
		if err != nil {
			logger.Fatal("Failed to create agent", zap.Error(err))
		}

		logger.Info("✅ Google Calendar agent created with AI capabilities")

		a2aServer, err = server.NewA2AServerBuilder(serverCfg, logger).
			WithAgent(agentInstance).
            WithDefaultBackgroundTaskHandler().
			WithDefaultStreamingTaskHandler().
			WithAgentCardFromFile(".well-known/agent.json", map[string]interface{}{
				"name":        AgentName,
				"description": AgentDescription,
				"version":     Version,
				"url":         serverCfg.AgentURL,
			}).
			Build()
		if err != nil {
			logger.Fatal("Failed to create agent server", zap.Error(err))
		}
	}

	ctx, cancel := context.WithCancel(ctx)
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
