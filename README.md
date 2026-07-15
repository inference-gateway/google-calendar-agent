<div align="center">

# Google Calendar Agent

[![CI](https://github.com/inference-gateway/google-calendar-agent/workflows/CI/badge.svg)](https://github.com/inference-gateway/google-calendar-agent/actions/workflows/ci.yml)
[![Go Report Card](https://img.shields.io/badge/Go%20Report%20Card-A+-brightgreen?style=flat&logo=go&logoColor=white)](https://goreportcard.com/report/github.com/inference-gateway/google-calendar-agent)
[![Go Version](https://img.shields.io/badge/Go-1.26.4+-00ADD8?style=flat&logo=go)](https://golang.org)
[![A2A Protocol](https://img.shields.io/badge/A2A-Protocol-blue?style=flat)](https://github.com/inference-gateway/adk)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)

**A Google Calendar A2A agent for AI assistants to interact with Google Calendar**

A enterprise-ready [Agent-to-Agent (A2A)](https://github.com/inference-gateway/adk) server that provides AI-powered capabilities through a standardized protocol.

</div>

## Quick Start

The generated binary is a CLI. `start` boots the A2A server; `--help` and
`--version` work as you'd expect.

```bash
# Run the agent
go run . start

# Or build and invoke the CLI directly
task build
./bin/google-calendar-agent --help
./bin/google-calendar-agent --version
./bin/google-calendar-agent start

# Or with Docker
docker build -t google-calendar-agent .
docker run -p 8080:8080 google-calendar-agent
```

### CLI

| Command | Description |
|---------|-------------|
| `google-calendar-agent start` | Start the A2A server (blocks until SIGINT/SIGTERM) |
| `google-calendar-agent --help` | Show top-level help (and per-subcommand with `<cmd> --help`) |
| `google-calendar-agent --version` | Print the embedded version and exit |

## Quick Install

Add this agent to your Inference Gateway CLI:

```bash
infer agents add google-calendar-agent http://localhost:8080 \
  --oci ghcr.io/inference-gateway/google-calendar-agent:latest \
  --run
```

## Features

- ✅ A2A protocol compliant
- ✅ AI-powered capabilities
- ✅ Streaming support
- ✅ Enterprise-ready
- ✅ Minimal dependencies

## Endpoints

- `GET /.well-known/agent-card.json` - Agent metadata and capabilities
- `GET /health` - Health check endpoint
- `POST /a2a` - A2A protocol endpoint

## Available Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `Read` | Read a file from disk. Returns its contents, optionally sliced by line offset/limit. Use this to load SKILL.md bodies on demand. | file_path, offset, limit |
| `list_calendar_events` | List upcoming events from Google Calendar | maxResults, query, timeMax, timeMin |
| `create_calendar_event` | Create a new event in Google Calendar | attendees, description, endTime, location, startTime, summary |
| `update_calendar_event` | Update an existing event in Google Calendar | description, endTime, eventId, location, startTime, summary |
| `delete_calendar_event` | Delete an event from Google Calendar | eventId |
| `get_calendar_event` | Get details of a specific event from Google Calendar | eventId |
| `find_available_time` | Find available time slots in the calendar | duration, endDate, startDate |
| `check_conflicts` | Check for scheduling conflicts in the specified time range | endTime, startTime |
| `get_current_datetime` | Return the current date/time and the user's IANA timezone. Call this FIRST for any time-relative request (today, tomorrow, next Friday) before emitting RFC3339 timestamps to other calendar tools, so events land in the user's local timezone instead of an LLM-assumed default. | None |

## Examples

| Example | Description |
|---------|-------------|
| List upcoming events | Ask "What's on my calendar this week?" and the agent calls list_calendar_events to return your upcoming events with their times, locations, and attendees. |
| Schedule a conflict-free meeting | Ask "Schedule a 30-minute sync with alice@example.com tomorrow afternoon." The schedule-meeting skill chains find_available_time, check_conflicts, and create_calendar_event to book a slot that does not overlap anything already on the calendar. |
| Find a free time slot | Ask "Find a free 1-hour slot on Thursday" and the agent uses find_available_time to propose open windows, anchoring "Thursday" to your timezone with get_current_datetime first. |
| Reschedule or cancel an event | Ask "Move my 2 PM meeting to 3 PM" or "Cancel my standup tomorrow"; the agent looks the event up with get_calendar_event and then calls update_calendar_event or delete_calendar_event. |

## Skills (loaded into the system prompt)

| Skill | Description | Source |
|-------|-------------|--------|
| `schedule-meeting` | Use this when the user asks to schedule a meeting, book a slot, or find a time that works. Resolves a conflict-free booking by finding open slots, validating no overlap, and creating the event. | bare scaffold (`skills/schedule-meeting.md`) |

## Documentation
- [Setup](docs/setup.md)
- [Configuration](docs/configuration.md)
- [Usage](docs/usage.md)

## Configuration

Configure the agent via environment variables:

### Custom Configuration

The following custom configuration variables are available. Defaults are
derived from `spec.config.*` in `agent.yaml`; the env vars below override
them at runtime.

| Category | Variable | Default |
|----------|----------|---------|
| **Google** | `GOOGLE_CREDENTIALS_PATH` | `` |
| **Google** | `GOOGLE_SERVICE_ACCOUNT_JSON` | `` |
| **GoogleCalendar** | `GOOGLE_CALENDAR_ID` | `primary` |
| **GoogleCalendar** | `GOOGLE_CALENDAR_MOCK_MODE` | `false` |
| **GoogleCalendar** | `GOOGLE_CALENDAR_TIMEZONE` | `UTC` |
| **Tools** | `TOOLS_READ_ENABLED` | `true` |
| **Tools** | `TOOLS_READ_MAX_LINES` | `2000` |

### Environment Variables

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

### Adding Dependencies

The generator owns the baseline toolchain pins (SDK, server framework,
logging, CLI, sandbox utilities). To extend the project without forking
the templates, declare extras in `agent.yaml` - every empty list below
is rendered by `adl init --defaults` precisely so it's discoverable:

| Where | Purpose | Example entry | Rendered into |
|-------|---------|---------------|---------------|
| `spec.language.go.vendor.deps` | Runtime Go modules | `github.com/stretchr/testify@v1.10.0` | `go.mod` `require` block |
| `spec.language.go.vendor.devdeps` | Executable dev tools (Go 1.24 `tool` directive) | `golang.org/x/tools/cmd/stringer@v0.20.0` | `go.mod` `tool` directive |
| `spec.development.deps` | Cross-cutting sandbox tools (not tied to one language) | `kubectl@1.31.0`, `terraform@1.9.5`, `deno@2.1.4` | Flox `manifest.toml` / devcontainer feature |

Entries use the `<package>@<version>` form. Built-in pins always win on
conflict; the generator prints a warning and skips the user entry when
shadowing is attempted. After editing `agent.yaml`, re-run `task generate`
to refresh the manifests.

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

- `VERSION` - Agent version (default: `0.4.29`)
- `AGENT_NAME` - Agent name (default: `google-calendar-agent`)
- `AGENT_DESCRIPTION` - Agent description (default: `A Google Calendar A2A agent for AI assistants to interact with Google Calendar`)

These values are embedded into the binary at build time using linker flags, making them accessible at runtime without requiring environment variables.

## License

Apache 2.0 License - see LICENSE file for details
