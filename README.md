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

- `APP_ENVIRONMENT` - Deployment environment (default: `dev`)
- `APP_DEMO_MODE` - Enable demo mode with mock services (default: `false`)
- `APP_MAX_REQUEST_SIZE` - Maximum request body size in bytes (default: `1048576`)
- `APP_REQUEST_TIMEOUT` - Maximum duration for handling requests (default: `30s`)

### Server Configuration

- `SERVER_PORT` - Server port (default: `8080`)
- `SERVER_HOST` - Host to bind to (default: `0.0.0.0`)
- `SERVER_GIN_MODE` - Gin server mode: `debug`, `release`, `test` (default: `release`)
- `SERVER_ENABLE_TLS` - Enable HTTPS (default: `false`)
- `SERVER_READ_TIMEOUT` - Maximum duration for reading requests (default: `10s`)
- `SERVER_WRITE_TIMEOUT` - Maximum duration for writing responses (default: `10s`)
- `SERVER_IDLE_TIMEOUT` - Maximum time to wait for next request (default: `60s`)

### Google Calendar Settings

- `GOOGLE_CALENDAR_ID` - Target Google Calendar ID (default: `primary`)
- `GOOGLE_CALENDAR_SA_JSON` - Google Service Account credentials (JSON format)
- `GOOGLE_APPLICATION_CREDENTIALS` - Path to Google credentials file (alternative to SA_JSON)
- `GOOGLE_CALENDAR_READ_ONLY` - Access calendar in read-only mode (default: `false`)
- `GOOGLE_CALENDAR_TIMEZONE` - Default timezone for time inputs (default: `UTC`)

### LLM Configuration

- `LLM_GATEWAY_URL` - Inference Gateway or OpenAI-compatible API URL (default: `http://localhost:8080/v1`)
- `LLM_PROVIDER` - LLM provider: `openai`, `anthropic`, `groq`, `ollama`, `deepseek`, `cohere`, `cloudflare` (default: `groq`)
- `LLM_MODEL` - Model to use (default: `deepseek-r1-distill-llama-70b`)
- `LLM_TIMEOUT` - Timeout for LLM requests (default: `30s`)
- `LLM_MAX_TOKENS` - Maximum tokens to generate (default: `2048`)
- `LLM_TEMPERATURE` - Generation randomness 0.0-2.0 (default: `0.7`)
- `LLM_ENABLED` - Enable LLM functionality (default: `true`)

### Logging Configuration

- `LOG_LEVEL` - Log level: `debug`, `info`, `warn`, `error` (default: `info`)
- `LOG_FORMAT` - Log format: `json`, `console` (default: `json`)
- `LOG_OUTPUT` - Log output: `stdout`, `stderr`, or file path (default: `stdout`)
- `LOG_ENABLE_CALLER` - Add caller info to logs (default: `true`)
- `LOG_ENABLE_STACKTRACE` - Add stacktrace to error logs (default: `true`)

### TLS Configuration (when `SERVER_ENABLE_TLS=true`)

- `TLS_CERT_PATH` - Path to TLS certificate file
- `TLS_KEY_PATH` - Path to TLS private key file
- `TLS_MIN_VERSION` - Minimum TLS version: `1.2`, `1.3` (default: `1.2`)
- `TLS_CIPHER_SUITES` - Comma-separated list of cipher suites

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
