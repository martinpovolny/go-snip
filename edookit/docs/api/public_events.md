# Veřejné události (Public Events)

> **Legacy module** — served at `<school>-login.edookit.net`, not the main domain.  
> **No authentication required.** Data are publicly accessible.

## Configuration

| Field | Value |
|-------|-------|
| Source name | Veřejné události |
| Domain | `https://<school>-login.edookit.net` |
| Auth | None |

## Endpoints

### List events

```
GET /api/public/v1/events
```

Returns school public events in a date range. Optionally returns iCalendar format.

**Query parameters**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `from` | date | no | Start of range (YYYY-MM-DD) |
| `to` | date | no | End of range (YYYY-MM-DD) |
| `ical` | — | no | If present (any value), returns iCalendar (.ics) instead of JSON |

**Response (200)**

```json
{
  "version": "2.0",
  "events": [
    {
      "id": 123,
      "title": "Třídnická hodina",
      "description": "",
      "place": "",
      "participant": "5.A",
      "from": "2017-01-16T08:00:00+00:00",
      "to": "2017-01-16T08:59:59+00:00"
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | int | Unique event ID |
| `title` | string | Event title |
| `description` | string | HTML description (may be absent if school hides it) |
| `place` | string | Comma-separated room names |
| `participant` | string | Comma-separated class/group names |
| `from` | string | Start time (ISO 8601) |
| `to` | string | End time (ISO 8601) |

## Go usage

```go
c := edookit.New("https://yourschool.edookit.net", user, pass)
events, err := c.ListPublicEvents(edookit.PublicEventsOpts{
    From: "2026-09-01",
    To:   "2026-12-31",
})
```

## curl

```bash
bin/api '/api/public/v1/events?from=2026-01-01&to=2026-12-31'
```
