package main

import (
	edookit "github.com/martinpovolny/go-snip/edookit"
)

// MockSource implements edookit.DataSource with realistic Czech school demo data.
type MockSource struct{}

func ptr[T any](v T) *T { return &v }

func (m *MockSource) SearchPerson(criterion string) ([]edookit.Person, error) {
	return []edookit.Person{
		{
			Firstname: ptr("Jana"), Lastname: ptr("Slépičková"),
			EdookitID: 165, Email: ptr("jana.slepickova@seznam.cz"),
			Roles: []string{"parent"},
			Children: []edookit.PersonRef{
				{EdookitID: 237, Firstname: ptr("Jakub"), Lastname: ptr("Slépička"), Class: ptr("5.A")},
			},
			Representatives: []edookit.PersonRef{},
		},
	}, nil
}

func (m *MockSource) ListLessons(date string, _ edookit.LessonListOpts) (map[string]edookit.LessonEntry, error) {
	lessons := map[string]edookit.LessonEntry{
		"45100": makeLesson(45100, date+" 08:00:00", date+" 08:45:00", "M - 5.A", "Novák", "NOV", "Učebna 12"),
		"45101": makeLesson(45101, date+" 08:55:00", date+" 09:40:00", "ČJ - 5.A", "Procházka", "PRO", "Učebna 7"),
		"45102": makeLesson(45102, date+" 09:50:00", date+" 10:35:00", "Aj - 5.A", "Kovářová", "KOV", "Jazyková učebna"),
		"45103": makeLesson(45103, date+" 10:55:00", date+" 11:40:00", "Př - 5.A", "Horáček", "HOR", "Učebna 12"),
		"45104": makeLesson(45104, date+" 11:50:00", date+" 12:35:00", "VV - 5.A", "Šimánek", "ŠIM", "Výtvarná dílna"),
		"45105": makeLesson(45105, date+" 12:40:00", date+" 13:25:00", "Tv - 5.A", "Beneš", "BEN", "Tělocvična"),
	}
	return lessons, nil
}

func makeLesson(id int, from, to, name, teacherName, abbr, room string) edookit.LessonEntry {
	state := edookit.LessonState{
		LessonID:     id,
		DatetimeFrom: from,
		DatetimeTo:   to,
		Name:         name,
		Courses:      map[string]edookit.CourseRef{"1": {CourseID: id, CourseCode: name, CourseTypeID: 1}},
		Teachers:     map[string]edookit.TeacherRef{"1": {PersonID: id + 100, PersonAbbr: abbr}},
		Students:     map[string]edookit.StudentGroupRef{"1": {SubjectID: 1, SubjectName: "Žáci 5.A"}},
		Rooms:        map[string]edookit.RoomRef{"1": {RoomID: id + 200, RoomName: room}},
	}
	_ = teacherName
	return edookit.LessonEntry{Actual: state, Scheduled: state}
}

func (m *MockSource) ListRooms() (map[string]edookit.Room, error) {
	return map[string]edookit.Room{
		"1": {RoomID: 1, Name: "Učebna 12"},
		"2": {RoomID: 2, Name: "Jazyková učebna"},
		"3": {RoomID: 3, Name: "Tělocvična"},
	}, nil
}

func (m *MockSource) ListCourseTypes() (map[string]edookit.CourseType, error) {
	return map[string]edookit.CourseType{
		"1": {CourseTypeID: 1, Name: "Povinná výuka"},
		"2": {CourseTypeID: 2, Name: "Volitelná výuka"},
	}, nil
}

func (m *MockSource) ListWorkTypes() (map[string]edookit.WorkType, error) {
	return map[string]edookit.WorkType{
		"1": {WorktypeID: 1, Name: "Výuka"},
	}, nil
}

