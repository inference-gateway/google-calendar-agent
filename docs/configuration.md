# Configuration

The agent is configured through environment variables. Defaults come from
`spec.config.*` in `agent.yaml`; the variables below override them at runtime.

## Google Calendar

| Variable | Description | Default |
|----------|-------------|---------|
| `GOOGLE_SERVICE_ACCOUNT_JSON` | Service account credentials as a single-line JSON string | `` |
| `GOOGLE_CREDENTIALS_PATH` | Path to a Google credentials JSON file | `` |
| `GOOGLE_CALENDAR_ID` | Calendar to operate on | `primary` |
| `GOOGLE_CALENDAR_MOCK_MODE` | Serve in-memory mock data instead of calling Google | `false` |
| `GOOGLE_CALENDAR_TIMEZONE` | Default IANA timezone when a request does not specify one | `UTC` |

Provide credentials with either `GOOGLE_SERVICE_ACCOUNT_JSON` (inline) or
`GOOGLE_CREDENTIALS_PATH` (file). Share the target calendar with the service
account's email address so it can read and write events.

When `GOOGLE_CALENDAR_MOCK_MODE=true`, credentials are not required and the
agent returns deterministic sample data — useful for demos and local testing.

## LLM client

| Variable | Description | Default |
|----------|-------------|---------|
| `A2A_AGENT_CLIENT_PROVIDER` | LLM provider (`openai`, `anthropic`, `azure`, `ollama`, `deepseek`) | `` |
| `A2A_AGENT_CLIENT_MODEL` | Model identifier | `` |
| `A2A_AGENT_CLIENT_API_KEY` | Provider API key | - |
| `A2A_AGENT_CLIENT_BASE_URL` | Custom endpoint (optional) | - |
| `A2A_AGENT_CLIENT_MAX_TOKENS` | Maximum tokens per response | `4096` |
| `A2A_AGENT_CLIENT_TEMPERATURE` | Sampling temperature | `0.7` |

## Server

| Variable | Description | Default |
|----------|-------------|---------|
| `A2A_PORT` | Server port | `8080` |
| `A2A_DEBUG` | Enable debug logging | `false` |
| `A2A_SERVER_READ_TIMEOUT` | HTTP read timeout | `120s` |
| `A2A_SERVER_WRITE_TIMEOUT` | HTTP write timeout | `120s` |

## Read tool

The agent loads skill playbooks from disk with a built-in `read` tool.

| Variable | Description | Default |
|----------|-------------|---------|
| `TOOLS_READ_ENABLED` | Enable the read tool | `true` |
| `TOOLS_READ_MAX_LINES` | Maximum lines returned per read | `2000` |

The [README](../README.md#environment-variables) lists the complete
environment variable reference, including task-retention and storage settings.
