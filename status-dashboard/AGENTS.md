# Agent Instructions

## Project

`status-dashboard` is a Go web service running on `ora-m` (130.61.58.229:9999).
It shows two tabs: WireGuard peer status and HTTP endpoint monitor status.

## Build & Deploy

```bash
make deploy          # build linux/amd64, copy binary + monitors.conf, restart service
make install         # first-time: runs deploy + copies systemd unit + enables it
```

Remote machine: `ora-m` (ssh alias), amd64, Ubuntu, systemd.
Service: `/etc/systemd/system/status-dashboard.service`
Binary + config: `/opt/status-dashboard/`

Check service health after deploy:
```bash
ssh ora-m 'sudo systemctl status status-dashboard'
```

## Verifying a Deployment with Playwright

After `make deploy`, use the Playwright MCP to verify the dashboard is working.

**Dashboard URL:** http://192.168.77.1:9999 (via WireGuard VPN) or http://130.61.58.229:9999 (public)

### Step-by-step verification

1. **Navigate to the dashboard**
   ```
   browser_navigate: http://130.61.58.229:9999
   ```

2. **Check the peers tab loads**
   Use `browser_snapshot` and verify:
   - The "WireGuard peers" tab is active
   - The peers table contains rows (not "No peers found")
   - At least some peers show "connected" badge

3. **Check the monitors tab**
   Click the "HTTP monitors" tab, then `browser_snapshot` and verify:
   - The monitors table contains rows for all 8 endpoints
   - At least some show a green "2xx" badge
   - None show a red "401" (basic auth is working for SCSN, Slovníky, Slovníky SK)

4. **Check for JS errors**
   Use `browser_console_messages` to confirm no errors in the console.

### What to look for

| Check | Pass | Fail |
|---|---|---|
| Page loads | Title = "status-dashboard" | HTTP error / blank page |
| Peers data | Table has ≥1 row | "No peers found" or error banner |
| Monitors data | All 8 rows visible | Error banner or missing rows |
| Auth working | SCSN/Slovníky show 2xx | Those rows show 401 |
| No JS errors | Console messages = 0 errors | Any console errors |

### Common issues

- **Service not running**: `ssh ora-m 'sudo systemctl restart status-dashboard'`
- **Wrong binary arch**: rebuild with `make linux-amd64` (target is amd64, not arm64)
- **Config not deployed**: `make deploy` copies `monitors.conf` automatically
- **Port not reachable**: check Oracle Cloud security list allows TCP 9999 inbound

## Source layout

```
main.go        # flags, HTTP handlers, startup wiring
wg.go          # WireGuard peer collection (Collector, Peer, wgDump, ping)
hosts.go       # /etc/hosts parser (parseHosts)
monitors.go    # HTTP monitor collection (Monitor, MonitorCollector, loadMonitors)
store.go       # (planned) SQLite ring buffer for monitor time-series
static/
  index.html   # embedded single-page dashboard (two tabs)
monitors.conf  # list of HTTP endpoints to monitor (name  url  [user:pass])
Makefile
status-dashboard.service
docs/tasks.md  # task tracking
```

## monitors.conf format

```
# name          url                          [user:password]
SNČJ            https://sncj.ujc.cas.cz/
Slovníky        http://ujc.hmpf.cz:3000/    mp:kompost11!
```