func (m *MockSource) ListStudents(_ edookit.StudentDataOpts) ([]edookit.Student, error) {
	grade5 := ptr(5)
	return []edookit.Student{
		{PersonID: 237, Firstname: "Jakub", Lastname: "Sláma", ClassName: ptr("5.A"), CurrentGradeNum: grade5, PrimaryEmail: ptr("jakub.slama@gmail.com"), DateOfBirth: ptr("2014-03-12"), Gender: ptr("M")},
		{PersonID: 238, Firstname: "Jana", Lastname: "Nováková", ClassName: ptr("5.A"), CurrentGradeNum: grade5, PrimaryEmail: ptr("jana.novakova@gmail.com"), DateOfBirth: ptr("2014-06-21"), Gender: ptr("F")},
		{PersonID: 239, Firstname: "Tomáš", Lastname: "Dvořák", ClassName: ptr("5.B"), CurrentGradeNum: grade5, PrimaryEmail: ptr("tomas.dvorak@gmail.com"), DateOfBirth: ptr("2014-01-08"), Gender: ptr("M")},
		{PersonID: 240, Firstname: "Petra", Lastname: "Horáková", ClassName: ptr("5.B"), CurrentGradeNum: grade5, PrimaryEmail: ptr("petra.horakova@seznam.cz"), DateOfBirth: ptr("2014-09-30"), Gender: ptr("F")},
		{PersonID: 241, Firstname: "Martin", Lastname: "Kopecký", ClassName: ptr("4.A"), CurrentGradeNum: ptr(4), PrimaryEmail: ptr("martin.kopecky@gmail.com"), DateOfBirth: ptr("2015-05-17"), Gender: ptr("M")},
		{PersonID: 242, Firstname: "Lucie", Lastname: "Marková", ClassName: ptr("4.A"), CurrentGradeNum: ptr(4), PrimaryEmail: ptr("lucie.markova@email.cz"), DateOfBirth: ptr("2015-11-02"), Gender: ptr("F")},
		{PersonID: 243, Firstname: "Ondřej", Lastname: "Blažek", ClassName: ptr("4.B"), CurrentGradeNum: ptr(4), PrimaryEmail: ptr("ondrej.blazek@gmail.com"), DateOfBirth: ptr("2015-07-14"), Gender: ptr("M")},
		{PersonID: 244, Firstname: "Karolína", Lastname: "Procházková", ClassName: ptr("4.B"), CurrentGradeNum: ptr(4), PrimaryEmail: ptr("karolina.prochazkova@seznam.cz"), DateOfBirth: ptr("2015-03-25"), Gender: ptr("F")},
	}, nil
}

func (m *MockSource) ListEmployees(_ edookit.StudentDataOpts) ([]edookit.Employee, error) {
	return []edookit.Employee{
		{PersonID: 301, Firstname: "Pavel", Lastname: "Novák", NameAbbr: ptr("NOV"), PrimaryEmail: ptr("p.novak@example-school.cz"), ClassNames: []string{"5.A"}},
		{PersonID: 302, Firstname: "Hana", Lastname: "Procházková", NameAbbr: ptr("PRO"), PrimaryEmail: ptr("h.prochazkova@example-school.cz"), ClassNames: []string{}},
		{PersonID: 303, Firstname: "Alena", Lastname: "Kovářová", NameAbbr: ptr("KOV"), PrimaryEmail: ptr("a.kovarova@example-school.cz"), ClassNames: []string{"4.A"}},
	}, nil
}

