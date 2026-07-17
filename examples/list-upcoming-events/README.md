# List upcoming events

Ask "What's on my calendar this week?" and the agent calls `list_calendar_events` to return your upcoming events with their times, locations, and attendees.

## Prerequisites

- The agent is running (see [Setup](../../docs/setup.md))
- Mock mode is enabled (`GOOGLE_CALENDAR_MOCK_MODE=true`) or real Google Calendar credentials are configured

## Example request

Send a natural-language prompt to the A2A endpoint:

```bash
curl -s http://localhost:8080/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tasks.send",
    "params": {
      "id": "task-1",
      "message": {
        "role": "user",
        "parts": [
          {"text": "What'\''s on my calendar this week?"}
        ]
      }
    }
  }' | jq .
```

## What the agent does

1. Receives the natural-language request.
2. Calls `get_current_datetime` to anchor the current time and the user's timezone.
3. Computes the start and end of "this week" in RFC3339 format.
4. Calls `list_calendar_events` with `timeMin` and `timeMax` set to the week boundaries.
5. Returns the list of events with their summaries, start/end times, locations, and attendees.

## Example response (mock mode)

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "id": "task-1",
    "status": {
      "state": "completed",
      "message": {
        "role": "agent",
        "parts": [
          {
            "text": "Here are the events on your calendar this week:\n\n**Monday, July 14**\n- 10:00 AM – 11:00 AM: **Sprint Planning** (Conference Room A)\n- 2:00 PM – 3:00 PM: **Design Review** (Virtual)\n\n**Wednesday, July 16**\n- 9:30 AM – 10:30 AM: **1:1 with Manager** (Zoom)\n\n**Friday, July 18**\n- 11:00 AM – 12:00 PM: **Team Standup** (Conference Room B)\n\nWould you like more details on any of these events?"
          }
        ]
      }
    }
  }
}
```

## Try variations

```bash
# Ask about a specific day
curl -s http://localhost:8080/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tasks.send",
    "params": {
      "id": "task-2",
      "message": {
        "role": "user",
        "parts": [
          {"text": "What meetings do I have tomorrow?"}
        ]
      }
    }
  }' | jq .

# Search for specific events
curl -s http://localhost:8080/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tasks.send",
    "params": {
      "id": "task-3",
      "message": {
        "role": "user",
        "parts": [
          {"text": "Find events about sprint planning"}
        ]
      }
    }
  }' | jq .
```

## Related tools

- `list_calendar_events` — the tool called to fetch events
- `get_calendar_event` — fetch details of a single event by ID
- `get_current_datetime` — anchors time-relative queries to the user's timezone
