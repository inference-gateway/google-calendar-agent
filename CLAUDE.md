# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

google-calendar-agent is an A2A (Agent-to-Agent) server implementing the [A2A Protocol](https://github.com/inference-gateway/adk) for agent-to-agent communication. A Google Calendar A2A agent for AI assistants to interact with Google Calendar. The project is automatically generated from ADL (Agent Definition Language) specifications defined in `agent.yaml`.

## Core Architecture

### ADL-Generated Structure

The codebase is generated using ADL CLI 0.20.4 and follows a strict generation pattern:
- **Generated Files**: Marked with `DO NOT EDIT` headers - manual changes will be overwritten
- **Configuration Source**: `agent.yaml` - defines agent capabilities, skills, and metadata
- **Server Implementation**: Built on the ADK (Agent Development Kit) framework from `github.com/inference-gateway/adk`

### Key Components

- **Main Entry Point**: `main.go` - Configures and starts the A2A server with:
  - OpenAI-compatible LLM client configuration
  - Agent builder with system prompt from `agent.yaml`
  - A2A server with streaming and background task handlers
  - Graceful shutdown handling

- **Agent Configuration**: `.well-known/agent.json` - Serves agent metadata at runtime
- **Environment Configuration**: Extensive env vars with `A2A_` prefix (see README for full list)

## Development Commands

```bash
# Generate/regenerate code from ADL specification
task generate

# Run the agent in development mode (debug enabled, port 8080)
task run

# Run tests (note: no tests currently exist)
task test
task test:cover  # with coverage

# Code quality
task lint         # Run golangci-lint
task fmt          # Format code with go fmt

# Build
task build        # Creates bin/google-calendar-agent
task docker:build # Build Docker image

# Clean build artifacts
task clean
```

## Testing Individual Components

```bash
# Run specific test file (when tests are added)
go test -v ./path/to/package -run TestFunctionName

# Debug with A2A Debugger
docker run --rm -it --network host ghcr.io/inference-gateway/a2a-debugger:latest \
  --server-url http://localhost:8080 tasks submit "Your query"
```

## LLM Provider Configuration

The agent uses OpenAI-compatible LLM client. Configure with:
- `A2A_AGENT_CLIENT_PROVIDER`: `openai`, `anthropic`, `azure`, `ollama`, `deepseek`
- `A2A_AGENT_CLIENT_MODEL`: Model identifier
- `A2A_AGENT_CLIENT_API_KEY`: Provider API key
- `A2A_AGENT_CLIENT_BASE_URL`: Custom endpoint (optional)

## Adding New Functionality

### Skills Implementation
The following skills are currently defined:
- **list_calendar_events**: List upcoming events from Google Calendar
- **create_calendar_event**: Create a new event in Google Calendar
- **update_calendar_event**: Update an existing event in Google Calendar
- **delete_calendar_event**: Delete an event from Google Calendar
- **get_calendar_event**: Get details of a specific event from Google Calendar
- **find_available_time**: Find available time slots in the calendar
- **check_conflicts**: Check for scheduling conflicts in the specified time range

To modify skills:
1. Update `agent.yaml` with skill definitions
2. Run `task generate` to regenerate the codebase
3. Implement skill logic in generated skill files (look for TODO placeholders)
4. Write tests for each skill

### Modifying Agent Behavior

- **System Prompt**: Edit in `agent.yaml`, then regenerate
- **Capabilities**: Modify in `agent.yaml` (streaming, pushNotifications, stateTransitionHistory)
- **Server Configuration**: Update environment variables or `agent.yaml` server section

## Testing Strategy

When implementing tests:
- Create `*_test.go` files alongside implementation files
- Use table-driven tests for comprehensive coverage
- Mock external dependencies (LLM client, Redis if used)
- Test A2A protocol compliance with integration tests

## Environment Management
The project includes Flox environment configuration (`.flox/env/manifest.toml`) providing:
- Go 1.24
- golangci-lint (linter)
- go-task (Task runner)
- Docker
- Claude Code CLI

Activate with: `flox activate` (if Flox is installed)

## Important Constraints

- **Generated Files**: Never manually edit files with "DO NOT EDIT" headers
- **Configuration Changes**: Always modify `agent.yaml` and regenerate
- **ADL Version**: Ensure ADL CLI 0.20.4 or compatible version for regeneration
- **Port Configuration**: Default 8080, configurable via `A2A_PORT` or `A2A_SERVER_PORT`

## Debugging Tips

- Enable debug mode: `A2A_DEBUG=true`
- Check health: `GET /health`
- View agent metadata: `GET /.well-known/agent.json`
- Monitor streaming updates: Set `A2A_STREAMING_STATUS_UPDATE_INTERVAL`
- Use A2A Debugger container for interactive testing
