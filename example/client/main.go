package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/inference-gateway/a2a/adk"
	"github.com/inference-gateway/a2a/adk/client"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
)

// Config represents the application configuration
type Config struct {
	ServerURL      string        `env:"A2A_SERVER_URL,default=http://localhost:8080"`
	PollInterval   time.Duration `env:"POLL_INTERVAL,default=1s"`
	MaxPollTimeout time.Duration `env:"MAX_POLL_TIMEOUT,default=60s"`
	LogLevel       string        `env:"LOG_LEVEL,default=info"`
	UseAsyncMode   bool          `env:"USE_ASYNC_MODE,default=true"`
}

type GoogleCalendarClient struct {
	client    client.A2AClient
	config    Config
	logger    *zap.Logger
	ctx       context.Context
	contextID string
}

func main() {
	ctx := context.Background()

	// Load configuration from environment variables
	var config Config
	if err := envconfig.Process(ctx, &config); err != nil {
		log.Fatalf("failed to process configuration: %v", err)
	}

	// Initialize logger based on log level
	var logger *zap.Logger
	var err error
	if config.LogLevel == "debug" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Create the Google Calendar client
	calendarClient, err := NewGoogleCalendarClient(ctx, config, logger)
	if err != nil {
		logger.Fatal("failed to create calendar client", zap.Error(err))
	}

	// Start interactive session
	calendarClient.StartInteractiveSession()
}

func NewGoogleCalendarClient(ctx context.Context, config Config, logger *zap.Logger) (*GoogleCalendarClient, error) {
	// Create A2A client
	a2aClient := client.NewClientWithLogger(config.ServerURL, logger)

	// Check agent capabilities
	logger.Info("connecting to Google Calendar Agent", zap.String("server_url", config.ServerURL))
	agentCard, err := a2aClient.GetAgentCard(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent card: %w", err)
	}

	logger.Info("connected to agent",
		zap.String("agent_name", agentCard.Name),
		zap.String("agent_version", agentCard.Version),
		zap.String("agent_description", agentCard.Description))

	return &GoogleCalendarClient{
		client: a2aClient,
		config: config,
		logger: logger,
		ctx:    ctx,
	}, nil
}

func (c *GoogleCalendarClient) StartInteractiveSession() {
	c.logger.Info("🗓️  Google Calendar Agent Client")
	c.logger.Info("Type your questions or commands. Type 'help' for examples, 'quit' to exit.")
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🗓️  Google Calendar Agent - Interactive Client")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Type your questions or commands about your Google Calendar.")
	fmt.Println("Examples:")
	fmt.Println("  • What meetings do I have today?")
	fmt.Println("  • Schedule a meeting with John tomorrow at 2 PM")
	fmt.Println("  • Show my calendar for next week")
	fmt.Println("  • Cancel my 3 PM meeting")
	fmt.Println("  • help - Show more examples")
	fmt.Println("  • status - Show session status")
	fmt.Println("  • debug - Show debug info and context ID")
	fmt.Println("  • reset - Start a new conversation")
	fmt.Println("  • quit - Exit the client")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("📅 You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch strings.ToLower(input) {
		case "quit", "exit", "q":
			fmt.Println("👋 Goodbye!")
			return
		case "help", "h":
			c.showHelp()
			continue
		case "clear":
			c.clearScreen()
			continue
		case "status", "s":
			c.showStatus()
			continue
		case "debug":
			fmt.Printf("Debug mode: %s\n", c.config.LogLevel)
			fmt.Printf("Async mode: %t\n", c.config.UseAsyncMode)
			fmt.Printf("Server URL: %s\n", c.config.ServerURL)
			if c.contextID != "" {
				fmt.Printf("Current Context ID: %s\n", c.contextID)
			} else {
				fmt.Printf("No active context\n")
			}
			continue
		case "reset", "new":
			if c.contextID != "" {
				c.logger.Info("🔄 resetting context", zap.String("old_context", c.contextID))
				fmt.Printf("Context reset. Starting new conversation.\n")
				c.contextID = ""
			} else {
				fmt.Printf("No active context to reset.\n")
			}
			continue
		}

		// Process the user's question/command
		c.processUserInput(input)
	}

	if err := scanner.Err(); err != nil {
		c.logger.Error("error reading input", zap.Error(err))
	}
}

