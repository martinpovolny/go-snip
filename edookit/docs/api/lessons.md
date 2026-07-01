# Lessons / Timetable

**Source name**: Hodiny  
**Auth**: HTTP Basic  
**Base path**: `/api/lesson/v2/`

## Endpoints

### Get API Version

```
GET /api/lesson/v2/version
```

Returns a plain-text string, e.g. `2.0.0`.

---

### List Lessons

```
GET /api/lesson/v2/list-lessons
```

Returns the timetable for a given day. Each lesson has an `actual` state (reflecting substitutions) and a `scheduled` state (original timetable).

#### Query Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `date` | string | yes | Date in `YYYY-MM-DD` format |
| `course_id` | int | no | Filter by course ID |
| `course_type_id` | int | no | Filter by course type ID |
| `room_id` | int | no | Filter by room ID |
| `student_person_id` | int | no | Filter by student ID |
| `teacher_person_id` | int | no | Filter by teacher ID |
| `work_type_id` | int | no | Filter by work type ID |

#### Response

Returns `{"lessons": {"<lesson_id>": {...}}}`. If no lessons found, `lessons` is an empty object.

**Lesson entry fields** (same structure for `actual` and `scheduled`):

| Field | Type | Nullable | Description |
|-------|------|----------|-------------|
| `lesson_id` | int | no | Unique lesson ID |
| `datetime_from` | string | no | Start: `YYYY-MM-DD HH:MM:SS` |
| `datetime_to` | string | no | End: `YYYY-MM-DD HH:MM:SS` |
| `lessonjoin_id` | int | yes | Links joined lessons; null in `scheduled` |
| `name` | string | no | Lesson name (course codes, comma-separated) |
| `courses` | map[id→CourseRef] | no | Courses in this lesson |
| `teachers` | map[id→TeacherRef] | no | Teachers |
| `students` | map[id→StudentGroupRef] | no | Student groups |
| `rooms` | map[id→RoomRef] | no | Rooms |

**CourseRef**: `course_id`, `course_code`, `course_type_id`  
**TeacherRef**: `person_id`, `person_abbr`  
**StudentGroupRef**: `subject_id`, `subject_name`  
**RoomRef**: `room_id`, `room_name`

#### Example Response

```json
{
  "lessons": {
    "45456": {
      "actual": {
        "lesson_id": 45456,
        "datetime_from": "2023-03-22 10:40:00",
        "datetime_to": "2023-03-22 11:25:00",
        "lessonjoin_id": null,
        "name": "ČJ - 5.A",
        "courses": {"5646": {"course_id": 5646, "course_code": "ČJ - 5.A", "course_type_id": 23}},
        "teachers": {"345": {"person_id": 345, "person_abbr": "NOV"}},
        "students": {"4587": {"subject_id": 4587, "subject_name": "Žáci 5.A"}},
        "rooms": {"655": {"room_id": 655, "room_name": "M305"}}
      },
      "scheduled": { "...": "same structure" }
    }
  }
}
```

---

### List Rooms

```
GET /api/lesson/v2/lists/rooms
```

Returns `{"rooms": {"<room_id>": {"room_id": int, "name": string}}}`.

---

### List Course Types

```
GET /api/lesson/v2/lists/types/course
```

Returns `{"course_types": {"<id>": {"course_type_id": int, "name": string}}}`.

---

### List Work Types

```
GET /api/lesson/v2/lists/types/work
```

Returns `{"work_types": {"<id>": {"worktype_id": int, "name": string}}}`.
