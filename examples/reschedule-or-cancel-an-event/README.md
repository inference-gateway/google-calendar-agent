# Reschedule or cancel an event

Ask "Move my 2 PM meeting to 3 PM" or "Cancel my standup tomorrow"; the agent looks the event up with `get_calendar_event` and then calls `update_calendar_event` or `delete_calendar_event`.

## Prerequisites

- The agent is running (see [Setup](../../docs/setup.md))
- Mock mode is enabled (`GOOGLE_CALENDAR_MOCK_MODE=true`) or real Google Calendar credentials are configured

## Example: Reschedule an event

### Request

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
          {"text": "Move my 2 PM meeting to 3 PM"}
        ]
      }
    }
  }' | jq .
```

### What the agent does

1. Receives the request and calls `get_current_datetime` to anchor "my 2 PM meeting" to the user's timezone.
2. Calls `list_calendar_events` with a time range around 2 PM today to find the meeting.
3. If multiple events match, asks which one you mean.
4. Calls `get_calendar_event` to fetch the full details of the identified event.
5. Calls `check_conflicts` to verify the new time (3 PM) is free.
6. If no conflicts, calls `update_calendar_event` with the new start/end times.
7. Returns the updated event details.

### Example response (mock mode)

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
            "text": "I'\''ve moved your **Design Review** from 2:00 PM to 3:00 PM today.\n\n**Updated event:**\n- **Title:** Design Review\n- **Date:** Thursday, July 17, 2026\n- **New time:** 3:00 PM – 4:00 PM (America/New_York)\n- **Event ID:** abc123def456\n\nThere were no conflicts at the new time."
          }
        ]
      }
    }
  }
}
```

## Example: Cancel an event

### Request

```bash
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
          {"text": "Cancel my standup tomorrow"}
        ]
      }
    }
  }' | jq .
```

### What the agent does

1. Receives the request and calls `get_current_datetime` to anchor "tomorrow."
2. Calls `list_calendar_events` with a time range covering tomorrow's business hours.
3. Finds the event matching "standup."
4. Calls `get_calendar_event` to confirm the event details.
5. Asks for confirmation before deleting.
6. After you confirm, calls `delete_calendar_event` with the event ID.
7. Confirms the deletion.

### Example response (mock mode)

After you confirm the deletion:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "id": "task-2",
    "status": {
      "state": "completed",
      "message": {
        "role": "agent",
        "parts": [
          {
            "text": "The **Team Standup** scheduled for **Friday, July 18, 2026 at 11:00 AM** has been cancelled and removed from your calendar."
          }
        ]
      }
    }
  }
}
```

## Related tools

- `list_calendar_events` — finds events matching a description or time range
- `get_calendar_event` — fetches full details of a specific event by ID
- `update_calendar_event` — changes the time, summary, or location of an event
- `delete_calendar_event` — removes an event from the calendar
- `check_conflicts` — verifies the new time slot is free before rescheduling
- `get_current_datetime` — anchors relative time expressions to the user's timezone
