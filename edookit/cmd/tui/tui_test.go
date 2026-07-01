package main

import (
	"os"
	"testing"

	edookit "github.com/martinpovolny/go-snip/edookit"
)

// ── Mock integration: every view, no live API ───────────────────────────────

func TestAllViewsMock(t *testing.T) {
	src := &MockSource{}
	m := newModel(src, "https://mock.test", true)

	tests := []struct {
		view    viewKind
		setup   func()
		minRows int
	}{
		{viewSchedule, nil, 1},
		{viewEvents, nil, 1},
		{viewStudents, nil, 1},
		{viewEmployees, nil, 1},
		{viewCourses, nil, 1},
		{viewPayments, nil, 1},
		{viewStats, nil, 1},
		{viewLessons, nil, 0}, // lessons mock may return 0
		{viewPerson, func() { m.personID = "237" }, 1},
		{viewGrades, func() { m.gradeStudentID = "237" }, 1},
		{viewAttendance, func() { m.attStudentID = "237" }, 1},
		{viewGoals, func() { m.goalsStudentID = "237" }, 1},
	}

	for _, tt := range tests {
		t.Run(tt.view.String(), func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			m.view = tt.view
			rows, err := m.fetchRows()
			if err != nil {
				t.Fatalf("fetchRows error: %v", err)
			}
			if len(rows) < tt.minRows {
				t.Errorf("want >= %d rows, got %d", tt.minRows, len(rows))
			}
		})
	}
}

// ── Navigation logic (mock) ─────────────────────────────────────────────────

func TestNavigationMock(t *testing.T) {
	src := &MockSource{}
	m := newModel(src, "https://mock.test", true)

	// Load students and build table so SelectedRow() returns something.
	m.view = viewStudents
	rows, err := m.fetchRows()
	if err != nil {
		t.Fatalf("load students: %v", err)
	}
	m.allRows = rows
	m.table = m.buildTable(rows)

	// Enter → Person
	cmd := m.handleEnter()
	if cmd == nil {
		t.Fatal("handleEnter on students: want load command, got nil")
	}
	if m.view != viewPerson {
		t.Errorf("after enter: want viewPerson, got %v", m.view)
	}
	if m.personID == "" {
		t.Error("personID should be set after enter from students")
	}
	if !m.hasPrev || m.prevView != viewStudents {
		t.Errorf("want hasPrev=true prevView=viewStudents, got hasPrev=%v prevView=%v", m.hasPrev, m.prevView)
	}

	// g → Grades from Person
	m.gradeStudentID = ""
	cmd = m.navigateTo(viewGrades)
	if cmd == nil {
		t.Fatal("navigateTo(grades) from person: want command, got nil")
	}
	if m.view != viewGrades {
		t.Errorf("after g: want viewGrades, got %v", m.view)
	}
	if m.gradeStudentID == "" {
		t.Error("gradeStudentID should be set after navigateTo(grades)")
	}

	// Esc → back to Person
	cmd = m.goBack()
	if cmd == nil {
		t.Fatal("goBack: want command, got nil")
	}
	if m.view != viewPerson {
		t.Errorf("after esc: want viewPerson, got %v", m.view)
	}
	if m.hasPrev {
		t.Error("hasPrev should be false after goBack")
	}

	// Courses Enter → Person
	m.view = viewCourses
	crows, err := m.fetchRows()
	if err != nil {
		t.Fatalf("load courses: %v", err)
	}
	m.allRows = crows
	m.table = m.buildTable(crows)
	m.personID = ""
	cmd = m.handleEnter()
	if cmd == nil {
		t.Fatal("handleEnter on courses: want command, got nil")
	}
	if m.view != viewPerson {
		t.Errorf("courses enter: want viewPerson, got %v", m.view)
	}
	if m.personID == "" {
		t.Error("personID should be set after enter from courses")
	}

	// Attendance → Goals via i
	m.attStudentID = "237"
	m.view = viewAttendance
	cmd = m.navigateTo(viewGoals)
	if cmd == nil {
		t.Fatal("navigateTo(goals) from attendance: want command, got nil")
	}
	if m.view != viewGoals {
		t.Errorf("att→goals: want viewGoals, got %v", m.view)
	}
	if m.goalsStudentID != "237" {
		t.Errorf("goalsStudentID: want 237, got %q", m.goalsStudentID)
	}
}

// ── Date shift ──────────────────────────────────────────────────────────────

func TestDateShift(t *testing.T) {
	src := &MockSource{}
	m := newModel(src, "https://mock.test", true)

	origWeek := m.weekFrom
	m.view = viewSchedule
	m.shiftDate(1)
	if m.weekFrom == origWeek {
		t.Error("weekFrom unchanged after shiftDate(+1) in schedule")
	}
	m.shiftDate(-1)
	if m.weekFrom != origWeek {
		t.Error("weekFrom should return to original after +1 -1")
	}

	origMonth := m.monthFrom
	m.view = viewGoals
	m.shiftDate(1)
	if m.monthFrom == origMonth {
		t.Error("monthFrom unchanged after shiftDate(+1) in goals")
	}
	m.shiftDate(-1)
	if m.monthFrom != origMonth {
		t.Error("monthFrom should return to original after +1 -1 in goals")
	}

	origAtt := m.attDate
	m.view = viewAttendance
	m.shiftDate(-1)
	if m.attDate == origAtt {
		t.Error("attDate unchanged after shiftDate(-1) in attendance")
	}
}

// ── Live API: all views + type checks ──────────────────────────────────────

func TestLiveAllViews(t *testing.T) {
	apiURL := os.Getenv("API_URL")
	user := os.Getenv("API_USER")
	pass := os.Getenv("API_PASSWORD")
	if apiURL == "" || user == "" || pass == "" {
		t.Skip("API_URL / API_USER / API_PASSWORD not set — skipping live tests")
	}

	c := edookit.New(apiURL, user, pass)
	m := newModel(c, apiURL, false)

	// Views that need no student ID
	for _, v := range []viewKind{viewSchedule, viewEvents, viewStudents, viewEmployees, viewCourses, viewPayments, viewStats} {
		v := v
		t.Run(v.String(), func(t *testing.T) {
			m.view = v
			rows, err := m.fetchRows()
			if err != nil {
				t.Fatalf("fetchRows: %v", err)
			}
			t.Logf("%d rows", len(rows))
		})
	}

	// Student-specific views — also exercises float64 points fix (student 86)
	for _, tc := range []struct {
		name  string
		view  viewKind
		setup func()
	}{
		{"Person/237", viewPerson, func() { m.personID = "237" }},
		{"Grades/237", viewGrades, func() { m.gradeStudentID = "237" }},
		{"Grades/86", viewGrades, func() { m.gradeStudentID = "86" }}, // has non-null float points
		{"Attendance/237", viewAttendance, func() { m.attStudentID = "237" }},
		{"Goals/237", viewGoals, func() { m.goalsStudentID = "237" }},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			m.view = tc.view
			rows, err := m.fetchRows()
			if err != nil {
				t.Fatalf("fetchRows: %v", err)
			}
			t.Logf("%d rows", len(rows))
		})
	}
}