func (c *GoogleCalendarClient) processUserInput(input string) {
	c.logger.Debug("processing user input", zap.String("input", input))

	// Create message ID
	messageID := fmt.Sprintf("msg-%d", time.Now().UnixNano())

	// Create the message
	message := adk.Message{
		Kind:      "message",
		MessageID: messageID,
		Role:      "user",
		Parts: []adk.Part{
			map[string]interface{}{
				"kind": "text",
				"text": input,
			},
		},
	}

	if c.contextID != "" {
		message.ContextID = &c.contextID
		c.logger.Info("🔗 using existing context",
			zap.String("context_id", c.contextID),
			zap.String("message_id", messageID))
	} else {
		c.logger.Info("🆕 starting new conversation - no context available",
			zap.String("message_id", messageID))
	}

	msgParams := adk.MessageSendParams{
		Message: message,
		Configuration: &adk.MessageSendConfiguration{
			Blocking:            boolPtr(!c.config.UseAsyncMode),
			AcceptedOutputModes: []string{"text"},
		},
	}

	fmt.Print("🤔 Thinking...")

	start := time.Now()

	if c.config.UseAsyncMode {
		c.handleAsyncResponse(msgParams)
	} else {
		c.handleSyncResponse(msgParams)
	}

	c.logger.Debug("request completed", zap.Duration("duration", time.Since(start)))
}

func (c *GoogleCalendarClient) handleSyncResponse(msgParams adk.MessageSendParams) {
	resp, err := c.client.SendTask(c.ctx, msgParams)
	if err != nil {
		fmt.Printf("\r❌ Error: %v\n", err)
		return
	}

	// Parse the response
	var task adk.Task
	if err := c.parseTaskFromResponse(resp.Result, &task); err != nil {
		fmt.Printf("\r❌ Error parsing response: %v\n", err)
		return
	}

	// Update context ID for conversation continuity
	if task.ContextID != "" {
		if c.contextID != task.ContextID {
			c.logger.Info("🔄 context updated",
				zap.String("old_context", c.contextID),
				zap.String("new_context", task.ContextID),
				zap.String("task_id", task.ID))
		} else {
			c.logger.Debug("✅ context ID unchanged",
				zap.String("context_id", c.contextID),
				zap.String("task_id", task.ID))
		}
		c.contextID = task.ContextID
	} else {
		c.logger.Warn("⚠️ task completed but no context ID returned",
			zap.String("task_id", task.ID))
	}

	c.displayTaskResult(&task)
}

func (c *GoogleCalendarClient) handleAsyncResponse(msgParams adk.MessageSendParams) {
	// Submit the task
	resp, err := c.client.SendTask(c.ctx, msgParams)
	if err != nil {
		fmt.Printf("\r❌ Error: %v\n", err)
		return
	}

	// Parse initial task response
	var task adk.Task
	if err := c.parseTaskFromResponse(resp.Result, &task); err != nil {
		fmt.Printf("\r❌ Error parsing response: %v\n", err)
		return
	}

	// Update context ID immediately from the initial response
	if task.ContextID != "" {
		if c.contextID != task.ContextID {
			c.logger.Info("🔄 context updated from initial response",
				zap.String("old_context", c.contextID),
				zap.String("new_context", task.ContextID),
				zap.String("task_id", task.ID))
		} else {
			c.logger.Debug("✅ context ID unchanged from initial response",
				zap.String("context_id", c.contextID),
				zap.String("task_id", task.ID))
		}
		c.contextID = task.ContextID
	} else {
		c.logger.Warn("⚠️ initial task response has no context ID",
			zap.String("task_id", task.ID))
	}

	// If already completed (shouldn't happen in async mode), display result
	if task.Status.State == adk.TaskStateCompleted {
		c.displayTaskResult(&task)
		return
	}

	// Start polling for completion
	c.pollForCompletion(&task)
}