func (m *MockSource) ListEvaluations(_ edookit.EvalListOpts) ([]edookit.Evaluation, error) {
	return []edookit.Evaluation{
		{EvaluationID: "1", StudentID: "237", CourseID: "101", CourseCode: "M - 5.A", AchievementLevel: ptr("1"), EvaluatorFullName: ptr("Pavel Novák"), EvaluationDate: ptr("2026-01-15"), EvaltermName: ptr("1. pololetí 25/26"), Weight: ptr("1.00")},
		{EvaluationID: "2", StudentID: "237", CourseID: "102", CourseCode: "ČJ - 5.A", AchievementLevel: ptr("2"), EvaluatorFullName: ptr("Hana Procházková"), EvaluationDate: ptr("2026-01-15"), EvaltermName: ptr("1. pololetí 25/26"), Weight: ptr("1.00")},
		{EvaluationID: "3", StudentID: "238", CourseID: "101", CourseCode: "M - 5.A", AchievementLevel: ptr("1"), EvaluatorFullName: ptr("Pavel Novák"), EvaluationDate: ptr("2026-01-15"), EvaltermName: ptr("1. pololetí 25/26"), Weight: ptr("1.00")},
		{EvaluationID: "4", StudentID: "238", CourseID: "103", CourseCode: "Aj - 5.A", AchievementLevel: ptr("3"), EvaluatorFullName: ptr("Alena Kovářová"), EvaluationDate: ptr("2026-01-16"), EvaltermName: ptr("1. pololetí 25/26"), Weight: ptr("1.00")},
		{EvaluationID: "5", StudentID: "239", CourseID: "101", CourseCode: "M - 5.B", AchievementLevel: ptr("2"), EvaluatorFullName: ptr("Pavel Novák"), EvaluationDate: ptr("2026-01-15"), EvaltermName: ptr("1. pololetí 25/26"), Weight: ptr("1.00")},
		{EvaluationID: "6", StudentID: "240", CourseID: "102", CourseCode: "ČJ - 5.B", AchievementLevel: ptr("1"), EvaluatorFullName: ptr("Hana Procházková"), EvaluationDate: ptr("2026-01-15"), EvaltermName: ptr("1. pololetí 25/26"), Weight: ptr("1.00")},
	}, nil
}

func (m *MockSource) ListCourses(_ edookit.CourseListOpts) ([]edookit.StudentWithCourses, error) {
	return []edookit.StudentWithCourses{
		{
			ID: 237, FirstName: "Jakub", LastName: "Sláma",
			Courses: []edookit.StudentCourse{
				{ID: 101, Code: "M - 5.A", AttendancePercentage: ptr("95.00"), FinalEvaluations: []edookit.CourseFinalEvaluation{{Grade: ptr("1")}}},
				{ID: 102, Code: "ČJ - 5.A", AttendancePercentage: ptr("98.00"), FinalEvaluations: []edookit.CourseFinalEvaluation{{Grade: ptr("2")}}},
				{ID: 103, Code: "Aj - 5.A", AttendancePercentage: ptr("92.00"), FinalEvaluations: []edookit.CourseFinalEvaluation{{Grade: ptr("1")}}},
			},
		},
		{
			ID: 238, FirstName: "Jana", LastName: "Nováková",
			Courses: []edookit.StudentCourse{
				{ID: 101, Code: "M - 5.A", AttendancePercentage: ptr("88.00"), FinalEvaluations: []edookit.CourseFinalEvaluation{{Grade: ptr("1")}}},
				{ID: 103, Code: "Aj - 5.A", AttendancePercentage: ptr("94.00"), FinalEvaluations: []edookit.CourseFinalEvaluation{{Grade: ptr("3")}}},
			},
		},
	}, nil
}

func (m *MockSource) ListPaymentPrescriptions(_ edookit.PrescriptionListOpts) ([]edookit.PaymentPrescription, error) {
	return []edookit.PaymentPrescription{
		{ID: 1, Name: "Výlet do Vídně", DueDate: ptr("2026-03-01"), Amount: ptr("800"), CurrencyID: ptr(1), State: 1},
		{ID: 2, Name: "Školní potřeby 2. pololetí", DueDate: ptr("2026-02-15"), Amount: ptr("350"), CurrencyID: ptr(1), State: 1},
		{ID: 3, Name: "Lyžařský výcvik", DueDate: ptr("2026-01-10"), Amount: ptr("4500"), CurrencyID: ptr(1), State: 2},
		{ID: 4, Name: "Divadelní představení", DueDate: ptr("2026-04-20"), Amount: ptr("120"), CurrencyID: ptr(1), State: 1},
	}, nil
}

