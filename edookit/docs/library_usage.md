# Using `edookit` as a Go Library

The root package (`client.go`, `students.go`, `types.go`, ...) has no TUI
dependencies — only stdlib. It can be imported directly into another Go
project without pulling in `bubbletea`/`bubbles`/`lipgloss` (those are only
used by `cmd/tui`).

## Module path

```
github.com/martinpovolny/go-snip/edookit
```

This is a subdirectory module inside the `go-snip` monorepo, not a
standalone repo. It must be committed and pushed to `origin` (github.com/martinpovolny/go-snip)
before another project can `go get` it.

## Import

```bash
go get github.com/martinpovolny/go-snip/edookit
```

```go
import edookit "github.com/martinpovolny/go-snip/edookit"

client := edookit.New(
    os.Getenv("EDOOKIT_URL"),
    os.Getenv("EDOOKIT_USER"),
    os.Getenv("EDOOKIT_PASS"),
)

students, err := client.ListStudents(edookit.StudentDataOpts{})
```

## Syncing a subset of fields

Keep field projection in the *consuming* project, not in `edookit`. The
library returns the full `Student`/`Employee`/etc. structs; define your own
narrow struct on the consumer side and map into it:

```go
type StudentSummary struct {
    ID    int
    Name  string
    Class string
}

func toSummary(s edookit.Student) StudentSummary {
    var class string
    if s.ClassName != nil {
        class = *s.ClassName
    }
    return StudentSummary{
        ID:    s.PersonID,
        Name:  s.Firstname + " " + s.Lastname,
        Class: class,
    }
}
```

This keeps `edookit` a plain data-access library. When you later need more
entities (grades, attendance, individual goals, ...), call the matching
`edookit.Client` method and add another projection — no changes needed to
`edookit` itself.

## Available list/search methods

| Method | Entity |
|--------|--------|
| `ListStudents` | Students |
| `ListEmployees` | Teachers/staff |
| `SearchPerson` / `GetPerson` | Any person by Edookit or Plus4U ID |
| `ListCourses` | Course enrollment |
| `ListEvaluations` | Grades |
| `ListPaymentPrescriptions` | Payments |
| `ListStudentStats` | Anonymised statistics |
| `ListAttendanceLessons` | Direct attendance record |
| `ListIndividualLessonContent` | Individual goals |
| `ListScheduleChanges` | Schedule changes (legacy domain) |
| `ListPublicEvents` | Public events (legacy domain, no auth) |

See `docs/plan.md` for live API status per module and `docs/api/` for
per-module request/response details.