func (c *GoogleCalendarClient) pollForCompletion(task *adk.Task) {
	ticker := time.NewTicker(c.config.PollInterval)
	defer ticker.Stop()

	timeout := time.NewTimer(c.config.MaxPollTimeout)
	defer timeout.Stop()

	dots := 0
	maxDots := 3

	for {
		select {
		case <-c.ctx.Done():
			fmt.Printf("\r❌ Request cancelled\n")
			return

		case <-timeout.C:
			fmt.Printf("\r⏰ Request timed out after %v\n", c.config.MaxPollTimeout)
			return

		case <-ticker.C:
			// Show animated thinking indicator
			fmt.Printf("\r🤔 Thinking%s%s", strings.Repeat(".", dots+1), strings.Repeat(" ", maxDots-dots))
			dots = (dots + 1) % (maxDots + 1)

			// Poll for task status
			taskResp, err := c.client.GetTask(c.ctx, adk.TaskQueryParams{
				ID: task.ID,
			})
			if err != nil {
				c.logger.Debug("failed to get task status", zap.Error(err))
				continue
			}

			// Parse updated task
			var updatedTask adk.Task
			if err := c.parseTaskFromResponse(taskResp.Result, &updatedTask); err != nil {
				c.logger.Debug("failed to parse task response", zap.Error(err))
				continue
			}

			// Check task state
			switch updatedTask.Status.State {
			case adk.TaskStateCompleted:
				// Update context ID from completed task
				if updatedTask.ContextID != "" {
					if c.contextID != updatedTask.ContextID {
						c.logger.Info("🔄 context updated from completed task",
							zap.String("old_context", c.contextID),
							zap.String("new_context", updatedTask.ContextID),
							zap.String("task_id", updatedTask.ID))
					} else {
						c.logger.Debug("✅ context ID unchanged from completed task",
							zap.String("context_id", c.contextID),
							zap.String("task_id", updatedTask.ID))
					}
					c.contextID = updatedTask.ContextID
				} else {
					c.logger.Warn("⚠️ completed task has no context ID",
						zap.String("task_id", updatedTask.ID))
				}
				c.displayTaskResult(&updatedTask)
				return

			case adk.TaskStateFailed:
				errorMsg := "Unknown error occurred"
				if updatedTask.Status.Message != nil {
					errorMsg = c.extractTextFromMessage(updatedTask.Status.Message)
				}
				fmt.Printf("\r❌ Task failed: %s\n", errorMsg)
				return

			case adk.TaskStateCanceled:
				fmt.Printf("\r❌ Task was cancelled\n")
				return

			case adk.TaskStateSubmitted, adk.TaskStateWorking:
				// Continue polling
				continue

			default:
				c.logger.Debug("task in unexpected state", zap.String("state", string(updatedTask.Status.State)))
				continue
			}
		}
	}
}

func (c *GoogleCalendarClient) parseTaskFromResponse(result interface{}, task *adk.Task) error {
	resultBytes, ok := result.(json.RawMessage)
	if !ok {
		return fmt.Errorf("unexpected response result type")
	}
	return json.Unmarshal(resultBytes, task)
}

func (c *GoogleCalendarClient) displayTaskResult(task *adk.Task) {
	// Clear the thinking indicator
	fmt.Print("\r" + strings.Repeat(" ", 20) + "\r")

	c.logger.Debug("displaying task result",
		zap.String("task_id", task.ID),
		zap.Int("history_count", len(task.History)))

	if len(task.History) == 0 {
		fmt.Println("🤖 Agent: No response received")
		return
	}

	// Find all assistant messages and display them
	var assistantMessages []*adk.Message
	for i := range task.History {
		if task.History[i].Role == "assistant" {
			assistantMessages = append(assistantMessages, &task.History[i])
		}
	}

	if len(assistantMessages) == 0 {
		fmt.Println("🤖 Agent: No assistant response found")
		c.logger.Debug("no assistant messages found", zap.Any("history", task.History))
		return
	}

	// Display all assistant messages
	for i, msg := range assistantMessages {
		responseText := c.extractTextFromMessage(msg)
		if responseText != "" {
			if i == 0 {
				fmt.Printf("🤖 Agent: %s\n", responseText)
			} else {
				fmt.Printf("🤖 Agent (continued): %s\n", responseText)
			}
		} else {
			c.logger.Debug("empty response text from message",
				zap.String("message_id", msg.MessageID),
				zap.Any("parts", msg.Parts))

			// If no text found, show a more detailed debug message
			if c.config.LogLevel == "debug" {
				fmt.Printf("🤖 Agent: (No text response - %d parts in message)\n", len(msg.Parts))
			} else {
				fmt.Println("🤖 Agent: (No text response)")
			}
		}
	}

	fmt.Println() // Add spacing before next prompt
}

