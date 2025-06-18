# Basic Example

This example demonstrates how to run the Google Calendar Agent with the Inference Gateway using Docker Compose. The setup includes both services configured to work together, providing a complete AI-powered calendar management solution.

## Architecture

### Through Inference Gateway

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│                 │    │                 │    |                 |
│ User/Client     │───▶│ Inference       │───▶│ Calendar Agent  │
│                 │    │ Gateway         │    │                 │
│                 │    │ (Port 8080)     │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │                 │    │                 │
                       │ LLM Providers   │    │ Google Calendar │
                       │ (OpenAI, Groq,  │    │ API             │
                       │  Anthropic,etc) │    │                 │
                       └─────────────────┘    └─────────────────┘
```

**Flow Description:**

1. User sends request to Inference Gateway (port 8080)
2. Inference Gateway processes the request and determines if an agent is needed
3. Calendar Agent interacts with Google Calendar API for calendar operations
4. Inference Gateway uses LLM providers for natural language processing
5. Response flows back through Gateway to the user

### Direct Access with A2A Debugger

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│                 │    │                 │    │                 |
│ A2A Debugger    │───▶│ Calendar Agent  │───▶│ Google Calendar │
│ CLI Tool        │    │ (A2A Server)    │    │ API             │
│                 │    │ (Port 8080)     │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │
        │                       │
        ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│                 │    │                 │
│ Task Management │    │ LLM Engine      │
│ • List tasks    │    │ (Optional)      │
│ • View history  │    │ For AI features │
│ • Submit tasks  │    │                 │
└─────────────────┘    └─────────────────┘
```

**Direct Access Flow:**

1. **A2A Debugger** connects directly to the Calendar Agent's A2A server endpoint
2. **Submit Tasks**: Send JSON-RPC 2.0 requests directly to the agent
3. **Task Management**: List, monitor, and retrieve task details and conversation history
4. **Real-time Debugging**: View agent responses, tool calls, and execution flow
5. **Agent Tools**: Direct access to calendar operations without gateway overhead

**A2A Debugger Commands:**

```bash
# Test connection and get agent capabilities
docker run --rm a2a-debugger connect
docker run --rm a2a-debugger agent-card

# Submit tasks directly to the agent
docker run --rm a2a-debugger tasks submit "List my events for today"
docker run --rm a2a-debugger tasks submit "Create a meeting tomorrow at 2 PM"

# Monitor and debug
docker run --rm a2a-debugger tasks list
docker run --rm a2a-debugger tasks get <task-id>
docker run --rm a2a-debugger tasks history <context-id>
```

**Benefits of Direct Access:**

- **Faster Response**: No gateway overhead for debugging
- **Direct Tool Access**: Immediate access to calendar operations
- **Enhanced Debugging**: Full visibility into agent internals
- **Task Monitoring**: Real-time task status and conversation history
- **Development Workflow**: Perfect for agent development and testing

## Features

- **Google Calendar Agent**: Manages calendar events with natural language processing
- **Inference Gateway**: High-performance LLM gateway supporting multiple providers
- **Multi-Provider Support**: OpenAI, Groq, Anthropic, DeepSeek, Cohere, Cloudflare
- **Demo Mode**: Run without Google Calendar integration for testing
- **Health Checks**: Built-in health monitoring for both services
- **Automatic Restart**: Services restart automatically on failure

## Prerequisites

- Docker and Docker Compose installed
- Google Calendar API credentials (unless running in demo mode)
- API keys for at least one LLM provider

## Quick Start

### 1. Clone and Setup

```bash
# Navigate to the basic example directory
cd examples/basic

# Copy the environment template
cp .env.gateway.example .env.gateway
cp .env.agent.example .env.agent
```

### 2. Configure Environment Variables

Edit the `.env` file and configure the required settings:

#### For Demo Mode (No Google Calendar Integration)

```bash
# Set demo mode to true
DEMO_MODE=true

# Configure at least one LLM provider
GROQ_API_KEY=your_groq_api_key_here
LLM_PROVIDER=groq
LLM_MODEL=deepseek-r1-distill-llama-70b
```

