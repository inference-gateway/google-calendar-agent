---
name: schedule-meeting
description: Use this when the user asks to schedule a meeting, book a slot, or "find a time that works". Resolves a conflict-free booking by finding open slots, validating no overlap, and creating the event.
tags:
  - calendar
  - scheduling
  - meeting
---

# schedule-meeting

## When to use

- "Schedule a meeting with X next week"
- "Find a 30-min slot tomorrow afternoon"
- "Book lunch with the team Thursday"
- Any request that combines availability discovery with event creation.

If the user only asks to *list* events or *delete* an event, this skill does
not apply - use the corresponding tool directly.

## Workflow

1. Clarify required inputs with the user before any tool call:
   - **title** (event summary)
   - **duration** in minutes
   - **attendees** (list of email addresses, may be empty)
   - **target window** as a date range (RFC3339 start and end)
   - **location** (optional)
2. Call `find_available_time` with the target window and duration. Take the
   first candidate slot the user accepts.
3. Call `check_conflicts` against the chosen slot. Conflicts here mean another
   event already overlaps - `find_available_time` results can be stale by the
   time the user confirms.
4. If conflicts exist:
   - Surface the conflicting events to the user.
   - Propose the next candidate slot.
   - Do **not** auto-overwrite or schedule on top of an existing event.
5. Once a slot is confirmed conflict-free, call `create_calendar_event` with
   the final `summary`, `startTime`, `endTime`, `attendees`, and `location`.
6. Report back the created event's ID and start time. If the call fails,
   surface the error verbatim - do not retry blindly.

## Tools

- `find_available_time` - discover candidate windows
- `check_conflicts` - verify the chosen slot is still free
- `create_calendar_event` - finalize the booking

## Notes

- All timestamps use RFC3339 (`2026-05-19T14:00:00Z`).
- Default duration when the user is vague: 30 minutes.
- When the user gives a relative window ("tomorrow afternoon"), resolve it
  against the agent's configured timezone before passing to the tools.
