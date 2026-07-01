# Změnový rozvrh (Schedule Changes)

> **Legacy module** — served at `<school>-login.edookit.net`, not the main domain.

## Configuration

| Field | Value |
|-------|-------|
| Source name | Změnový rozvrh |
| Domain | `https://<school>-login.edookit.net` |
| Auth | HTTP Basic |

## Endpoints

### List changes

```
GET /api/scheduler/v1/change
```

Returns only lessons where something changed from the original timetable.

**Query parameters**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `from` | date | no | Start date (YYYY-MM-DD). Default: today |
| `to` | date | no | End date (YYYY-MM-DD). Default: today |
| `teacherNameFormat` | string | no | `code` (default), `name`, or `full` |

**Response (200)**

```json
{
  "version": "2.0",
  "change": [
    {
      "scheduled": {
        "courses": ["D - 9.A"],
        "rooms": ["A101"],
        "teachers": ["JER"],
        "students": ["9.A"],
        "timerange": { "from": "2017-01-10 10:10:00", "to": "2017-01-10 10:55:00" }
      },
      "actual": {
        "courses": ["D - 9.A", "D - 9.B"],
        "rooms": ["A102"],
        "teachers": ["JIL"],
        "students": ["9.A", "9.B"],
        "timerange": null,
        "event": "Prázdniny"
      }
    }
  ]
}
```

Each change has a `scheduled` (original timetable) and `actual` (what will really happen) slot.

| Field | Type | Description |
|-------|------|-------------|
| `courses` | string[] | Course abbreviations |
| `rooms` | string[] | Room names |
| `teachers` | string[] | Teacher codes/names (format per `teacherNameFormat`) |
| `students` | string[] | Participating classes/groups |
| `timerange` | object\|null | `from`/`to` timestamps; `null` = lesson cancelled |
| `event` | string | Present only on `actual` when replaced by a named event |

## Go usage

```go
c := edookit.New("https://yourschool.edookit.net", user, pass)
changes, err := c.ListScheduleChanges(edookit.ScheduleChangeOpts{
    From: "2026-05-19",
    To:   "2026-05-23",
})
```

## curl

```bash
bin/api '/api/scheduler/v1/change?from=2026-05-19&to=2026-05-23'
```
