# SPA Internal API — Edookit Parent Portal

Investigated 2026-05-21 against `yourschool.edookit.net` (parent account).

## Architecture

This is **not a JSON API app**. It is a **Nette Framework PHP application** that
server-renders full HTML pages. Data arrives embedded in the initial HTML response.
JavaScript (jQuery + nette.ajax) is used only for lightweight interactions
(term switching, person filter) and form submissions.

## Authentication

Cookie-based PHP session. Log in via `/user/login` with Edookit username/password.
The session cookie is sent automatically with all requests — no API key needed for
scraping, just a valid logged-in browser session.

The OIDC/Plus4U integration (`/oidc/introspect`) runs on page load to validate the
session. No Bearer token is needed for the HTML pages.

## How to Fetch Data

```
GET /<route>
Cookie: <session cookie from browser>
```

Returns a full HTML page. The main content is inside `<div id="content">`.

For AJAX interactions, send:
```
GET|POST /<route>?<params>&do=<signal-name>
X-Requested-With: XMLHttpRequest
Cookie: <session cookie>
```

Returns JSON:
- `{"redirect": "https://...", "state": []}` — full page reload needed
- `{"snippets": {"snippet--foo": "<html fragment>"}, "state": []}` — partial update

## Base URL

`https://yourschool.edookit.net` (parent/student portal)

The teacher/admin portal lives on `https://yourschool-login.edookit.net` and
redirects parent accounts back to the main portal. Teacher data (entering grades,
creating lessons) is only accessible there with teacher credentials.

## Routes and Data Available

### Dashboard
```
GET /
```
Summary of recent activity.

### Grades

```
GET /evaluation/               # Grades grouped by subject, all terms side-by-side
GET /evaluation/listing        # Grade listing sorted by date (newest first)
GET /evaluation/overall        # Final/summary grades per term
GET /evaluation/trends         # Grade trend charts
GET /evaluation/records        # Behaviour records
```

**Example data** (`/evaluation/`):
```
Anglický jazyk Aj - 3.
  Nováková Zdislava
  Docházka: 2. pololetí 25/26: 97%, 1. pololetí 25/26: 83%
  2. pololetí 25/26: 1, 1-, 1, 2, 1, 1, 1, 2, 1, 1, 1, 2
  1. pololetí 25/26: 1, 1, 1, 1-, 1, nez., 2, 1-, 1, 1, 1-
```

### Courses (Předměty)

```
GET /courses/                  # All enrolled courses, current term
GET /courses/enroll            # Course enrollment form
GET /groups/                   # Class groups
```

**Example data** (`/courses/`):
```
Anglický jazyk Aj - 3.  2025/26
  Učitel: Nováková Zdislava
  Známky: 1, 1-, 1, 2, 1, 1, 1, 2, 1, 2, 1, 1
  Docházka: 1 absence (97%)
  Učivo: I'm wearing.... PB str. 105/3 WB str. 91
  Dom. úkoly: žádné nadcházející
```

### Timetable (Rozvrh)

```
GET /timetable/                # Weekly timetable grid (current week by default)
GET /timetable/upcoming        # Upcoming school events
GET /timetable/archive         # Past school events
```

**Example data** (`/timetable/`):
```
Týden od 18. 5. do 22. 5.
Štěpánka Povolná  3.

Po 18.5 | Út 19.5 | St 20.5 | Čt 21.5 | Pá 22.5
Period 1 08:00–08:45: ŠD/ML/Učebna 2 | Čj/ZN/Učebna 4 | Čj/ZN/Učebna 4 | Čj/ZN/Učebna 4 | Čj/ZN/Projektový den - animace
...
```

**Week navigation** via query param:
```
GET /timetable/?timetable-weekSelector-week=2026-05-18
```

**Example data** (`/timetable/upcoming`):
```
22. 5., od 8:00 do 11:40 — Projektový den - animace  (created by Šárka Povolná, 20.5.2026 9:53)
```

### Attendance (Docházka)

```
GET /attendance/               # Retroactive absence excusing (write-capable form)
GET /attendance/stats          # Attendance summary per term
GET /attendance/family-overview
GET /attendance/report         # Detailed attendance report
GET /attendance/report-lesson  # Per-lesson attendance report
GET /attendance/excused        # Excused absences
GET /attendance/reserve-schedule-cards  # Reserve schedule cards
```