func (m *MockSource) ListStudentStats(_ string) ([]edookit.StudentStat, error) {
	return []edookit.StudentStat{
		{PersonAnonIdent: "9dc8eba3-f973-4a1a-af44-9cdb81722f76", Age: ptr(11), Gender: ptr("M"), ClassName: ptr("5.A"), CurrentGradeNum: ptr(5), CitizenshipCountryCode: ptr("CZE")},
		{PersonAnonIdent: "1bc2dfe4-a123-4b56-cd78-9e0f12345678", Age: ptr(11), Gender: ptr("F"), ClassName: ptr("5.A"), CurrentGradeNum: ptr(5), CitizenshipCountryCode: ptr("CZE")},
	}, nil
}

func (m *MockSource) ListAttendanceLessons(_ int, date string) ([]edookit.AttendanceLesson, error) {
	return []edookit.AttendanceLesson{
		{PersonID: "237", LessonID: "45100", Timerange: `["` + date + ` 08:00:00","` + date + ` 08:45:00")`, CourseNames: []string{"M - 1."}, RoomNames: []string{"Učebna 12"}, AttendanceStatus: ptr("P")},
		{PersonID: "237", LessonID: "45101", Timerange: `["` + date + ` 08:55:00","` + date + ` 09:40:00")`, CourseNames: []string{"ČJ - 1."}, RoomNames: []string{"Učebna 7"}, AttendanceStatus: ptr("P")},
		{PersonID: "237", LessonID: "45102", Timerange: `["` + date + ` 10:55:00","` + date + ` 11:40:00")`, CourseNames: []string{"Aj - 1."}, RoomNames: []string{"Jazyková učebna"}, AttendanceStatus: ptr("A")},
	}, nil
}

func (m *MockSource) ListIndividualLessonContent(_ int, _ int, _, _ string) ([]edookit.IndividualLessonContent, error) {
	return []edookit.IndividualLessonContent{
		{
			StudentID: 237, StudentIdentifier: "237",
			Lessons: []edookit.IndividualLessonEntry{
				{LessonID: "45100", ActualTimerange: ptr(`["2026-05-21 08:00:00","2026-05-21 08:45:00")`), CourseCode: ptr("M - 1."), SubjectName: ptr("Matematika"), StudentLessonTopic: ptr("Sčítání do 20")},
				{LessonID: "45101", ActualTimerange: ptr(`["2026-05-21 08:55:00","2026-05-21 09:40:00")`), CourseCode: ptr("ČJ - 1."), SubjectName: ptr("Český jazyk"), StudentLessonTopic: ptr("Písmeno B")},
			},
		},
	}, nil
}

func (m *MockSource) ListPublicEvents(_ edookit.PublicEventsOpts) ([]edookit.PublicEvent, error) {
	return []edookit.PublicEvent{
		{ID: 1, Title: "Třídnická hodina", Place: "Učebna 5", Participant: "5.A", From: "2026-05-21T08:00:00+02:00", To: "2026-05-21T08:45:00+02:00"},
		{ID: 2, Title: "Školní výlet", Place: "", Participant: "4.A, 4.B", From: "2026-05-23T07:00:00+02:00", To: "2026-05-23T16:00:00+02:00"},
	}, nil
}

func (m *MockSource) ListScheduleChanges(_ edookit.ScheduleChangeOpts) ([]edookit.ScheduleChange, error) {
	return []edookit.ScheduleChange{
		{
			Scheduled: edookit.ScheduleSlot{
				Courses:  []string{"Čj - 1."},
				Rooms:    []string{"Učebna 2"},
				Teachers: []string{"MB"},
				Students: []string{"1."},
				Timerange: &edookit.ScheduleRange{From: "2026-05-21 08:00:00", To: "2026-05-21 08:45:00"},
			},
			Actual: edookit.ScheduleSlot{
				Courses:  []string{},
				Rooms:    []string{},
				Teachers: []string{"MB"},
				Students: []string{"1."},
				Timerange: &edookit.ScheduleRange{From: "2026-05-21 08:00:00", To: "2026-05-21 08:45:00"},
				Event:    "Projektový den",
			},
		},
	}, nil
}
