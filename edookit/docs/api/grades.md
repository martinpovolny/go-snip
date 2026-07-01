# Grades / Evaluation

**Source name**: Hodnocení  
**Auth**: HTTP Basic  
**Base path**: `/api/evaluation/v1/`

Provides both ongoing and final evaluation data.

## Endpoint

### List Evaluations

```
GET /api/evaluation/v1/list
```

Parameters can be sent as query string or JSON body. At least one of `student_id`, `course_id`, or `evalterm_id` must be provided.

#### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `student_id` | int | cond. | Filter by student |
| `course_id` | int | cond. | Filter by course |
| `evalterm_id` | int | cond. | Filter by evaluation term |
| `date_from` | date | no | Include results from this date |
| `date_to` | date | no | Include results up to this date |

#### Response Fields

Returns a JSON array of evaluation objects.

> **Note**: Several ID fields are returned as JSON strings (e.g. `"evaluation_id": "112"`).

| Field | Type | Nullable | Description |
|-------|------|----------|-------------|
| `evaluation_id` | string (int) | no | Evaluation ID |
| `student_id` | string (int) | no | Student ID |
| `course_id` | string (int) | no | Course ID |
| `course_code` | string | no | Course code / name |
| `achievement_level` | string | yes | Grade (e.g. `"1"`, `"2"`) |
| `percentage` | **string (float)** | yes | Percentage — encoded as JSON string when non-null |
| `points` | **string (float)** | yes | Points — encoded as JSON string; can be decimal (e.g. `"7.5"`) |
| `bonus_points` | **string (float)** | yes | Bonus points — encoded as JSON string when non-null |
| `verbal_evaluation` | string | yes | Verbal comment |
| `assignment_id` | string (int) | yes | Assignment ID |
| `assignment_name` | string | yes | Assignment name |
| `evaluator_id` | string (int) | yes | Evaluator (teacher) ID |
| `evaluator_full_name` | string | yes | Evaluator full name |
| `evaluation_date` | date | yes | Date of evaluation |
| `evalterm_id` | string (int) | yes | Evaluation term ID |
| `evalterm_name` | string | yes | Evaluation term name (e.g. "2. pololetí 24/25") |
| `not_evaluated` | bool | no | `true` if student was not evaluated |
| `no_eval_reason_id` | string (int) | yes | Reason code for not being evaluated |
| `resit_evaluation_id` | string (int) | yes | ID of resit evaluation — encoded as JSON string |
| `weight` | string (float) | yes | Grade weight |

#### Example Response

```json
[
  {
    "evaluation_id": "112",
    "student_id": "221",
    "course_id": "5644",
    "course_code": "M - 2.A",
    "achievement_level": "2",
    "evaluator_full_name": "David Dvořák",
    "evaluation_date": "2025-06-22",
    "evalterm_name": "2. pololetí 24/25",
    "not_evaluated": false,
    "weight": "0.00"
  }
]
```
