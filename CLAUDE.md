# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Google Calendar A2A (Agent-to-Agent) agent written in Go, implementing the A2A protocol for AI assistants to interact with Google Calendar. The agent provides calendar operations like listing events, creating events, managing schedules, and finding available time slots.

## Key Commands

### Development
```bash
task build:dev        # Build for development (no version info)
task build           # Build with version information
task test            # Run all tests
task test:coverage   # Run tests with coverage
task lint            # Run golangci-lint
task tidy            # Clean up Go module dependencies
```

### Code Generation
```bash
task a2a:download:schema  # Download latest A2A schema
task generate            # Generate Go code from A2A schema - MUST run after schema changes
```

### Docker
```bash
task build:docker    # Build Docker image with version tags
```

### Running
```bash
go run cmd/agent/main.go  # Run the agent directly
./dist/agent              # Run compiled binary (after task build)
```

## Architecture

### Core Components

1. **A2A Server** (`cmd/agent/main.go`): Entry point that initializes the A2A server with Google Calendar tools. Handles both demo mode (AI disabled) and production mode.

2. **Google Calendar Service** (`google/`): 
   - `calendar.go`: Core service implementing CalendarService interface
   - `credentials.go`: Handles Google API authentication
   - Uses Google Calendar API v3 for operations
   - Supports both read-only and full access modes

3. **Toolbox** (`toolbox/`):
   - `toolbox.go`: Registers Google Calendar operations as A2A tools
   - `handlers.go`: Individual tool handlers for calendar operations
   - Provides mock mode for testing without Google credentials

4. **Configuration** (`config/`):
   - Environment-based configuration via `sethvargo/go-envconfig`
   - Supports Google credentials via JSON string or file path
   - Extensive A2A server configuration options

5. **A2A Protocol** (`a2a/`):
   - `generated_types.go`: Auto-generated types from A2A schema (DO NOT EDIT)
   - `schema.yaml`: A2A protocol schema definition
   - Custom types and error handling

## Important Development Notes

1. **Generated Files**: Never modify files with `generated_` prefix - they're auto-generated from schema
2. **Testing Philosophy**: Use table-driven tests with isolated mock servers for each test case
3. **Code Style**: 
   - Use early returns to avoid deep nesting
   - Prefer switch statements over if-else chains
   - Code to interfaces for easier mocking
4. **Type Safety**: Always prefer strong typing and interfaces over dynamic typing

## Configuration

The agent is configured via environment variables with three main groups:
- **Google Calendar**: `GOOGLE_CALENDAR_*` for calendar-specific settings
- **A2A Agent**: `A2A_*` for agent protocol configuration  
- **Logging**: `LOG_*` for logging configuration

Key settings:
- `DEMO_MODE=true`: Run without Google credentials using mocks
- `GOOGLE_CALENDAR_SA_JSON`: Service account credentials as JSON string
- `A2A_AGENT_URL`: Required agent URL configuration

## Testing

Run tests with proper isolation:
```bash
task test                     # Run all tests
go test ./google/...         # Test Google Calendar service
go test ./toolbox/...        # Test toolbox handlers
go test ./config/...         # Test configuration
```

## Workflow Before Committing

1. Make your code changes
2. Run `task generate` if A2A schema was updated
3. Run `task lint` to check code quality
4. Run `task build` to verify compilation
5. Run `task test` to ensure all tests pass
6. Commit your changes

## Mock Mode

The agent supports a mock mode for development and testing:
- Automatically activates when `DEMO_MODE=true`
- Falls back to mock in dev environment if Google credentials fail
- Provides simulated calendar operations without external dependencies