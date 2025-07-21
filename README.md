<div align="center">

# Google Calendar A2A Agent

[![CI](https://github.com/inference-gateway/google-calendar-agent/workflows/CI/badge.svg)](https://github.com/inference-gateway/google-calendar-agent/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/inference-gateway/google-calendar-agent)](https://goreportcard.com/report/github.com/inference-gateway/google-calendar-agent)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Release](https://img.shields.io/github/v/release/inference-gateway/google-calendar-agent)](https://github.com/inference-gateway/google-calendar-agent/releases)
[![Docker](https://img.shields.io/badge/docker-available-blue?style=flat&logo=docker)](https://github.com/inference-gateway/google-calendar-agent/pkgs/container/google-calendar-agent)

**A production-ready [Agent-to-Agent (A2A)](https://github.com/inference-gateway/a2a) that seamlessly integrates with Google Calendar.**

Enables AI assistants and automated systems to manage calendar events, schedule meetings, and query availability through a standardized protocol. Built with Go for high performance and reliability, with optional mock mode for testing and development.

</div>

## Quick Start

```bash
# Run the agent
go run main.go

# Or with Docker
docker build -t google-calendar-agent .
docker run -p 8080:8080 google-calendar-agent
```

## Features

- ✅ A2A protocol compliant
- ✅ Google Calendar integration (when configured)
- ✅ Minimal dependencies
- ✅ Production ready
- ✅ Mock mode for testing

## Endpoints

- `GET /.well-known/agent.json` - Agent metadata
- `GET /health` - Health check
- `POST /a2a` - A2A protocol endpoint

## Configuration

Configure the agent via environment variables:

### Core Application Settings

- `ENVIRONMENT` - Deployment environment (default: `dev`)
- `DEMO_MODE` - Enable demo mode with mock services (default: `false`)

### Google Calendar Settings

- `GOOGLE_CALENDAR_ID` - Target Google Calendar ID (default: `primary`)
- `GOOGLE_CALENDAR_SA_JSON` - Google Service Account credentials (JSON format)
- `GOOGLE_APPLICATION_CREDENTIALS` - Path to Google credentials file (alternative to SA_JSON)
- `GOOGLE_CALENDAR_READ_ONLY` - Access calendar in read-only mode (default: `false`)
- `GOOGLE_CALENDAR_TIMEZONE` - Default timezone for time inputs (default: `UTC`)

### Logging Configuration

- `LOG_LEVEL` - Log level: `debug`, `info`, `warn`, `error` (default: `info`)
- `LOG_FORMAT` - Log format: `json`, `console` (default: `json`)
- `LOG_OUTPUT` - Log output: `stdout`, `stderr`, or file path (default: `stdout`)
- `LOG_ENABLE_CALLER` - Add caller info to logs (default: `true`)
- `LOG_ENABLE_STACKTRACE` - Add stacktrace to error logs (default: `true`)

### A2A Agent Configuration (ADK)

#### Agent Identity

- `A2A_AGENT_URL` - Agent URL (default: `http://helloworld-agent:8080`)

#### Server Configuration

- `A2A_DEBUG` - Enable debug mode (default: `false`)
- `A2A_TIMEZONE` - Timezone for timestamps (default: `UTC`)
- `A2A_STREAMING_STATUS_UPDATE_INTERVAL` - Interval for streaming status updates (default: `1s`)

#### LLM Client Configuration

- `A2A_AGENT_CLIENT_PROVIDER` - LLM provider: `openai`, `anthropic`, `groq`, `ollama`, `deepseek`, `cohere`, `cloudflare`
- `A2A_AGENT_CLIENT_MODEL` - Model to use
- `A2A_AGENT_CLIENT_API_KEY` - API key for LLM provider
- `A2A_AGENT_CLIENT_BASE_URL` - Custom LLM API endpoint
- `A2A_AGENT_CLIENT_TIMEOUT` - Timeout for LLM requests (default: `30s`)
- `A2A_AGENT_CLIENT_MAX_RETRIES` - Maximum retries for LLM requests (default: `3`)
- `A2A_AGENT_CLIENT_MAX_CHAT_COMPLETION_ITERATIONS` - Maximum chat completion iterations (default: `10`)
- `A2A_AGENT_CLIENT_MAX_TOKENS` - Maximum tokens for LLM responses (default: `4096`)
- `A2A_AGENT_CLIENT_TEMPERATURE` - Controls randomness of LLM output (default: `0.7`)
- `A2A_AGENT_CLIENT_TOP_P` - Top-p sampling parameter (default: `1.0`)
- `A2A_AGENT_CLIENT_FREQUENCY_PENALTY` - Frequency penalty (default: `0.0`)
- `A2A_AGENT_CLIENT_PRESENCE_PENALTY` - Presence penalty (default: `0.0`)
- `A2A_AGENT_CLIENT_SYSTEM_PROMPT` - System prompt to guide the LLM (default: `You are a helpful AI assistant processing an A2A (Agent-to-Agent) task. Please provide helpful and accurate responses.`)
- `A2A_AGENT_CLIENT_MAX_CONVERSATION_HISTORY` - Maximum conversation history per context (default: `20`)
- `A2A_AGENT_CLIENT_USER_AGENT` - User agent string (default: `a2a-agent/1.0`)

#### Capabilities Configuration

- `A2A_CAPABILITIES_STREAMING` - Enable streaming support (default: `true`)
- `A2A_CAPABILITIES_PUSH_NOTIFICATIONS` - Enable push notifications (default: `true`)
- `A2A_CAPABILITIES_STATE_TRANSITION_HISTORY` - Enable state transition history (default: `false`)

#### Authentication Configuration

- `A2A_AUTH_ENABLE` - Enable OIDC authentication (default: `false`)
- `A2A_AUTH_ISSUER_URL` - OIDC issuer URL (default: `http://keycloak:8080/realms/inference-gateway-realm`)
- `A2A_AUTH_CLIENT_ID` - OIDC client ID (default: `inference-gateway-client`)
- `A2A_AUTH_CLIENT_SECRET` - OIDC client secret

#### TLS Configuration

- `A2A_SERVER_TLS_ENABLE` - Enable TLS (default: `false`)
- `A2A_SERVER_TLS_CERT_PATH` - Path to TLS certificate file
- `A2A_SERVER_TLS_KEY_PATH` - Path to TLS private key file

#### Queue Configuration

- `A2A_QUEUE_MAX_SIZE` - Queue maximum size (default: `100`)
- `A2A_QUEUE_CLEANUP_INTERVAL` - Queue cleanup interval (default: `30s`)

#### Server Configuration

- `A2A_SERVER_PORT` - Server port (default: `8080`)
- `A2A_SERVER_READ_TIMEOUT` - Maximum duration for reading requests (default: `120s`)
- `A2A_SERVER_WRITE_TIMEOUT` - Maximum duration for writing responses (default: `120s`)
- `A2A_SERVER_IDLE_TIMEOUT` - Maximum time to wait for next request (default: `120s`)
- `A2A_SERVER_DISABLE_HEALTHCHECK_LOG` - Disable logging for health check requests (default: `true`)

#### Telemetry Configuration

- `A2A_TELEMETRY_ENABLE` - Enable OpenTelemetry metrics collection (default: `false`)
- `A2A_TELEMETRY_METRICS_PORT` - Metrics server port (default: `9090`)
- `A2A_TELEMETRY_METRICS_HOST` - Metrics server host
- `A2A_TELEMETRY_METRICS_READ_TIMEOUT` - Metrics server read timeout (default: `30s`)
- `A2A_TELEMETRY_METRICS_WRITE_TIMEOUT` - Metrics server write timeout (default: `30s`)
- `A2A_TELEMETRY_METRICS_IDLE_TIMEOUT` - Metrics server idle timeout (default: `60s`)

## Example Usage

For a complete working example with Docker Compose setup, see the [example directory](./example/).

```bash
# Test the agent
curl -X POST http://localhost:8080/a2a \
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
  }'
```

## License

MIT
