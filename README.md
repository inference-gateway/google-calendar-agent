# Google Calendar A2A Agent

A minimal [Agent-to-Agent (A2A)](https://github.com/inference-gateway/a2a) compatible agent for Google Calendar operations.

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

Set via environment variables:

- `PORT` - Server port (default: 8080)
- `AGENT_NAME` - Agent name (default: google-calendar-agent)
- `AGENT_DESCRIPTION` - Agent description
- `GOOGLE_CREDENTIALS` - Google service account JSON (optional)

## Example Usage

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
