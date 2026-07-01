# Edookit API Client — Development Plan

## Principles

* When a new fact is discovered about the API, the API docs in ./docs/api/ need to be updated.
* This plan documents needs to be kept up-to-date after each task.
* When changes are done to the TUI, integration tests need to pass. Any new feature in the TUI needs to have a test.
* Use unit tests where it makes sense.
* Do not do or test any mutating or desctructive API calls.
* Use curl and similar tools freely to test the API, do not ask for permission for GET requests.
* Use playwright any time needed, but no mutating of the state. Do not ask for permissions for navigations, screenshots, console.

## Module Coverage

Tested against `yourschool.edookit.net` on 2026-05-21.

| Module | Edookit source name | Go file | API status | TUI view | TUI status |
|--------|---------------------|---------|------------|----------|------------|
| Vyhledání osob | Vyhledávání osob | `persons.go` | ✅ 200 | `person <id>` | ✅ works |
| Změnový rozvrh | Změnový rozvrh | `schedule.go` | ✅ 200 (legacy domain) | `sc` | ✅ works |
| Veřejné události | Veřejné události | `events.go` | ✅ 200 (public, legacy domain) | `ev` | ✅ works (empty — school has no events yet) |
| Hodiny | Hodiny | `lessons.go` | ⚠️ 400 server error | `l` | ⚠️ enabled but list-lessons fails server-side |
| Osobní údaje žáků | Osobní údaje žáků | `students.go` | ✅ 200 — 85 rows | `s` | ✅ works |
| Osobní údaje zaměstnanců | Osobní údaje zaměstnanců | `students.go` | ✅ 200 — 17 rows | `emp` | ✅ works (newly added) |
| Hodnocení | Hodnocení | `grades.go` | ✅ 200 — 193 rows (student 237) | `grades <id>` | ✅ works |
| Kurzy | Kurzy | `courses.go` | ✅ 200 — 85 students | `c` | ✅ works |
| Platby | Platby | `payments.go` | ✅ 200 — 82 prescriptions | `pay` | ✅ works |
| Statistiky | Statistické informace | `stats.go` | ✅ 200 — 188 rows | `stats` | ✅ works |
| Přímý zápis docházky | Přímý zápis docházky | `attendance_direct.go` | ✅ 200 — 6 rows (student 237) | `att <id>` | ✅ works |
| Individuální cíle | Individuální cíle | `individual_goals.go` | ✅ 200 | `goals <id>` | ✅ works |
| Docházka | Docházka | — | ⛔ HMAC + vendor reg | — | blocked — needs vendor registration |

**Summary:** 11 of 13 modules working live (1 server-side error, 1 blocked on vendor reg).  
TUI covers 12 views: sc, ev, person, students, emp, grades, courses, payments, stats, att, goals, lessons (lessons has server-side 400).

## TUI Navigation

### Views and commands

| Command | View | Columns | Source |
|---------|------|---------|--------|
| `sc` | Schedule Changes | Date · Time · Original → Change · Class | `ListScheduleChanges` |
| `ev` | Public Events | Date · Time · Title · Place · Classes | `ListPublicEvents` |
| `s` / `students` | Students | ID · Name · Class · Email | `ListStudents` |
| `emp` / `employees` | Employees | ID · Name · Abbr · Class(es) · Email | `ListEmployees` |
| `person <id>` / `p <id>` | Person | ID · Name · Role · Class · Email (+ children/reps) | `SearchPerson` |
| `l` / `lessons` | Lessons | Time · Subject · Teacher · Room | `ListLessons` (⚠️ server 400) |
| `g <id>` / `grades <id>` | Grades | Course · Grade · Assignment · Evaluator · Date | `ListEvaluations` |
| `c` / `courses` | Courses | Student · Course · Attendance% · Grade | `ListCourses` |
| `pay` / `payments` | Payments | ID · Name · Due Date · Spec.Symbol · State | `ListPaymentPrescriptions` |
| `stats` | Student Stats | Class · Gender · Grade · Age · Citizenship | `ListStudentStats` |
| `att <id>` / `attendance <id>` | Attendance | Time · Course · Room · Status | `ListAttendanceLessons` |
| `goals <id>` | Individual Goals | Time · Course · Subject · Topic | `ListIndividualLessonContent` |

### Keyboard shortcuts

| Key | Context | Action |
|-----|---------|--------|
| `enter` | Students, Employees row | Open Person view for selected student/employee | ✅ |
| `g` | Students, Employees, Person | Open Grades for selected/current person | ✅ |
| `a` | Students, Employees, Person | Open Attendance for selected/current person | ✅ |
| `i` | Students, Employees, Person | Open Individual Goals for selected/current person | ✅ |
| `esc` | Any view with back-history | Return to previous view (one-level) | ✅ |
| `/` | All views | Filter current table |
| `r` | All views | Refresh (reload from API) |
| `:` | All views | Command bar (switch view) |
| `q` | All views | Quit |

### Navigation graph

```
Schedule (default)
Students ──enter──→ Person ──g──→ Grades
         ──g──────→ Grades        ──a──→ Attendance
         ──a──────→ Attendance    ──i──→ Individual Goals
         ──i──────→ Individual Goals
Employees ─── (same as Students)
Any view with history ──esc──→ previous view
```

