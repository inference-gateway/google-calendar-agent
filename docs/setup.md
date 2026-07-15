# Setup

This guide covers installing, building, and running the Google Calendar
agent — an [A2A](https://github.com/inference-gateway/adk) server that exposes
Google Calendar operations to AI assistants.

## Prerequisites

- Go 1.26.4+ (only needed to build or run from source)
- An OpenAI-compatible LLM provider and API key
- Google Calendar credentials — a service account JSON or a credentials file
  (not required in mock mode; see [Configuration](configuration.md))

## Run from source

```bash
# Start the A2A server (defaults to port 8080)
go run . start
```

## Build the binary

```bash
task build
./bin/google-calendar-agent --version
./bin/google-calendar-agent start
```

`start` boots the server and blocks until it receives SIGINT/SIGTERM. The
root command also exposes `--help` and `--version`.

## Run with Docker

```bash
docker build -t google-calendar-agent .
docker run -p 8080:8080 google-calendar-agent
```

## Minimal configuration

The agent needs an LLM provider to interpret requests. Set these before
starting:

```bash
export A2A_AGENT_CLIENT_PROVIDER=openai   # openai, anthropic, azure, ollama, deepseek
export A2A_AGENT_CLIENT_MODEL=gpt-4o
export A2A_AGENT_CLIENT_API_KEY=sk-...
```

To try the agent without Google credentials, enable mock mode:

```bash
export GOOGLE_CALENDAR_MOCK_MODE=true
```

See [Configuration](configuration.md) for the full list of variables and
[Usage](usage.md) for example requests.

## Verify it is running

```bash
curl http://localhost:8080/health
curl http://localhost:8080/.well-known/agent-card.json
```
