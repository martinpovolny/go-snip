# Edookit API — Module Overview

All modules use HTTP Basic authentication unless noted. Credentials are configured
per-module by the school admin at **Nastavení → API přístupové údaje**.

## Base URLs

| Domain | Purpose |
|--------|---------|
| `https://<school>.edookit.net` | All standard modules |
| `https://<school>-login.edookit.net` | Legacy modules (Změnový rozvrh, Veřejné události) |

> **Note:** The `-login` domain is the teacher/admin system. Most modules are on
> the parent/student portal (`<school>.edookit.net`). Do not mix them up — this
> is the most common source of errors.

## Module Summary

| Module | Source name | Domain | Auth | Go file | Status |
|--------|-------------|--------|------|---------|--------|
| [Docházka](attendance.md) | Docházka | main | HMAC-SHA1 + vendor reg | — | needs vendor |
| [Hodiny](lessons.md) | Hodiny | main | HTTP Basic | `lessons.go` | — |
| [Vyhledání osob](persons.md) | Vyhledávání osob | main | HTTP Basic | `persons.go` | ✅ |
| [Osobní údaje](students.md) | Osobní údaje žáků / zaměstnanců | main | HTTP Basic | `students.go` | — |
| [Statistiky](stats.md) | Statistické informace | main | HTTP Basic | `stats.go` | — |
| [Hodnocení](grades.md) | Hodnocení | main | HTTP Basic | `grades.go` | — |
| [Platby](payments.md) | Platby | main | HTTP Basic | `payments.go` | — |
| [Kurzy](courses.md) | Kurzy | main | HTTP Basic | `courses.go` | — |
| [Veřejné události](public_events.md) | Veřejné události | **legacy** | **none (public)** | `events.go` | ✅ |
| [Změnový rozvrh](schedule.md) | Změnový rozvrh | **legacy** | HTTP Basic | `schedule.go` | ✅ |
| [Přímý zápis docházky](attendance_direct.md) | Přímý zápis docházky | main | HTTP Basic | `attendance_direct.go` | — |
| [Individuální cíle](individual_goals.md) | Individuální cíle | main | HTTP Basic | `individual_goals.go` | — |

## SPA Scraping (fallback)

When REST API modules are not provisioned (401), the parent portal at
`https://<school>.edookit.net` can be scraped with a cookie-authenticated HTTP
client. See [spa_internal.md](spa_internal.md) for full route map, data shapes,
Nette signal protocol, and a Go scraping sketch.

## Quick-start

```bash
# Fill in credentials
cp .env.example .env && $EDITOR .env

# Smoke-test all enabled modules
go run ./cmd/apitest/

# Ad-hoc curl (credentials from .env, never exposed on CLI)
bin/api /api/person-search/v1/200
bin/api '/api/scheduler/v1/change?from=2026-05-21&to=2026-05-27'
bin/api '/api/public/v1/events?from=2026-01-01&to=2026-12-31'
```

## Authentication details

### HTTP Basic
Standard RFC 7617. Username and password are set by the school admin in Edookit
at **Nastavení → API přístupové údaje**. These are separate credentials from the
web-login username/password — the admin must create them explicitly per module.

### HMAC-SHA1 (Docházka only)
Three custom headers required:

| Header | Value |
|--------|-------|
| `com.edookit.Client` | Vendor client ID (from Edookit vendor registration) |
| `com.edookit.Auth` | `<značka>:<HMAC-SHA1 hex>` |
| `com.edookit.Time` | Timestamp `YYYY-MM-DD HH:MM:SS.sss` (within ±15 min of server) |

HMAC-SHA1 input string: `METHOD+PATH+TIMESTAMP+HESLO`  
HMAC key: the vendor-assigned `clientKey`.  
See [attendance.md](attendance.md) for full details and example.
