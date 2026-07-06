package edookit

// DataSource abstracts the Edookit API so the TUI can work with
// either a real HTTP client or a mock in tests / demo mode.
type DataSource interface {
	SearchPerson(criterion string) ([]Person, error)

	ListLessons(date string, opts LessonListOpts) (map[string]LessonEntry, error)
	ListRooms() (map[string]Room, error)
	ListCourseTypes() (map[string]CourseType, error)
	ListWorkTypes() (map[string]WorkType, error)

	ListStudents(opts StudentDataOpts) ([]Student, error)
	ListEmployees(opts StudentDataOpts) ([]Employee, error)

	ListEvaluations(opts EvalListOpts) ([]Evaluation, error)

	ListCourses(opts CourseListOpts) ([]StudentWithCourses, error)

	ListPaymentPrescriptions(opts PrescriptionListOpts) ([]PaymentPrescription, error)

	ListStudentStats(date string) ([]StudentStat, error)

	ListAttendanceLessons(studentID int, date string) ([]AttendanceLesson, error)

	ListIndividualLessonContent(studentID, lessonID int, dateFrom, dateTo string) ([]IndividualLessonContent, error)

	ListScheduleChanges(opts ScheduleChangeOpts) ([]ScheduleChange, error)

	ListPublicEvents(opts PublicEventsOpts) ([]PublicEvent, error)
}
