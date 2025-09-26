<div align="center">

# Google-Calendar-Agent
[![CI](https://github.com/inference-gateway/google-calendar-agent/workflows/CI/badge.svg)](https://github.com/inference-gateway/google-calendar-agent/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25.0+-00ADD8?style=flat&logo=go)](https://golang.org)
[![A2A Protocol](https://img.shields.io/badge/A2A-Protocol-blue?style=flat)](https://github.com/inference-gateway/adk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**A Google Calendar A2A agent for AI assistants to interact with Google Calendar**

A production-ready [Agent-to-Agent (A2A)](https://github.com/inference-gateway/adk) server that provides AI-powered capabilities through a standardized protocol.

</div>

## Quick Start

```bash
# Run the agent
go run .

# Or with Docker
docker build -t google-calendar-agent .
docker run -p 8080:8080 google-calendar-agent
```

## Features

- ✅ A2A protocol compliant
- ✅ AI-powered capabilities
- ✅ Streaming support
- ✅ Production ready
- ✅ Minimal dependencies

## Endpoints

- `GET /.well-known/agent-card.json` - Agent metadata and capabilities
- `GET /health` - Health check endpoint
- `POST /a2a` - A2A protocol endpoint

## Available Skills

| Skill | Description | Parameters |
|-------|-------------|------------|
| `list_calendar_events` | List upcoming events from Google Calendar |maxResults, query, timeMax, timeMin |
| `create_calendar_event` | Create a new event in Google Calendar |attendees, description, endTime, location, startTime, summary |
| `update_calendar_event` | Update an existing event in Google Calendar |description, endTime, eventId, location, startTime, summary |
| `delete_calendar_event` | Delete an event from Google Calendar |eventId |
| `get_calendar_event` | Get details of a specific event from Google Calendar |eventId |
| `find_available_time` | Find available time slots in the calendar |duration, endDate, startDate |
| `check_conflicts` | Check for scheduling conflicts in the specified time range |endTime, startTime |

## Configuration

Configure the agent via environment variables:

### Custom Configuration

The following custom configuration variables are available:

| Category | Variable | Description | Default |
|----------|----------|-------------|---------|
| **Google** | `GOOGLE_CREDENTIALS_PATH` | CredentialsPath configuration | `` |
| **Google** | `GOOGLE_SERVICE_ACCOUNT_JSON` | ServiceAccountJson configuration | `` |
| **GoogleCalendar** | `GOOGLE_CALENDAR_ID` | Id configuration | `primary` |
| **GoogleCalendar** | `GOOGLE_CALENDAR_MOCK_MODE` | MockMode configuration | `false` |
| **GoogleCalendar** | `GOOGLE_CALENDAR_TIMEZONE` | Timezone configuration | `UTC` |

| Category | Variable | Description | Default |
|----------|----------|-------------|---------|
| **Server** | `A2A_PORT` | Server port | `8080` |
| **Server** | `A2A_DEBUG` | Enable debug mode | `false` |
| **Server** | `A2A_AGENT_URL` | Agent URL for internal references | `http://localhost:8080` |
| **Server** | `A2A_STREAMING_STATUS_UPDATE_INTERVAL` | Streaming status update frequency | `1s` |
| **Server** | `A2A_SERVER_READ_TIMEOUT` | HTTP server read timeout | `120s` |
| **Server** | `A2A_SERVER_WRITE_TIMEOUT` | HTTP server write timeout | `120s` |
| **Server** | `A2A_SERVER_IDLE_TIMEOUT` | HTTP server idle timeout | `120s` |
| **Server** | `A2A_SERVER_DISABLE_HEALTHCHECK_LOG` | Disable logging for health check requests | `true` |
| **Agent Metadata** | `A2A_AGENT_CARD_FILE_PATH` | Path to agent card JSON file | `.well-known/agent-card.json` |
| **LLM Client** | `A2A_AGENT_CLIENT_PROVIDER` | LLM provider (`openai`, `anthropic`, `azure`, `ollama`, `deepseek`) |`` |
| **LLM Client** | `A2A_AGENT_CLIENT_MODEL` | Model to use |`` |
| **LLM Client** | `A2A_AGENT_CLIENT_API_KEY` | API key for LLM provider | - |
| **LLM Client** | `A2A_AGENT_CLIENT_BASE_URL` | Custom LLM API endpoint | - |
| **LLM Client** | `A2A_AGENT_CLIENT_TIMEOUT` | Timeout for LLM requests | `30s` |
| **LLM Client** | `A2A_AGENT_CLIENT_MAX_RETRIES` | Maximum retries for LLM requests | `3` |
| **LLM Client** | `A2A_AGENT_CLIENT_MAX_CHAT_COMPLETION_ITERATIONS` | Max chat completion rounds | `10` |
| **LLM Client** | `A2A_AGENT_CLIENT_MAX_TOKENS` | Maximum tokens for LLM responses |`4096` |
| **LLM Client** | `A2A_AGENT_CLIENT_TEMPERATURE` | Controls randomness of LLM output |`0.7` |
| **Capabilities** | `A2A_CAPABILITIES_STREAMING` | Enable streaming responses | `true` |
| **Capabilities** | `A2A_CAPABILITIES_PUSH_NOTIFICATIONS` | Enable push notifications | `false` |
| **Capabilities** | `A2A_CAPABILITIES_STATE_TRANSITION_HISTORY` | Track state transitions | `false` |
| **Task Management** | `A2A_TASK_RETENTION_MAX_COMPLETED_TASKS` | Max completed tasks to keep (0 = unlimited) | `100` |
| **Task Management** | `A2A_TASK_RETENTION_MAX_FAILED_TASKS` | Max failed tasks to keep (0 = unlimited) | `50` |
| **Task Management** | `A2A_TASK_RETENTION_CLEANUP_INTERVAL` | Cleanup frequency (0 = manual only) | `5m` |
| **Storage** | `A2A_QUEUE_PROVIDER` | Storage backend (`memory` or `redis`) | `memory` |
| **Storage** | `A2A_QUEUE_URL` | Redis connection URL (when using Redis) | - |
| **Storage** | `A2A_QUEUE_MAX_SIZE` | Maximum queue size | `100` |
| **Storage** | `A2A_QUEUE_CLEANUP_INTERVAL` | Task cleanup interval | `30s` |
| **Authentication** | `A2A_AUTH_ENABLE` | Enable OIDC authentication | `false` |

## Development

```bash
# Generate code from ADL
task generate

# Run tests
task test

# Build the application
task build

# Run linter
task lint

# Format code
task fmt
```

### Debugging

Use the [A2A Debugger](https://github.com/inference-gateway/a2a-debugger) to test and debug your A2A agent during development. It provides a web interface for sending requests to your agent and inspecting responses, making it easier to troubleshoot issues and validate your implementation.

```bash
docker run --rm -it --network host ghcr.io/inference-gateway/a2a-debugger:latest --server-url http://localhost:8080 tasks submit "What are your skills?"
```

```bash
docker run --rm -it --network host ghcr.io/inference-gateway/a2a-debugger:latest --server-url http://localhost:8080 tasks list
```

```bash
docker run --rm -it --network host ghcr.io/inference-gateway/a2a-debugger:latest --server-url http://localhost:8080 tasks get <task ID>
```

## Deployment

### Docker

The Docker image can be built with custom version information using build arguments:

```bash
# Build with default values from ADL
docker build -t google-calendar-agent .

# Build with custom version information
docker build \
  --build-arg VERSION=1.2.3 \
  --build-arg AGENT_NAME="My Custom Agent" \
  --build-arg AGENT_DESCRIPTION="Custom agent description" \
  -t google-calendar-agent:1.2.3 .
```

**Available Build Arguments:**
- `VERSION` - Agent version (default: `0.4.18`)
- `AGENT_NAME` - Agent name (default: `google-calendar-agent`)
- `AGENT_DESCRIPTION` - Agent description (default: `A Google Calendar A2A agent for AI assistants to interact with Google Calendar`)

These values are embedded into the binary at build time using linker flags, making them accessible at runtime without requiring environment variables.

## License

MIT License - see LICENSE file for details
