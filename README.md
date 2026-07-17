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
- ✅ OpenTelemetry instrumentation
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
| [List upcoming events](examples/list-upcoming-events/) | Ask "What's on my calendar this week?" and the agent calls list_calendar_events to return your upcoming events with their times, locations, and attendees. |
| [Schedule a conflict-free meeting](examples/schedule-a-conflict-free-meeting/) | Ask "Schedule a 30-minute sync with alice@example.com tomorrow afternoon." The schedule-meeting skill chains find_available_time, check_conflicts, and create_calendar_event to book a slot that does not overlap anything already on the calendar. |
| [Find a free time slot](examples/find-a-free-time-slot/) | Ask "Find a free 1-hour slot on Thursday" and the agent uses find_available_time to propose open windows, anchoring "Thursday" to your timezone with get_current_datetime first. |
| [Reschedule or cancel an event](examples/reschedule-or-cancel-an-event/) | Ask "Move my 2 PM meeting to 3 PM" or "Cancel my standup tomorrow"; the agent looks the event up with get_calendar_event and then calls update_calendar_event or delete_calendar_event. |

## Skills (loaded into the system prompt)

| Skill | Description | Source |
|-------|-------------|--------|
| `schedule-meeting` | Use this when the user asks to schedule a meeting, book a slot, or find a time that works. Resolves a conflict-free booking by finding open slots, validating no overlap, and creating the event. | bare scaffold (`skills/schedule-meeting.md`) |

## Documentation
- [Setup](docs/setup.md)
- [Configuration](docs/configuration.md)
- [Usage](docs/usage.md)

## Configuration

The agent is configured via environment variables. Defaults are derived
from `agent.yaml`; see [CONFIGURATIONS.md](CONFIGURATIONS.md) for the
full reference of custom and `A2A_*` variables.

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
