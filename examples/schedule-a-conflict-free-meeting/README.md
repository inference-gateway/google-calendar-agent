# Schedule a conflict-free meeting

Ask "Schedule a 30-minute sync with alice@example.com tomorrow afternoon." The `schedule-meeting` skill chains `find_available_time`, `check_conflicts`, and `create_calendar_event` to book a slot that does not overlap anything already on the calendar.

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
          {"text": "Schedule a 30-minute sync with alice@example.com tomorrow afternoon"}
        ]
      }
    }
  }' | jq .
```

## What the agent does

1. Receives the natural-language request and loads the `schedule-meeting` skill.
2. Calls `get_current_datetime` to anchor "tomorrow afternoon" to the user's timezone.
3. Calls `find_available_time` with the target window (e.g., tomorrow 1 PM – 5 PM) and a 30-minute duration to discover open slots.
4. Presents the candidate slots to you and asks which one you prefer.
5. After you confirm a slot, calls `check_conflicts` to verify the slot is still free.
6. If no conflicts, calls `create_calendar_event` with the summary, start/end times, and attendee.
7. Returns the created event details.

## Example response (mock mode)

After the agent finds available slots and you confirm one:

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
            "text": "I'\''ve scheduled a 30-minute sync with alice@example.com.\n\n**Event details:**\n- **Title:** Sync\n- **Date:** Thursday, July 17, 2026\n- **Time:** 2:00 PM – 2:30 PM (America/New_York)\n- **Attendees:** alice@example.com\n- **Event ID:** abc123def456\n\nThe event has been created and added to your calendar."
          }
        ]
      }
    }
  }
}
```

## Full interaction flow

The agent may ask clarifying questions before booking. Here is a typical multi-turn flow:

**Turn 1 — User request:**
```json
{"text": "Schedule a 30-minute sync with alice@example.com tomorrow afternoon"}
```

**Turn 1 — Agent response (asking for confirmation):**
```json
{
  "text": "I found a few open slots tomorrow afternoon:\n\n1. **1:00 PM – 1:30 PM**\n2. **2:00 PM – 2:30 PM**\n3. **3:30 PM – 4:00 PM**\n\nWhich slot works best for you?"
}
```

**Turn 2 — User confirms:**
```json
{"text": "The 2 PM slot works"}
```

**Turn 2 — Agent books and confirms:**
```json
{
  "text": "Great! I'\''ve booked the 2:00 PM – 2:30 PM sync with alice@example.com. Event ID: abc123def456."
}
```

## Related tools

- `find_available_time` — discovers open slots in a date range
- `check_conflicts` — verifies a specific slot has no overlapping events
- `create_calendar_event` — creates the event once a slot is confirmed
- `get_current_datetime` — anchors relative time expressions to the user's timezone
