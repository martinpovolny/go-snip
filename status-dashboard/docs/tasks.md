# Tasks

## Done

- [x] Initial WireGuard peer status dashboard
- [x] Embed static HTML into binary
- [x] Cross-compile + deploy via Makefile to ora-m (amd64)
- [x] Systemd service with auto-restart
- [x] Resolve peer names from `/etc/hosts`
- [x] Fix "never" handshake label → "no handshake" badge + "—" in column
- [x] HTTP monitors tab (second tab in UI)
- [x] `monitors.conf` config file (name + URL + optional basic auth credentials)
- [x] Scrape monitor list from UptimeRobot dashboard via Playwright
- [x] Basic auth support for SCSN, Slovníky, Slovníky SK
- [x] Refactor `main.go` → `main.go`, `wg.go`, `hosts.go`, `monitors.go`
- [x] Fix Makefile: `deploy` (routine) vs `install` (first-time, depends on deploy)
- [x] Rename "wg-status" → "status-dashboard" in UI
- [x] SQLite ring buffer (`store.go`) — `modernc.org/sqlite`, keeps 2016 rows/monitor (~1 week at 5 min)
- [x] Background poller goroutine — checks all monitors every 5 min, stores to DB
- [x] `/monitors-history?monitor=X&limit=N` endpoint
- [x] ntfy notifications on monitor state change (down + recovery), basic auth via `NTFY_PASSWORD` env var
- [x] `.env` / `.env.example` pattern, `.gitignore`, `EnvironmentFile=` in systemd unit
- [x] Alert rate limiting: 2 consecutive failures before down alert, one alert per outage, immediate recovery alert
- [x] `notifications` table in SQLite, `/notifications` endpoint, see `docs/alerting.md`

## Done (continued)

- [x] Sparkline / uptime bar in the monitors tab UI (48 blocks = last 4h, grey=no data, green=up, red=down)
- [x] Alerts tab — notifications log in the UI, refreshes with main poll loop
- [x] Seed alert state from monitor_checks history when no persisted row exists
- [x] Project cleanup: remove stale binaries and names.conf, gitignore .playwright-mcp/

## Upcoming

- [ ] (nothing planned)
