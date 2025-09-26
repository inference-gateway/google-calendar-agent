# AGENTS.md

This file describes the agents available in this A2A (Agent-to-Agent) system.

## Agent Overview

### google-calendar-agent
**Version**: 0.4.16  
**Description**: A Google Calendar A2A agent for AI assistants to interact with Google Calendar

This agent is built using the Agent Definition Language (ADL) and provides A2A communication capabilities.

## Agent Capabilities



- **Streaming**: ✅ Real-time response streaming supported


- **Push Notifications**: ❌ Server-sent events not supported


- **State History**: ❌ State transition history not tracked



## AI Configuration





**System Prompt**: You are a Google Calendar AI agent specialized in calendar management and scheduling operations.

Your primary capabilities:
1. **Event Management**: Create, update, delete, and retrieve calendar events
2. **Scheduling Intelligence**: Find available time slots and check for conflicts
3. **Calendar Operations**: List events with flexible time ranges and search queries

Key features:
- Support for both mock mode (demo/testing) and production Google Calendar API
- RFC3339 timestamp handling for accurate scheduling
- Intelligent conflict detection and availability checking
- Attendee management and location tracking
- Comprehensive event search and filtering

When helping users:
- Always validate time formats and ranges
- Provide clear feedback on scheduling conflicts
- Suggest alternative time slots when conflicts are detected
- Handle both simple and complex scheduling scenarios
- Maintain data accuracy and consistency with Google Calendar

Your responses should be accurate, helpful, and focused on calendar management tasks.



**Configuration:**




## Skills


This agent provides 7 skills:


### list_calendar_events
- **Description**: List upcoming events from Google Calendar
- **Tags**: calendar, events, list, google
- **Input Schema**: Defined in agent configuration
- **Output Schema**: Defined in agent configuration


### create_calendar_event
- **Description**: Create a new event in Google Calendar
- **Tags**: calendar, events, create, google
- **Input Schema**: Defined in agent configuration
- **Output Schema**: Defined in agent configuration


### update_calendar_event
- **Description**: Update an existing event in Google Calendar
- **Tags**: calendar, events, update, google
- **Input Schema**: Defined in agent configuration
- **Output Schema**: Defined in agent configuration


### delete_calendar_event
- **Description**: Delete an event from Google Calendar
- **Tags**: calendar, events, delete, google
- **Input Schema**: Defined in agent configuration
- **Output Schema**: Defined in agent configuration


### get_calendar_event
- **Description**: Get details of a specific event from Google Calendar
- **Tags**: calendar, events, get, google
- **Input Schema**: Defined in agent configuration
- **Output Schema**: Defined in agent configuration


### find_available_time
- **Description**: Find available time slots in the calendar
- **Tags**: calendar, availability, scheduling, google
- **Input Schema**: Defined in agent configuration
- **Output Schema**: Defined in agent configuration


### check_conflicts
- **Description**: Check for scheduling conflicts in the specified time range
- **Tags**: calendar, conflicts, scheduling, google
- **Input Schema**: Defined in agent configuration
- **Output Schema**: Defined in agent configuration




## Server Configuration

**Port**: 8080

**Debug Mode**: ❌ Disabled



**Authentication**: ❌ Not required


## API Endpoints

The agent exposes the following HTTP endpoints:

- `GET /.well-known/agent-card.json` - Agent metadata and capabilities
- `POST /skills/{skill_name}` - Execute a specific skill
- `GET /skills/{skill_name}/stream` - Stream skill execution results

## Environment Setup

### Required Environment Variables

Key environment variables you'll need to configure:



- `PORT` - Server port (default: 8080)

### Development Environment


**Flox Environment**: ✅ Configured for reproducible development setup




## Usage

### Starting the Agent

```bash
# Install dependencies
go mod download

# Run the agent
go run main.go

# Or use Task
task run
```


### Communicating with the Agent

The agent implements the A2A protocol and can be communicated with via HTTP requests:

```bash
# Get agent information
curl http://localhost:8080/.well-known/agent-card.json



# Execute list_calendar_events skill
curl -X POST http://localhost:8080/skills/list_calendar_events \
  -H "Content-Type: application/json" \
  -d '{"input": "your_input_here"}'

# Execute create_calendar_event skill
curl -X POST http://localhost:8080/skills/create_calendar_event \
  -H "Content-Type: application/json" \
  -d '{"input": "your_input_here"}'

# Execute update_calendar_event skill
curl -X POST http://localhost:8080/skills/update_calendar_event \
  -H "Content-Type: application/json" \
  -d '{"input": "your_input_here"}'

# Execute delete_calendar_event skill
curl -X POST http://localhost:8080/skills/delete_calendar_event \
  -H "Content-Type: application/json" \
  -d '{"input": "your_input_here"}'

# Execute get_calendar_event skill
curl -X POST http://localhost:8080/skills/get_calendar_event \
  -H "Content-Type: application/json" \
  -d '{"input": "your_input_here"}'

# Execute find_available_time skill
curl -X POST http://localhost:8080/skills/find_available_time \
  -H "Content-Type: application/json" \
  -d '{"input": "your_input_here"}'

# Execute check_conflicts skill
curl -X POST http://localhost:8080/skills/check_conflicts \
  -H "Content-Type: application/json" \
  -d '{"input": "your_input_here"}'


```

## Deployment


**Deployment Type**: Manual
- Build and run the agent binary directly
- Use provided Dockerfile for containerized deployment



### Docker Deployment
```bash
# Build image
docker build -t google-calendar-agent .

# Run container
docker run -p 8080:8080 google-calendar-agent
```


## Development

### Project Structure

```
.
├── main.go              # Server entry point
├── skills/              # Business logic skills

│   └── list_calendar_events.go   # List upcoming events from Google Calendar

│   └── create_calendar_event.go   # Create a new event in Google Calendar

│   └── update_calendar_event.go   # Update an existing event in Google Calendar

│   └── delete_calendar_event.go   # Delete an event from Google Calendar

│   └── get_calendar_event.go   # Get details of a specific event from Google Calendar

│   └── find_available_time.go   # Find available time slots in the calendar

│   └── check_conflicts.go   # Check for scheduling conflicts in the specified time range

├── .well-known/         # Agent configuration
│   └── agent-card.json  # Agent metadata
├── go.mod               # Go module definition
└── README.md            # Project documentation
```


### Testing

```bash
# Run tests
task test
go test ./...

# Run with coverage
task test:coverage
```


## Contributing

1. Implement business logic in skill files (replace TODO placeholders)
2. Add comprehensive tests for new functionality
3. Follow the established code patterns and conventions
4. Ensure proper error handling throughout
5. Update documentation as needed

## Agent Metadata

This agent was generated using ADL CLI v0.4.16 with the following configuration:

- **Language**: Go
- **Template**: Minimal A2A Agent
- **ADL Version**: adl.dev/v1

---

For more information about A2A agents and the ADL specification, visit the [ADL CLI documentation](https://github.com/inference-gateway/adl-cli).
