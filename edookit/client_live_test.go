//go:build live

// Live API tests for the edookit client library. These make real, read-only
// requests against a live Edookit school instance — no destructive or
// mutating calls (no CreatePayment, CreatePaymentPrescription, etc).
//
// Run with:
//
//	go test -tags live .
//
// Requires API_URL / API_USER / API_PASSWORD, loaded from .env in this
// directory (or already exported in the environment).
package edookit

import (
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	_ = godotenv.Load()
	os.Exit(m.Run())
}

func liveClient(t *testing.T) *Client {
	t.Helper()
	apiURL := os.Getenv("API_URL")
	user := os.Getenv("API_USER")
	pass := os.Getenv("API_PASSWORD")
	if apiURL == "" || user == "" || pass == "" {
		t.Skip("API_URL / API_USER / API_PASSWORD not set — skipping live API test")
	}
	return New(apiURL, user, pass)
}

func TestLiveListStudents(t *testing.T) {
	c := liveClient(t)
	students, err := c.ListStudents(StudentDataOpts{})
	if err != nil {
		t.Fatalf("ListStudents: %v", err)
	}
	t.Logf("%d students", len(students))
}

func TestLiveListEmployees(t *testing.T) {
	c := liveClient(t)
	employees, err := c.ListEmployees(StudentDataOpts{})
	if err != nil {
		t.Fatalf("ListEmployees: %v", err)
	}
	t.Logf("%d employees", len(employees))
}

func TestLiveSearchPerson(t *testing.T) {
	c := liveClient(t)
	people, err := c.GetPerson(237)
	if err != nil {
		t.Fatalf("GetPerson: %v", err)
	}
	t.Logf("%d person records", len(people))
}

func TestLiveListCourses(t *testing.T) {
	c := liveClient(t)
	courses, err := c.ListCourses(CourseListOpts{})
	if err != nil {
		t.Fatalf("ListCourses: %v", err)
	}
	t.Logf("%d students with courses", len(courses))
}

func TestLiveListEvaluations(t *testing.T) {
	c := liveClient(t)
	studentID := 237
	evals, err := c.ListEvaluations(EvalListOpts{StudentID: &studentID})
	if err != nil {
		t.Fatalf("ListEvaluations: %v", err)
	}
	t.Logf("%d evaluations for student %d", len(evals), studentID)
}

func TestLiveListPaymentPrescriptions(t *testing.T) {
	t.Skip("payments not needed right now; endpoint intermittently returns HTTP 500 when called back-to-back with other live tests (passes standalone) — re-enable if/when payments are needed")
	c := liveClient(t)
	prescriptions, err := c.ListPaymentPrescriptions(PrescriptionListOpts{})
	if err != nil {
		t.Fatalf("ListPaymentPrescriptions: %v", err)
	}
	t.Logf("%d payment prescriptions", len(prescriptions))
}

func TestLiveListStudentStats(t *testing.T) {
	c := liveClient(t)
	stats, err := c.ListStudentStats(time.Now().Format("2006-01-02"))
	if err != nil {
		t.Fatalf("ListStudentStats: %v", err)
	}
	t.Logf("%d stat rows", len(stats))
}

func TestLiveListAttendanceLessons(t *testing.T) {
	c := liveClient(t)
	lessons, err := c.ListAttendanceLessons(237, time.Now().Format("2006-01-02"))
	if err != nil {
		t.Fatalf("ListAttendanceLessons: %v", err)
	}
	t.Logf("%d attendance lessons", len(lessons))
}

func TestLiveListIndividualLessonContent(t *testing.T) {
	c := liveClient(t)
	now := time.Now()
	monthFrom := now.AddDate(0, 0, -30).Format("2006-01-02")
	monthTo := now.Format("2006-01-02")
	content, err := c.ListIndividualLessonContent(237, 0, monthFrom, monthTo)
	if err != nil {
		t.Fatalf("ListIndividualLessonContent: %v", err)
	}
	t.Logf("%d individual lesson content rows", len(content))
}

func TestLiveListScheduleChanges(t *testing.T) {
	c := liveClient(t)
	now := time.Now()
	changes, err := c.ListScheduleChanges(ScheduleChangeOpts{
		From: now.Format("2006-01-02"),
		To:   now.AddDate(0, 0, 6).Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("ListScheduleChanges: %v", err)
	}
	t.Logf("%d schedule changes", len(changes))
}

func TestLiveListPublicEvents(t *testing.T) {
	c := liveClient(t)
	now := time.Now()
	events, err := c.ListPublicEvents(PublicEventsOpts{
		From: now.AddDate(0, -1, 0).Format("2006-01-02"),
		To:   now.AddDate(0, 1, 0).Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("ListPublicEvents: %v", err)
	}
	t.Logf("%d public events", len(events))
}

// ListLessons is known to return HTTP 400 server-side regardless of input —
// see docs/plan.md ("Hodiny"). Not covered here until that's resolved.
