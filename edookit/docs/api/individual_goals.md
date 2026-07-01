# Individuální cíle (Individual Goals)

## Configuration

| Field | Value |
|-------|-------|
| Source name | Individuální cíle |
| Domain | `https://<school>.edookit.net` |
| Auth | HTTP Basic |

## Endpoints

### Individual lesson content

```
GET /api/individual-goals/v1/individual-lesson-content
```

Returns per-lesson individual curriculum for students. At least one of `student_id`,
`lesson_id`, or `date_from` is required.

**Parameters** (query string or JSON body)

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `student_id` | int | cond. | Filter by student |
| `lesson_id` | int | cond. | Filter by lesson ID (cannot combine with `date_from`/`date_to`) |
| `date_from` | date | cond. | Lessons from this date inclusive |
| `date_to` | date | no | Lessons up to this date inclusive |

**Response (200)** — array of students, each with a `lessons` array

```json
[
  {
    "student_id": 1234,
    "student_identifier": "436bb83d-ffb1-438d-942b-5b4fab76ee72",
    "lessons": [
      {
        "lesson_id": "123456",
        "actual_timerange": "[\"2020-01-01 08:30:00\",\"2020-01-01 08:45:00\")",
        "student_lesson_topic": "Did exercise 56, 57",
        "student_planned_lesson_topic": "Do exercise 56, 57, 58",
        "student_lesson_note": null,
        "subject_name": "Math",
        "course_code": "M - 3.A"
      }
    ]
  }
]
```

`actual_timerange` is `null` for cancelled lessons.

### Learning agreements (výukové plány)

```
GET /api/individual-goals/v1/learning-agreements
```

Returns student learning agreements.

**Parameters**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `student_id` | int | no | Filter by student |
| `date_from` | date | no | Plans valid from this date |
| `date_to` | date | no | Plans valid up to this date |

**Response** — array of students, each with a `learning_agreements` array containing
feedback fields (`student_feedback_positive/negative/other`, `coach_feedback_*`),
`next_goals`, `next_agreement_date`, `coaches`, and `target_subcompetences`.

## Go usage

```go
c := edookit.New("https://yourschool.edookit.net", user, pass)

content, err := c.ListIndividualLessonContent(237, 0, "2026-05-01", "2026-05-31")
agreements, err := c.ListLearningAgreements(237, "2026-01-01", "")
```

## curl

```bash
bin/api '/api/individual-goals/v1/individual-lesson-content?student_id=237&date_from=2026-05-01'
bin/api '/api/individual-goals/v1/learning-agreements?student_id=237'
```