#### For Production Mode (With Google Calendar)

```bash
# Disable demo mode
DEMO_MODE=false

# Configure Google Calendar
GOOGLE_CALENDAR_ID=primary
GOOGLE_CALENDAR_SA_JSON={"type":"service_account","project_id":"..."}

# Configure LLM provider
GROQ_API_KEY=your_groq_api_key_here
LLM_PROVIDER=groq
LLM_MODEL=deepseek-r1-distill-llama-70b
```

### 3. Start the Services

```bash
# Using Task (recommended)
task up

# Or using Docker Compose directly
docker-compose up -d

# View logs
task logs
# or
docker-compose logs -f

# Check service status
task status
# or
docker-compose ps
```

### 4. Test the Setup

#### Check Health Status

```bash
# Test Inference Gateway
curl http://localhost:8080/health
```

#### Get Agent Information

Set on the Inference Gateway `A2A_EXPOSE=true` and bring up the containers.

```bash
curl http://localhost:8080/v1/a2a/agents
```

#### Test Chat Completions

```bash
# Test through Inference Gateway
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek/deepseek-chat",
    "messages": [
      {
        "role": "user",
        "content": "List my events for today"
      }
    ]
  }'
```

## Available Tasks

This example includes a Taskfile for easy management. Here are the available commands:

```bash
# Service management
task up                 # Start all services
task down               # Stop all services
task restart            # Restart all services
task status             # Show service status

# Monitoring
task logs               # Show logs for all services
task logs-gateway       # Show logs for inference gateway only
task logs-agent         # Show logs for calendar agent only
task health             # Check health of all services

# Testing
task test-gateway       # Test Inference Gateway
task test-agent         # Test Calendar Agent directly
task agent-info         # Get agent information

# Maintenance
task clean              # Stop services and remove volumes
task clean-all          # Stop services and remove everything
task pull               # Pull latest images

# Modes
task demo               # Start in demo mode
task prod               # Start in production mode
task debug              # Start with debug logging

# Validation
task validate-env       # Check environment configuration
```

## Configuration Options

### Google Calendar Configuration

| Environment Variable             | Description                             | Default   | Required |
| -------------------------------- | --------------------------------------- | --------- | -------- |
| `DEMO_MODE`                      | Run without Google Calendar integration | `false`   | No       |
| `GOOGLE_CALENDAR_ID`             | Target calendar ID                      | `primary` | No       |
| `GOOGLE_CALENDAR_SA_JSON`        | Service account JSON (single line)      | -         | Yes\*    |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to credentials file                | -         | Yes\*    |
| `GOOGLE_CALENDAR_READ_ONLY`      | Read-only calendar access               | `false`   | No       |
| `GOOGLE_CALENDAR_TIMEZONE`       | Default timezone                        | `UTC`     | No       |

\* Required unless `DEMO_MODE=true`

### LLM Provider Configuration

| Environment Variable | Description                | Default                         | Required |
| -------------------- | -------------------------- | ------------------------------- | -------- |
| `LLM_PROVIDER`       | LLM provider to use        | `groq`                          | No       |
| `LLM_MODEL`          | Model to use               | `deepseek-r1-distill-llama-70b` | No       |
| `LLM_ENABLED`        | Enable LLM functionality   | `true`                          | No       |
| `LLM_TIMEOUT`        | Request timeout            | `30s`                           | No       |
| `LLM_MAX_TOKENS`     | Maximum tokens to generate | `2048`                          | No       |
| `LLM_TEMPERATURE`    | Creativity level (0.0-2.0) | `0.7`                           | No       |

### Supported LLM Providers

#### Groq (Recommended for Speed)

```bash
GROQ_API_KEY=your_groq_api_key
LLM_PROVIDER=groq
LLM_MODEL=deepseek-r1-distill-llama-70b
```

#### OpenAI

```bash
OPENAI_API_KEY=your_openai_api_key
LLM_PROVIDER=openai
LLM_MODEL=gpt-4o
```