### What is missing / TODO

- [x] Courses view: Enter on a row → Person for that student (ID stored as col 0)
- [x] Attendance: `i` → Goals for same student; `g` → Grades for same student
- [x] Goals: `a` → Attendance for same student; `g` → Grades for same student
- [x] Grades: `a` → Attendance; `i` → Goals for same student
- [x] Date navigation: `<`/`>` shifts schedule by week, events/goals by month, attendance by day
- [x] Person view: Enter on a child/rep row → Person for that person
- [x] Employees: `g`/`a`/`i` → Grades/Attendance/Goals (same as Students)

### Notes on `Hodiny` (lessons)
The module is enabled (version endpoint returns 200, rooms list works) but `/api/lesson/v2/list-lessons?date=<date>` returns HTTP 400 "An unknown error occurred" regardless of date. Likely a server-side misconfiguration or no lessons in the system for this school.

## Current Status

| Step | Status | Notes |
|------|--------|-------|
| API documentation read | ✅ Done | Scraped from uuapp.plus4u.net (requires Plus4U login) |
| Library skeleton (`types.go`, `client.go`, interfaces) | ✅ Done | |
| API method implementations | ✅ Done | All 13 modules implemented |
| TUI scaffold (k9s-style, 9 views) | ✅ Done | Runs in `--mock` mode without API |
| Live: Změnový rozvrh, Vyhledání osob, Veřejné události | ✅ Done | 3 modules confirmed working |
| Live: Osobní údaje, Hodnocení, Kurzy, Platby, Statistiky, Docházka, Individuální cíle | ✅ Done | 7 more modules enabled and verified |
| TUI: Employees view added | ✅ Done | Command: `emp` |
| Type fixes: course ID fields returned as JSON strings | ✅ Done | `json:",string"` tags on CourseLesson, StudentCourse, StudentWithCourses, CourseFinalEvaluation |
| Type fixes: Evaluation.points/bonus_points are JSON-string floats | ✅ Done | `*int` → `*float64 json:",string"` on Evaluation + CourseFinalEvaluation numeric fields |
| Integration tests | ✅ Done | `cmd/tui/tui_test.go`: all views (mock + live), navigation, date shift |
| Hodiny (list-lessons) | ⚠️ Server error | Module enabled, but list-lessons returns 400 server-side |
| Docházka (HMAC) | ⛔ Blocked | Needs vendor registration with Edookit |

## Blockers

**Docházka (HMAC)**: Requires vendor registration with Edookit — email support@edookit.com with company/system info to obtain `clientId` + `clientKey`.

**All other 401 modules**: See coverage table above — admin must create per-module credentials in Edookit.

## Implementation Steps

### Phase 1 — Library (current)
- [x] `client.go` — HTTP Basic auth transport, error handling
- [x] `types.go` — all response/request structs
- [x] `source.go` — `DataSource` interface (real + mock can implement it)
- [x] `persons.go` — `GET /api/person-search/v1/{criterion}`
- [x] `lessons.go` — `GET /api/lesson/v2/list-lessons` + support endpoints
- [x] `students.go` — `GET /api/student-data/v1/list`, `GET /api/employee-data/v1/list`
- [x] `grades.go` — `GET /api/evaluation/v1/list`
- [x] `courses.go` — `GET /api/course-data/v1/courses`
- [x] `payments.go` — full CRUD `/api/payment/v1/`
- [x] `stats.go` — `GET /api/stats/v1/students`

### Phase 2 — TUI (current)
- [x] k9s-style layout (header / table / hints bar)
- [x] 5 views: students, lessons, grades, courses, payments
- [x] `:` command bar for switching views
- [x] `/` filter for current view
- [x] Mock data mode (runs without API)
- [ ] Wire views to real `*Client` when credentials available
- [ ] Detail view on `<enter>`
- [ ] Date navigation in lessons view

### Phase 3 — Polish
- [ ] Config file (`~/.edookit.yaml` or `.env`)
- [ ] Multiple school instances
- [ ] Attendance module (needs vendor HMAC registration)
- [ ] Export to CSV / JSON

## Architecture

```
edookit/
├── go.mod
├── client.go          HTTP Basic transport
├── source.go          DataSource interface
├── types.go           All structs
├── persons.go         Person search
├── lessons.go         Timetable
├── students.go        Student / employee personal data
├── grades.go          Evaluations / grades
├── courses.go         Course enrollment
├── payments.go        Payment prescriptions + payments (CRUD)
├── stats.go           Anonymised statistics
├── docs/
│   ├── plan.md        This file
│   └── api/           API documentation per module
└── cmd/
    └── tui/
        ├── main.go    Entry point (flags, config)
        ├── model.go   bubbletea root model
        ├── mock.go    MockSource with Czech school demo data
        └── styles.go  lipgloss purple theme
```

## Environment Variables (TUI)

```
EDOOKIT_URL       Base URL, e.g. https://yourschool.edookit.net
EDOOKIT_USER      API username (from Edookit admin)
EDOOKIT_PASS      API password (from Edookit admin)
```

If `EDOOKIT_URL` is unset or `--mock` flag is passed, the TUI runs with mock data.
