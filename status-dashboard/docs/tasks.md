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

## In Progress

- [ ] SQLite ring buffer (`store.go`) for HTTP monitor time-series
- [ ] Background goroutine: check all monitors every 5 minutes (independent of browser)

## Upcoming

- [ ] Expose historical data via `/monitors-history` endpoint
- [ ] Sparkline / uptime bar in the monitors tab UI
