# Configuration

Google Calendar Agent is configured via environment variables. Defaults are
derived from `agent.yaml`; the env vars below override them at runtime.

## Custom Configuration

| Category | Variable | Default |
|----------|----------|---------|
| **Google** | `GOOGLE_CREDENTIALS_PATH` | `` |
| **Google** | `GOOGLE_SERVICE_ACCOUNT_JSON` | `` |
| **GoogleCalendar** | `GOOGLE_CALENDAR_ID` | `primary` |
| **GoogleCalendar** | `GOOGLE_CALENDAR_MOCK_MODE` | `false` |
| **GoogleCalendar** | `GOOGLE_CALENDAR_TIMEZONE` | `UTC` |
| **Tools** | `TOOLS_READ_ENABLED` | `true` |
| **Tools** | `TOOLS_READ_MAX_LINES` | `2000` |

## Environment Variables

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
| **Telemetry** | `A2A_TELEMETRY_ENABLE` | Enable OpenTelemetry instrumentation | `true` |
| **Telemetry** | `A2A_OTEL_TRACES_EXPORTER` | Trace exporter (`otlp` or `none`) | `none` |
| **Telemetry** | `A2A_OTEL_METRICS_EXPORTER` | Metrics exporter (`otlp`, `prometheus`, or `none`) | `prometheus` |
| **Telemetry** | `A2A_OTEL_EXPORTER_PROMETHEUS_PORT` | Prometheus metrics exporter port | `9464` |