func (c *GoogleCalendarClient) extractTextFromMessage(message *adk.Message) string {
	var text strings.Builder

	c.logger.Debug("extracting text from message",
		zap.Int("parts_count", len(message.Parts)),
		zap.String("role", message.Role))

	for i, part := range message.Parts {
		c.logger.Debug("processing part", zap.Int("part_index", i), zap.Any("part", part))

		if partMap, ok := part.(map[string]interface{}); ok {
			// Check for text field
			if textContent, exists := partMap["text"]; exists {
				if textStr, ok := textContent.(string); ok {
					c.logger.Debug("found text content", zap.String("text", textStr))
					text.WriteString(textStr)
				}
			}

			// Also check for content field (alternative structure)
			if contentField, exists := partMap["content"]; exists {
				if contentStr, ok := contentField.(string); ok {
					c.logger.Debug("found content field", zap.String("content", contentStr))
					text.WriteString(contentStr)
				}
			}

			// Check for type-specific content
			if partType, exists := partMap["type"]; exists {
				if partType == "text" {
					// Look for text in various possible fields
					for _, field := range []string{"text", "content", "value", "data"} {
						if fieldContent, exists := partMap[field]; exists {
							if fieldStr, ok := fieldContent.(string); ok {
								c.logger.Debug("found text in field", zap.String("field", field), zap.String("text", fieldStr))
								text.WriteString(fieldStr)
							}
						}
					}
				}
			}
		} else {
			// Handle case where part is directly a string
			if partStr, ok := part.(string); ok {
				c.logger.Debug("found direct string part", zap.String("text", partStr))
				text.WriteString(partStr)
			}
		}
	}

	result := text.String()
	c.logger.Debug("extracted text result", zap.String("result", result), zap.Int("length", len(result)))
	return result
}

func (c *GoogleCalendarClient) showHelp() {
	fmt.Println("\n📖 Available Commands and Examples:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("Calendar Queries:")
	fmt.Println("  • What's on my calendar today?")
	fmt.Println("  • Show me my meetings for tomorrow")
	fmt.Println("  • What meetings do I have this week?")
	fmt.Println("  • Do I have any free time on Friday?")
	fmt.Println()
	fmt.Println("Event Management:")
	fmt.Println("  • Schedule a meeting with Sarah at 3 PM tomorrow")
	fmt.Println("  • Create a 1-hour lunch meeting next Tuesday")
	fmt.Println("  • Book a team standup every Monday at 9 AM")
	fmt.Println("  • Cancel my 2 PM meeting today")
	fmt.Println("  • Move my 4 PM meeting to 5 PM")
	fmt.Println()
	fmt.Println("Time Management:")
	fmt.Println("  • When is my next meeting?")
	fmt.Println("  • How much free time do I have today?")
	fmt.Println("  • Find a 30-minute slot for a call this week")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  • help or h - Show this help message")
	fmt.Println("  • status or s - Show current session status and context")
	fmt.Println("  • debug - Show debug information including context ID")
	fmt.Println("  • reset or new - Reset context and start a new conversation")
	fmt.Println("  • clear - Clear the screen")
	fmt.Println("  • quit, exit, or q - Exit the client")
	fmt.Println(strings.Repeat("-", 50) + "\n")
}

func (c *GoogleCalendarClient) clearScreen() {
	fmt.Print("\033[H\033[2J")
	fmt.Println("🗓️  Google Calendar Agent - Interactive Client")
	fmt.Println(strings.Repeat("=", 60) + "\n")
}

func (c *GoogleCalendarClient) showStatus() {
	fmt.Println("\n📊 Client Status:")
	fmt.Println(strings.Repeat("-", 30))
	if c.contextID != "" {
		fmt.Printf("Context ID: %s\n", c.contextID)
		fmt.Println("📝 Conversation active - messages are linked")
	} else {
		fmt.Println("Context ID: (none)")
		fmt.Println("🆕 No active conversation - next message will start new session")
	}
	fmt.Printf("Server URL: %s\n", c.config.ServerURL)
	fmt.Printf("Async Mode: %v\n", c.config.UseAsyncMode)
	fmt.Printf("Log Level: %s\n", c.config.LogLevel)
	fmt.Println(strings.Repeat("-", 30) + "\n")
}

func boolPtr(b bool) *bool {
	return &b
}