**Example data** (`/attendance/stats`):
```
2. pololetí 25/26:
  Absence omluvená: 6 hodin
  Školní akce:      68 hodin
1. pololetí 25/26:
  Absence omluvená: 71 hodin
  Školní akce:      86 hodin
```

### Homework & Exams

```
GET /assignments/              # Upcoming homework (grouped by subject)
GET /assignments/archive       # Past homework
GET /exams/                    # Upcoming written tests / oral exams
GET /exams/archive             # Past exams
```

**Example data** (`/assignments/`):
```
Nedokončené (2)
Anglický jazyk (1), Člověk a jeho svět (1), Český jazyk (3), Matematika (1)
```

### Payments (Platby)

```
GET /payments/                 # Payment history (school fees, trips, etc.)
```

### Messages (Schránka)

```
GET /overview/updates          # Received messages
GET /overview/sent             # Sent messages
GET /overview/archive          # Archived messages
GET /messages/new              # Compose new message
GET /messages/new?recipients=<personId>[;<personId>]   # Pre-filled recipients
GET /messages/addressees       # Address book
GET /messages/tech-support     # Message to Edookit support
```

### Learning Content

```
GET /lesson-plans/             # Lesson content / curriculum (učivo v hodinách)
GET /lesson-plans/thematic-plans  # Thematic plans
GET /activity/                 # In-class activity records
GET /materials/                # Educational materials
GET /portfolios/               # Portfolio of student work
```

### Other

```
GET /stream/                   # Social stream
GET /discussions/              # Discussions
GET /consents/                 # Consents (current)
GET /consents/archive          # Archived consents
GET /meal-orders/              # Meal orders
GET /meal-orders/archive       # Past meal orders
GET /achievements/             # Edookit badges
GET /course-kit/my-courses     # External (Red Monster) courses
GET /settings/preferences      # App preferences
GET /settings/children         # Child account settings
GET /settings/idm              # Login / Plus4U pairing
GET /settings/data-check       # Personal data verification
GET /user/logout               # Logout
```

## Nette Signals (?do=...)

These are AJAX handlers. Call with `X-Requested-With: XMLHttpRequest` to get JSON.

| Signal | Method | Params | Effect |
|--------|--------|--------|--------|
| `mainMenu-termSelector-change` | GET | `mainMenu-termSelector-new=<termId>` | Switch school year/term |
| `header-changeQuickPersonFilter` | POST | (form body with person selector) | Switch active child (multi-child parents) |
| `mainMenu-searchBar-searchBar` | POST | Form: search input | Person search |
| `header-searchBar-searchBar` | POST | Form: search input | Global search |

## Term IDs Observed

IDs from oldest → newest (highest = most recent):

`18, 23, 26, 29, 32, 35, 38, 41, 44, 47, 50`

Current active terms:
- `50` → 2026/27
- `47` → 2025/26  
- `44` → 2024/25

## One Real JSON API Endpoint

```
POST /api/notification/v1/register
```
Used internally by the SPA for push notification (Firebase) registration.
Not useful for data extraction.

## Scraping Strategy

Since the public REST API (`/api/...`) requires per-module credentials that may not
be provisioned, the SPA pages can be scraped as a fallback:

1. **Authenticate**: `POST /user/login` with credentials, capture session cookie
2. **Fetch page**: `GET /evaluation/` etc. with `Cookie` header
3. **Parse**: extract `<div id="content">`, parse with an HTML parser (e.g. `golang.org/x/net/html`)
4. **Term switching**: `GET /evaluation/?mainMenu-termSelector-new=47&do=mainMenu-termSelector-change` then reload

```go
// Rough sketch
client := &http.Client{Jar: jar}  // jar holds session cookies

// Login
client.PostForm("https://yourschool.edookit.net/user/login", url.Values{
    "username": {"api_user"}, "password": {"..."}, "do": {"login-form-submit"},
})

// Fetch grades
resp, _ := client.Get("https://yourschool.edookit.net/evaluation/")
// parse resp.Body with golang.org/x/net/html, find #content
```

## Limitations

- **Parent account only**: this account is a parent; teacher/admin routes on
  `yourschool-login.edookit.net` redirect back to the parent portal.
- **No bulk/structured export**: all data is in human-readable HTML tables, not JSON.
- **Session expiry**: cookies expire; need re-authentication.
- **CSRF protection**: write operations (POST forms) require Nette CSRF tokens
  embedded in the HTML form. Read-only scraping is safe.
