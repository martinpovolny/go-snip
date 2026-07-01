# Edookit — Claude Code Guide

## Project status

- **Root package (`edookit`, this directory's `.go` files)** — the API client library. **This is going to production.** Treat changes here with production-grade care: correctness, tests, no speculative features.
- **`cmd/tui/`** — a bubbletea terminal UI. This is a research/experiment scaffold for exploring the API interactively, not a production artifact. Lower rigor is fine here.
- Import path for consumers: `github.com/martinpovolny/go-snip/edookit`. See `docs/library_usage.md` for how another Go project should import and use it.

## Project docs

- `docs/plan.md` — module-by-module live API status, principles, architecture
- `docs/api/` — per-module request/response reference (scraped from Edookit docs)
- `docs/library_usage.md` — how to import this package from another Go project

## Testing rules

- **Never run mutating or destructive API calls in tests — reads only.** No `CreatePayment`, `CreatePaymentPrescription`, `Update*`, `Delete*`, etc. This applies to ad-hoc `curl`/`bin/api` exploration too.
- Live API tests live in `client_live_test.go` at the package root, gated behind the `live` build tag so they never run as part of a normal `go test ./...`:
  ```bash
  go test -tags live -v .
  ```
- Credentials come from `.env` in this directory (`API_URL`, `API_USER`, `API_PASSWORD`), loaded automatically by `TestMain` via `godotenv.Load()`. Tests skip (not fail) if those aren't set.
- `TestLiveListPaymentPrescriptions` is currently skipped — the endpoint intermittently returns HTTP 500 when called back-to-back with other live tests (passes standalone). Payments aren't needed right now; re-enable and investigate if/when they are.
- `ListLessons` (Hodiny) is not covered by a live test — it returns a server-side HTTP 400 regardless of input; see `docs/plan.md`.
- Use `curl`/`bin/api` freely for read-only exploration against the live API; no need to ask permission for GET requests. Never issue mutating calls, live or in tests.
