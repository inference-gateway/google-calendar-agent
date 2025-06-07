<div align="center">

# Google Calendar Agent (A2A)

[![CI](https://github.com/inference-gateway/google-calendar-agent/workflows/CI/badge.svg)](https://github.com/inference-gateway/google-calendar-agent/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/badge/Docker-Supported-2496ED?style=flat&logo=docker)](https://hub.docker.com/)
[![Go Report Card](https://goreportcard.com/badge/github.com/inference-gateway/google-calendar-agent)](https://goreportcard.com/report/github.com/inference-gateway/google-calendar-agent)
[![GitHub release](https://img.shields.io/github/release/inference-gateway/google-calendar-agent.svg)](https://github.com/inference-gateway/google-calendar-agent/releases)
[![GitHub issues](https://img.shields.io/github/issues/inference-gateway/google-calendar-agent.svg)](https://github.com/inference-gateway/google-calendar-agent/issues)
[![GitHub stars](https://img.shields.io/github/stars/inference-gateway/google-calendar-agent.svg?style=social&label=Star)](https://github.com/inference-gateway/google-calendar-agent)

**A comprehensive Google Calendar agent built with Go that implements the Agent-to-Agent (A2A) protocol for seamless calendar management through natural language interactions.**

</div>

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Development](#development)
- [Deployment](#deployment)
- [Testing](#testing)
- [Contributing](#contributing)

## Overview

This agent provides a natural language interface to Google Calendar through the A2A protocol, enabling users to manage their calendar events using conversational commands. The agent supports listing calendars, viewing events, creating appointments, updating meetings, and canceling events - all through simple text commands.

## Architecture

### System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client App    │    │  A2A Protocol   │    │ Google Calendar │
│                 │◄──►│     Agent       │◄──►│      API        │
│ (Chat/Voice UI) │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Google Calendar Agent                    │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   HTTP      │  │    A2A      │  │   Natural Language  │  │
│  │   Server    │  │ Protocol    │  │     Processing      │  │
│  │ (Gin/REST)  │  │  Handler    │  │                     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Calendar   │  │   Request   │  │    Calendar API     │  │
│  │  Service    │  │  Parser &   │  │    Integration      │  │
│  │ Interface   │  │ Dispatcher  │  │                     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Google    │  │   Mock      │  │      Logging &      │  │
│  │  Calendar   │  │  Service    │  │     Monitoring      │  │
│  │    API      │  │ (Demo Mode) │  │     (Zap Logger)    │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Request Flow

```
User Input → A2A Protocol → Natural Language Parser → Calendar Service → Google API
    ↓             ↓                    ↓                     ↓              ↓
"Schedule      JSON-RPC         Pattern Matching        Calendar        Create
meeting       Request          & Event Parsing         Interface       Event
tomorrow"     Validation                               Abstraction      Call
    ↓             ↓                    ↓                     ↓              ↓
Response ← A2A Response ← Formatted Response ← Service Response ← API Response
```

## Features

### Core Capabilities

- **📅 Calendar Discovery**: List and explore available Google Calendars
- **📋 Event Listing**: View events for today, tomorrow, this week, or custom date ranges
- **➕ Event Creation**: Schedule new meetings, appointments, and events using natural language
- **✏️ Event Updates**: Modify existing events (time, location, title)
- **🗑️ Event Deletion**: Cancel and remove events from calendar
- **🔄 Demo Mode**: Test functionality without Google API credentials

### Supported Commands

| Operation          | Example Commands                                                                     |
| ------------------ | ------------------------------------------------------------------------------------ |
| **List Calendars** | "List my calendars", "What calendars do I have?", "Find my calendar ID"              |
| **View Events**    | "Show my events today", "What's on my calendar this week?", "List meetings tomorrow" |
| **Create Events**  | "Schedule meeting with John at 2pm tomorrow", "Book dentist appointment Friday 10am" |
| **Update Events**  | "Move my 2pm meeting to 3pm", "Change meeting location to Conference Room A"         |
| **Delete Events**  | "Cancel my dentist appointment", "Delete the lunch meeting with Sarah"               |

### A2A Protocol Support

- **JSON-RPC 2.0**: Compliant request/response handling
- **Message Streaming**: Real-time communication support
- **Task Management**: Stateful task tracking and management
- **Agent Discovery**: Self-describing capabilities via `.well-known/agent.json`
- **Multiple Content Types**: Text and JSON response formats

## Quick Start

### Prerequisites

- Go 1.24+ installed
- Google Cloud Project with Calendar API enabled
- Service Account with Calendar API permissions
- Docker (optional, for containerized deployment)

### 1. Clone Repository

```bash
git clone https://github.com/inference-gateway/google-calendar-agent.git
cd google-calendar-agent
```

### 2. Setup Google Calendar API

1. Create a Google Cloud Project
2. Enable the Google Calendar API
3. Create a Service Account
4. Download the service account JSON key
5. Share your calendar with the service account email

### 3. Environment Configuration

```bash
# Required: Google Service Account JSON (as string)
export GOOGLE_CALENDAR_SA_JSON='{"type":"service_account",...}'

# Optional: Specific calendar ID (defaults to "primary")
export GOOGLE_CALENDAR_ID="your-calendar-id@gmail.com"

# Optional: Server port (defaults to 8080)
export PORT=8080
```

### 4. Run the Agent

```bash
# Development mode
task build:dev
./bin/google-calendar-agent

# Or with demo mode (no Google API required)
./bin/google-calendar-agent -demo

# Or with Docker
docker build -t calendar-agent .
docker run -p 8080:8080 -e GOOGLE_CALENDAR_SA_JSON='...' calendar-agent
```

### 5. Test the Agent

```bash
# Health check
curl http://localhost:8080/health

# Agent capabilities
curl http://localhost:8080/.well-known/agent.json

# Send a calendar request
curl -X POST http://localhost:8080/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "message/send",
    "params": {
      "message": {
        "parts": [{"kind": "text", "text": "Show my events today"}]
      }
    },
    "id": "1"
  }'
```

## Configuration

### Environment Variables

| Variable                  | Description                      | Default     | Required               |
| ------------------------- | -------------------------------- | ----------- | ---------------------- |
| `GOOGLE_CALENDAR_SA_JSON` | Service account JSON credentials | -           | Yes (unless demo mode) |
| `GOOGLE_CALENDAR_ID`      | Target calendar ID               | `"primary"` | No                     |
| `PORT`                    | HTTP server port                 | `8080`      | No                     |

### Command Line Options

```bash
./google-calendar-agent [options]

Options:
  -calendar-id string     Google calendar ID to use
  -credentials string     Path to Google credentials file
  -demo                   Run in demo mode with mock service
  -help                   Show help information
  -log-level string       Log level (debug, info, warn, error) (default "debug")
  -port string           Server port
  -version               Show version information
```

## API Reference

### A2A Endpoints

| Endpoint                  | Method | Description                     |
| ------------------------- | ------ | ------------------------------- |
| `/a2a`                    | POST   | Main A2A protocol endpoint      |
| `/health`                 | GET    | Health check endpoint           |
| `/.well-known/agent.json` | GET    | Agent capabilities and metadata |

### Supported A2A Methods

- `message/send` - Send a message and receive response
- `message/stream` - Send a streaming message (maps to message/send)
- `task/get` - Get task status (not implemented)
- `task/cancel` - Cancel a running task (not implemented)

### Response Format

```json
{
  "jsonrpc": "2.0",
  "id": "request-id",
  "result": {
    "taskId": "task-uuid",
    "status": "completed",
    "message": {
      "role": "assistant",
      "parts": [{ "kind": "text", "text": "Response message" }]
    },
    "artifacts": [
      {
        "artifactId": "artifact-uuid",
        "name": "calendar-response",
        "parts": [{ "kind": "text", "text": "Formatted response" }]
      }
    ]
  }
}
```

## Development

### Project Structure

```
├── cmd/
│   ├── codegen/                # Code generation from A2A schema
│   └── google-calendar-agent/  # Main application entry point
├── a2a/
│   ├── agent.go                # A2A protocol handler and calendar logic
│   ├── generated_types.go      # Generated A2A protocol types
│   └── a2a-schema.yaml         # A2A protocol schema definition
├── google/
│   ├── calendar.go             # Google Calendar API service interface
│   ├── credentials.go          # Google credentials management
│   └── mocks/                  # Mock implementations for testing
└── internal/
    └── codegen/                # Internal code generation utilities
```

### Development Workflow

```bash
# Install dependencies
go mod download

# Generate code from A2A schema
task generate

# Run linting
task lint

# Build the project
task build

# Run tests
task test

# Run in development mode
task build:dev && ./bin/google-calendar-agent -demo
```

### Available Tasks

- `task a2a:download:schema` - Download latest A2A schema
- `task generate` - Generate Go code from schema
- `task lint` - Run code linters
- `task build` - Build with version information
- `task build:dev` - Build for development
- `task build:docker` - Build Docker image
- `task test` - Run test suite

## Deployment

### Docker Deployment

```dockerfile
# Build
docker build -t calendar-agent .

# Run
docker run -d \
  --name calendar-agent \
  -p 8080:8080 \
  -e GOOGLE_CALENDAR_SA_JSON='{"type":"service_account",...}' \
  -e GOOGLE_CALENDAR_ID="your-calendar@gmail.com" \
  calendar-agent
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: calendar-agent
spec:
  replicas: 2
  selector:
    matchLabels:
      app: calendar-agent
  template:
    metadata:
      labels:
        app: calendar-agent
    spec:
      containers:
        - name: calendar-agent
          image: calendar-agent:latest
          ports:
            - containerPort: 8080
          env:
            - name: GOOGLE_CALENDAR_SA_JSON
              valueFrom:
                secretKeyRef:
                  name: google-credentials
                  key: service-account.json
```

When running on GKE, please use Identity Workload Federation to authenticate with Google APIs securely.

## Testing

The project includes comprehensive testing with mock implementations:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./a2a/...
```

### Mock Service

The agent includes a mock calendar service for testing and demo purposes:

```bash
# Run in demo mode
./google-calendar-agent -demo
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following the coding standards
4. Run tests and linting (`task test && task lint`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Coding Standards

- Follow Go best practices and idioms
- Use early returns to reduce nesting
- Prefer switch statements over if-else chains
- Implement table-driven tests
- Code to interfaces for better testability
- Always run `task generate`, `task lint`, `task build`, and `task test` before committing

---

**Version**: See `./google-calendar-agent -version` for current version information  
**License**: See LICENSE file for details  
**Support**: Open an issue on GitHub for questions or bug reports
