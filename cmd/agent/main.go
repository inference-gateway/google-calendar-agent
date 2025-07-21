package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	adk "github.com/inference-gateway/a2a/adk"
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

	cfg, logger := mustInitialize(ctx)
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	logStartup(cfg, logger)

	a2aServer, err := createServer(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	runServer(ctx, a2aServer, logger)
}

func mustInitialize(ctx context.Context) (*config.Config, *zap.Logger) {
	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger, err := logging.NewLogger(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	return cfg, logger
}

func logStartup(cfg *config.Config, logger *zap.Logger) {
	logger.Info("Starting Google Calendar A2A Agent",
		zap.String("version", server.BuildAgentVersion),
		zap.String("commit", commit),
		zap.String("date", date),
		zap.String("environment", cfg.Environment),
		zap.Bool("demo_mode", cfg.DemoMode),
		zap.Bool("debug", cfg.IsDebugEnabled()),
	)
}

func createServer(cfg *config.Config, logger *zap.Logger) (server.A2AServer, error) {
	toolBox := server.NewDefaultToolBox()
	calendarTools, err := toolbox.NewGoogleCalendarTools(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create Google Calendar tools", zap.Error(err))
	}
	calendarTools.RegisterTools(toolBox)

	serverCfg := cfg.ADK
	if cfg.IsDebugEnabled() {
		serverCfg.Debug = true
	}

	agentCard := a2a.GetAgentCard(serverCfg)

	if cfg.DemoMode {
		return createDemoServer(serverCfg, toolBox, agentCard, logger)
	}
	return createAgentServer(serverCfg, toolBox, agentCard, logger)
}

func createDemoServer(serverCfg serverconfig.Config, toolBox *server.DefaultToolBox, agentCard adk.AgentCard, logger *zap.Logger) (server.A2AServer, error) {
	logger.Info("✅ Google Calendar agent created in demo mode (AI disabled)")

	demoHandler := toolbox.NewDemoTaskHandler(toolBox, logger)
	return server.NewA2AServerBuilder(serverCfg, logger).
		WithTaskHandler(demoHandler).
		WithAgentCard(agentCard).
		Build()
}

func createAgentServer(serverCfg serverconfig.Config, toolBox *server.DefaultToolBox, agentCard adk.AgentCard, logger *zap.Logger) (server.A2AServer, error) {
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

	return server.NewA2AServerBuilder(serverCfg, logger).
		WithAgent(agentInstance).
		WithAgentCard(agentCard).
		Build()
}

func runServer(ctx context.Context, a2aServer server.A2AServer, logger *zap.Logger) {
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
