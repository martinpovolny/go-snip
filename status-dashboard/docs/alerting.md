# Alerting logic

Alerts are sent to ntfy (`status-dashboard` topic) when a monitored HTTP endpoint
changes state. The goal is one alert per event — no floods, no repeats.

## Rules

### Down alert
Fires when a monitor has failed **2 consecutive polls** in a row.

- 1 failed poll → silent (transient blip)
- 2 failed polls in a row → 🔴 DOWN alert sent once
- Further failures while still down → **suppressed** (no repeat alerts)

With the default 5-minute poll interval, a real outage triggers an alert after ~10 minutes.

### Recovery alert
Fires **immediately** on the first successful poll after a confirmed outage
(i.e. after a down alert was sent).

- ✅ Recovery alert sent once
- Failure counter reset, suppression cleared

### First observation
On service startup, the first poll result for each monitor establishes a silent
baseline. No alert is sent regardless of status, to avoid noise on restarts.

## Notification log

Every sent notification is recorded in the SQLite `notifications` table:

```
id | sent_at | monitor | event (down/recovery) | title | message
```

Exposed via `GET /notifications?limit=N` (default 100, newest first).

## Configuration

| Constant | Value | Location |
|---|---|---|
| `failuresBeforeAlert` | `2` | `notify.go` |
| Poll interval | `5m` | `main.go` `StartPoller` call |

## Notification format

| Event | Priority | Title |
|---|---|---|
| Down | `urgent` | 🔴 `<name>` is DOWN |
| Recovery | `default` | ✅ `<name>` is back up |

Message body includes the HTTP status code (or error string) and response time.
