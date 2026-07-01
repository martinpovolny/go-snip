# Přímý zápis docházky (Direct Attendance Entry)

## Configuration

| Field | Value |
|-------|-------|
| Source name | Přímý zápis docházky |
| Domain | `https://<school>.edookit.net` |
| Auth | HTTP Basic |

## Endpoints

### List lessons for a student — read-only

```
GET /api/attendance/v3/lessons
```

Returns lessons for a student on a given day with their current attendance status.

**Parameters** (query string or JSON body)

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `student_id` | int | yes | Edookit person ID of the student |
| `date` | date | no | YYYY-MM-DD; defaults to today |

**Response (200)** — array

```json
[
  {
    "person_id": "4541",
    "lesson_id": "1151",
    "timerange": "[\"2025-10-01 07:50:00\",\"2025-10-01 08:35:00\")",
    "room_ids": [8],
    "course_ids": [1225],
    "room_names": ["T-U132"],
    "course_names": ["ZEL - M1.A"],
    "attendance_status": null,
    "minutes_late": null,
    "minutes_early": null
  }
]
```

`timerange` uses PostgreSQL range notation `["start","end")`.

### Submit future absence notice — **write**

```
POST /api/attendance/v3/futureExcuse
```

| Param | Type | Description |
|-------|------|-------------|
| `student_id` | int | Student Edookit ID |
| `lesson_id` | int | Lesson ID |
| `excuse` | string | Reason |

### Submit current excuse — **write**

```
POST /api/attendance/v3/currentExcuse
```

Same parameters as `futureExcuse`.

### Record attendance — **write**

```
POST /api/attendance/v3/record
```

| Param | Type | Description |
|-------|------|-------------|
| `student_id` | int | Student Edookit ID |
| `lesson_id` | int | Lesson ID |
| `minutes_late` | int | Minutes late |
| `minutes_early` | int | Minutes early departure |
| `status` | string | Attendance status code (omit = highest-priority present state) |

## Go usage

```go
c := edookit.New("https://yourschool.edookit.net", user, pass)

// Read-only
lessons, err := c.ListAttendanceLessons(237, "2026-05-21")
```

Write operations (`futureExcuse`, `currentExcuse`, `record`) are not implemented
in the Go library to prevent accidental data mutation.

## curl

```bash
bin/api '/api/attendance/v3/lessons?student_id=237&date=2026-05-21'
```
