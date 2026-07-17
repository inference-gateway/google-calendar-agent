# Find a free time slot

Ask "Find a free 1-hour slot on Thursday" and the agent uses `find_available_time` to propose open windows, anchoring "Thursday" to your timezone with `get_current_datetime` first.

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
          {"text": "Find a free 1-hour slot on Thursday"}
        ]
      }
    }
  }' | jq .
```

## What the agent does

1. Receives the natural-language request.
2. Calls `get_current_datetime` to determine the current time and the user's IANA timezone.
3. Computes the date of the next Thursday in the user's timezone.
4. Calls `find_available_time` with `startDate` and `endDate` covering Thursday's business hours (e.g., 9 AM – 5 PM) and `duration` set to 60 minutes.
5. Returns the list of available time slots.

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
            "text": "Here are the available 1-hour slots on **Thursday, July 17, 2026** (America/New_York):\n\n1. **9:00 AM – 10:00 AM**\n2. **10:00 AM – 11:00 AM**\n3. **11:00 AM – 12:00 PM**\n4. **1:00 PM – 2:00 PM**\n5. **2:00 PM – 3:00 PM**\n6. **3:00 PM – 4:00 PM**\n\nWould you like me to book any of these slots?"
          }
        ]
      }
    }
  }
}
```

## Try variations

```bash
# Find a 30-minute slot on a specific date range
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
          {"text": "What times are free next Monday for a 45-minute meeting?"}
        ]
      }
    }
  }' | jq .

# Find availability in the afternoon only
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
          {"text": "Find a free 2-hour slot this Friday afternoon"}
        ]
      }
    }
  }' | jq .
```

## Related tools

- `find_available_time` — the tool called to discover open slots
- `get_current_datetime` — anchors relative day names ("Thursday", "next Monday") to the user's timezone
- `check_conflicts` — verify a specific slot is still free before booking