#### Anthropic

```bash
ANTHROPIC_API_KEY=your_anthropic_api_key
LLM_PROVIDER=anthropic
LLM_MODEL=claude-3-opus-20240229
```

#### DeepSeek (Cost-Effective)

```bash
DEEPSEEK_API_KEY=your_deepseek_api_key
LLM_PROVIDER=deepseek
LLM_MODEL=deepseek-chat
```

#### Cohere

```bash
COHERE_API_KEY=your_cohere_api_key
LLM_PROVIDER=cohere
LLM_MODEL=command-r-plus
```

#### Cloudflare Workers AI

```bash
CLOUDFLARE_API_TOKEN=your_cloudflare_token
CLOUDFLARE_ACCOUNT_ID=your_account_id
LLM_PROVIDER=cloudflare
LLM_MODEL=@cf/meta/llama-3.1-8b-instruct
```

## Google Calendar Setup

### Option 1: Service Account (Recommended)

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable the Google Calendar API
4. Create a Service Account
5. Download the JSON credentials file
6. Share your calendar with the service account email
7. Set `GOOGLE_CALENDAR_SA_JSON` to the JSON content (single line)

## API Usage Examples

### List Calendar Events

```bash
# Through Inference Gateway
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek/deepseek-chat",
    "messages": [
      {
        "role": "user",
        "content": "What events do I have this week?"
      }
    ]
  }'
```

### Create Calendar Event

```bash
# Through Inference Gateway
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek/deepseek-chat",
    "messages": [
      {
        "role": "user",
        "content": "Schedule a team meeting tomorrow at 2 PM for 1 hour"
      }
    ]
  }'
```

### Update Calendar Event

```bash
# Through Inference Gateway
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek/deepseek-chat",
    "messages": [
      {
        "role": "user",
        "content": "Move my 2 PM meeting to 3 PM"
      }
    ]
  }'
```

### Delete Calendar Event

```bash
# Through Inference Gateway
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek/deepseek-chat",
    "messages": [
      {
        "role": "user",
        "content": "Cancel my meeting with John tomorrow"
      }
    ]
  }'
```

## Troubleshooting

### Common Issues

#### Services Won't Start

```bash
# Check logs for errors
docker-compose logs

# Restart services
docker-compose down
docker-compose up -d
```

#### Health Checks Failing

```bash
# Check service status
docker-compose ps

# Test connectivity
curl http://localhost:8080/health
```

#### Google Calendar Authentication Issues

- Verify credentials are correctly formatted
- Ensure calendar is shared with service account
- Check API quotas in Google Cloud Console

#### LLM Provider Issues

- Verify API keys are correct
- Check provider-specific rate limits
- Try different models if current one fails

### Debug Mode

Enable debug logging for more detailed output:

```bash
# In .env file
LOG_LEVEL=debug
SERVER_GIN_MODE=debug
```

### Viewing Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f google-calendar-agent
docker-compose logs -f inference-gateway

# Last 100 lines
docker-compose logs --tail=100
```

## Cleanup

```bash
# Stop services
docker-compose down

# Remove volumes and networks
docker-compose down -v

# Remove images (optional)
docker-compose down --rmi all
```

## Security Considerations

- Store API keys securely (use Docker secrets in production)
- Use HTTPS in production environments
- Regularly rotate API keys
- Limit Google Calendar permissions to necessary scopes
- Monitor API usage and set up alerts

## Production Deployment

For production deployments, consider:

- Using Docker secrets for sensitive data
- Setting up reverse proxy with SSL termination
- Implementing proper monitoring and logging
- Using managed databases for persistence
- Setting up automated backups
- Implementing health check endpoints

## Support

For issues and questions:

- [Google Calendar Agent Issues](https://github.com/inference-gateway/google-calendar-agent/issues)
- [Inference Gateway Issues](https://github.com/inference-gateway/inference-gateway/issues)
- [Documentation](https://github.com/inference-gateway/docs)
