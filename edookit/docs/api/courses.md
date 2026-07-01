# Courses

**Source name**: Úd. o kurzech  
**Auth**: HTTP Basic  
**Base path**: `/api/course-data/v1/`

## Endpoints

### Get API Version

```
GET /api/course-data/v1/version
```

Returns plain text, e.g. `1.0.0`.

---

### List Students with Courses

```
GET /api/course-data/v1/courses
```

Returns students with their enrolled courses for a given evaluation term, including lesson-level attendance and final grades.

#### Query Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `person_id` | int | no | Return courses for one student. Omit for all students. |
| `evalterm_id` | int | no | Evaluation term. Defaults to current term. |

#### Response Fields

Returns `{"students": [...]}`.

> **Wire type gotcha**: All `id` fields in this API are returned as JSON **strings** (e.g. `"id": "341"`), not integers, despite the logical type being int. Apply `json:",string"` tags in Go structs.

**Student fields**:

| Field | Type | Nullable | Description |
|-------|------|----------|-------------|
| `id` | **string (int)** | no | Student ID — JSON string |
| `first_name` | string | no | First name |
| `last_name` | string | no | Last name |
| `middle_name` | string | yes | Middle name |
| `courses` | array | no | Enrolled courses (may be empty) |

**Course fields**:

| Field | Type | Nullable | Description |
|-------|------|----------|-------------|
| `id` | **string (int)** | no | Course ID — JSON string |
| `code` | string | no | Course code / name |
| `lessons` | array | no | Lessons with attendance status |
| `final_evaluations` | array | no | Final grades for this term |
| `attendance_percentage` | string | yes | e.g. `"85.00"` |

**Lesson fields**: `id` (string), `timerange` (PostgreSQL range format), `attendance_status` (e.g. `"P"` present, `"A"` absent)

**Final evaluation fields** (`final_evaluations` array):

| Field | Type | Nullable | Description |
|-------|------|----------|-------------|
| `id` | **string (int)** | no | Evaluation ID — JSON string |
| `evaluation_term_id` | **string (int)** | no | Term ID — JSON string |
| `evaluation_term_name` | string | no | e.g. `"1. pololetí 25/26"` |
| `grade` | string | yes | Grade value (e.g. `"1"`, `"2"`) |
| `points` | **string (float)** | yes | Points — JSON string when non-null |
| `extrapoints` | **string (float)** | yes | Extra points — JSON string when non-null |
| `percentages` | **string (float)** | yes | Percentages — JSON string when non-null |
| `verbal_evaluation` | string | yes | Verbal comment |
| `weight` | string (float) | yes | Grade weight |
| `resit_evaluation_id` | **string (int)** | yes | Resit evaluation ID — JSON string |
| `included_in_overall_evaluation` | bool | no | Whether included in overall grade |

#### Example Response

```json
{
  "students": [
    {
      "id": 341,
      "first_name": "Jana",
      "last_name": "Anažková",
      "courses": [
        {
          "id": 590,
          "code": "Tv - 9.A",
          "lessons": [
            {"id": 48322, "timerange": "[\"2023-01-04 14:25:00\",\"2023-01-04 15:10:00\")", "attendance_status": "P"}
          ],
          "final_evaluations": [
            {"id": 18363, "grade": "1", "evaluation_term_name": "1. pololetí 22/23", "included_in_overall_evaluation": true}
          ],
          "attendance_percentage": "84.12"
        }
      ]
    }
  ]
}
```
