# Usage

Once the agent is running (see [Setup](setup.md)), you interact with it over
the A2A protocol. It translates natural-language calendar requests into calls
to the tools below.

## Endpoints

- `POST /a2a` — A2A protocol endpoint (JSON-RPC 2.0)
- `GET /.well-known/agent-card.json` — agent metadata and capabilities
- `GET /health` — health check

## Tools

| Tool | What it does |
|------|--------------|
| `list_calendar_events` | List upcoming events, optionally filtered by time range or search query |
| `get_calendar_event` | Fetch the details of a single event by ID |
| `create_calendar_event` | Create an event with a summary, start/end time, attendees, and location |
| `update_calendar_event` | Change the time, summary, or location of an existing event |
| `delete_calendar_event` | Remove an event by ID |
| `find_available_time` | Propose open slots of a given duration within a date range |
| `check_conflicts` | Report whether a time range overlaps existing events |
| `get_current_datetime` | Return the current time and the user's IANA timezone |

## Timezone handling

For any time-relative request ("today", "tomorrow", "next Friday"), the agent
calls `get_current_datetime` first to anchor the current time and timezone,
then emits RFC3339 timestamps with the correct offset. Set
`GOOGLE_CALENDAR_TIMEZONE` (see [Configuration](configuration.md)) to control
the default when a request does not name a timezone.

## schedule-meeting skill

When you ask to book a meeting, the agent loads the `schedule-meeting`
playbook (`skills/schedule-meeting/SKILL.md`) and chains
`find_available_time` → `check_conflicts` → `create_calendar_event` to produce
a conflict-free booking.

## Try it with the A2A Debugger

```bash
docker run --rm -it --network host \
  ghcr.io/inference-gateway/a2a-debugger:latest \
  --server-url http://localhost:8080 tasks submit "What's on my calendar today?"
```

More example requests to try:

```text
Schedule a 30 minute sync with alice@example.com tomorrow afternoon
Find a free 1-hour slot for the team on Thursday
Move my 2 PM meeting to 3 PM
Cancel my standup tomorrow
```
