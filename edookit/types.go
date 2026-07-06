package edookit

// --- Person Search ---

type PersonRef struct {
	LegalRole  *string `json:"legalrole"`
	EdookitID  int     `json:"edookit_id"`
	Firstname  *string `json:"firstname"`
	Lastname   *string `json:"lastname"`
	Email      *string `json:"email"`
	Plus4UID   *string `json:"plus4u_id"`
	Class      *string `json:"class"`
}

type Person struct {
	Firstname       *string     `json:"firstname"`
	Lastname        *string     `json:"lastname"`
	EdookitID       int         `json:"edookit_id"`
	Email           *string     `json:"email"`
	Roles           []string    `json:"roles"`
	Children        []PersonRef `json:"children"`
	Representatives []PersonRef `json:"representatives"`
	Plus4UID        *string     `json:"plus4u_id"`
	Class           *string     `json:"class"`
}

// --- Lessons ---

type CourseRef struct {
	CourseID     int    `json:"course_id"`
	CourseCode   string `json:"course_code"`
	CourseTypeID int    `json:"course_type_id"`
}

type TeacherRef struct {
	PersonID   int    `json:"person_id"`
	PersonAbbr string `json:"person_abbr"`
}

type StudentGroupRef struct {
	SubjectID   int    `json:"subject_id"`
	SubjectName string `json:"subject_name"`
}

type RoomRef struct {
	RoomID   int    `json:"room_id"`
	RoomName string `json:"room_name"`
}

type LessonState struct {
	LessonID     int                        `json:"lesson_id"`
	DatetimeFrom string                     `json:"datetime_from"`
	DatetimeTo   string                     `json:"datetime_to"`
	LessonJoinID *int                       `json:"lessonjoin_id"`
	Name         string                     `json:"name"`
	Courses      map[string]CourseRef       `json:"courses"`
	Teachers     map[string]TeacherRef      `json:"teachers"`
	Students     map[string]StudentGroupRef `json:"students"`
	Rooms        map[string]RoomRef         `json:"rooms"`
}

type LessonEntry struct {
	Actual    LessonState `json:"actual"`
	Scheduled LessonState `json:"scheduled"`
}

type LessonsResponse struct {
	Lessons map[string]LessonEntry `json:"lessons"`
}

type LessonListOpts struct {
	CourseID        *int
	CourseTypeID    *int
	RoomID          *int
	StudentPersonID *int
	TeacherPersonID *int
	WorkTypeID      *int
}

type Room struct {
	RoomID int    `json:"room_id"`
	Name   string `json:"name"`
}

type RoomsResponse struct {
	Rooms map[string]Room `json:"rooms"`
}

type CourseType struct {
	CourseTypeID int    `json:"course_type_id"`
	Name         string `json:"name"`
}

type CourseTypesResponse struct {
	CourseTypes map[string]CourseType `json:"course_types"`
}

type WorkType struct {
	WorktypeID int    `json:"worktype_id"`
	Name       string `json:"name"`
}

type WorkTypesResponse struct {
	WorkTypes map[string]WorkType `json:"work_types"`
}

// --- Students & Employees ---

type Address struct {
	CountryCode    *string `json:"CountryCode"`
	PostalCode     *string `json:"PostalCode"`
	City           *string `json:"City"`
	Street         *string `json:"Street"`
	HouseNo        *string `json:"HouseNo"`
	LandRegistryNo *string `json:"LandRegistryNo"`
}

type Student struct {
	PersonID               int      `json:"PersonId"`
	Firstname              string   `json:"Firstname"`
	Middlename             *string  `json:"Middlename"`
	Lastname               string   `json:"Lastname"`
	DegreePreceding        *string  `json:"DegreePreceding"`
	DegreeFollowing        *string  `json:"DegreeFollowing"`
	DateOfBirth            *string  `json:"DateOfBirth"`
	PlaceOfBirth           *string  `json:"PlaceOfBirth"`
	Age                    *int     `json:"Age"`
	Gender                 *string  `json:"Gender"`
	PersonalNumber         *string  `json:"PersonalNumber"`
	PermanentStayAddress   *Address `json:"PermanentStayAddress"`
	CitizenshipCountryCode *string  `json:"CitizenshipCountryCode"`
	HealthInsuranceCode    *string  `json:"HealthInsuranceCompanyCode"`
	OrganizationName       string   `json:"OrganizationName"`
	OrganizationIdent      string   `json:"OrganizationIdent"`
	EnrolledSince          string   `json:"EnrolledSince"`
	UnenrolledSince        *string  `json:"UnenrolledSince"`
	ClassName              *string  `json:"ClassName"`
	LearningGroupNames     []string `json:"LearningGroupNames"`
	InitialGradeNum        *int     `json:"InitialGradeNum"`
	CurrentGradeNum        *int     `json:"CurrentGradeNum"`
	Phone                  *string  `json:"Phone"`
	PhoneMobile            *string  `json:"PhoneMobile"`
	PrimaryEmail           *string  `json:"PrimaryEmail"`
	Plus4UId               *string  `json:"Plus4UId"`
	VariableSymbol         *string  `json:"VariableSymbol"`
}

type StudentsResponse struct {
	Students []Student `json:"Students"`
}

type Employee struct {
	PersonID         int      `json:"PersonId"`
	Firstname        string   `json:"Firstname"`
	Middlename       *string  `json:"Middlename"`
	Lastname         string   `json:"Lastname"`
	DegreePreceding  *string  `json:"DegreePreceding"`
	DegreeFollowing  *string  `json:"DegreeFollowing"`
	NameAbbr         *string  `json:"NameAbbr"`
	CustomIdentifier *string  `json:"CustomIdentifier"`
	DateOfBirth      *string  `json:"DateOfBirth"`
	Age              *int     `json:"Age"`
	Gender           *string  `json:"Gender"`
	FamilyStatus     *string  `json:"FamilyStatus"`
	OrganizationName string   `json:"OrganizationName"`
	OrganizationIdent string  `json:"OrganizationIdent"`
	EnrolledSince    string   `json:"EnrolledSince"`
	UnenrolledSince  *string  `json:"UnenrolledSince"`
	ClassNames       []string `json:"ClassNames"`
	Phone            *string  `json:"Phone"`
	PhoneMobile      *string  `json:"PhoneMobile"`
	PrimaryEmail     *string  `json:"PrimaryEmail"`
	Plus4UId         *string  `json:"Plus4UId"`
	VariableSymbol   *string  `json:"VariableSymbol"`
}

type EmployeesResponse struct {
	Employees []Employee `json:"Employees"`
}

type StudentDataOpts struct {
	Date                string // YYYY-MM-DD, empty = today
	IncludeInactiveSince string // YYYY-MM-DD, empty = active only
}

// --- Grades / Evaluation ---

// Evaluation IDs are returned as JSON strings by the API.
type Evaluation struct {
	EvaluationID           string  `json:"evaluation_id"`
	StudentID              string  `json:"student_id"`
	CourseID               string  `json:"course_id"`
	CourseCode             string  `json:"course_code"`
	AchievementLevel       *string  `json:"achievement_level"`
	Percentage             *float64 `json:"percentage,string"`
	Points                 *float64 `json:"points,string"`
	BonusPoints            *float64 `json:"bonus_points,string"`
	VerbalEvaluation       *string `json:"verbal_evaluation"`
	AssignmentID           *string `json:"assignment_id"`
	AssignmentName         *string `json:"assignment_name"`
	AssignmentCategoryID   *string `json:"assignment_category_id"`
	AssignmentCategoryName *string `json:"assignment_category_name"`
	EvaluatorID            *string `json:"evaluator_id"`
	EvaluatorFullName      *string `json:"evaluator_full_name"`
	EvaluationDate         *string `json:"evaluation_date"`
	EvaluationTime         *string `json:"evaluation_time"`
	EvaltermID             *string `json:"evalterm_id"`
	EvaltermName           *string `json:"evalterm_name"`
	NotEvaluated           bool    `json:"not_evaluated"`
	NoEvalReasonID         *string `json:"no_eval_reason_id"`
	NoEvalReasonName       *string `json:"no_eval_reason_name"`
	ResitEvaluationID      *string `json:"resit_evaluation_id"`
	PublishedComment       *string `json:"published_comment"`
	Weight                 *string `json:"weight"`
}

type EvalListOpts struct {
	StudentID  *int
	CourseID   *int
	EvaltermID *int
	DateFrom   *string
	DateTo     *string
}

// --- Courses ---

type CourseFinalEvaluation struct {
	ID                         int     `json:"id,string"`
	EvaluationTermID           int     `json:"evaluation_term_id,string"`
	EvaluationTermName         string  `json:"evaluation_term_name"`
	Grade                      *string  `json:"grade"`
	Points                     *float64 `json:"points,string"`
	Extrapoints                *float64 `json:"extrapoints,string"`
	Percentages                *float64 `json:"percentages,string"`
	VerbalEvaluation           *string  `json:"verbal_evaluation"`
	EvaluationDate             *string  `json:"evaluation_date"`
	Weight                     *string  `json:"weight"`
	PublishedComment           *string  `json:"published_comment"`
	ResitEvaluationID          *int     `json:"resit_evaluation_id,string"`
	NotEvaluatedReasonCode     *string `json:"not_evaluated_reason_code"`
	IncludedInOverallEvaluation bool   `json:"included_in_overall_evaluation"`
}

type CourseLesson struct {
	ID               int    `json:"id,string"`
	Timerange        string `json:"timerange"`
	AttendanceStatus string `json:"attendance_status"`
}

type StudentCourse struct {
	ID                   int                     `json:"id,string"`
	Code                 string                  `json:"code"`
	Lessons              []CourseLesson           `json:"lessons"`
	FinalEvaluations     []CourseFinalEvaluation  `json:"final_evaluations"`
	AttendancePercentage *string                  `json:"attendance_percentage"`
}

type StudentWithCourses struct {
	ID         int             `json:"id,string"`
	FirstName  string          `json:"first_name"`
	LastName   string          `json:"last_name"`
	MiddleName *string         `json:"middle_name"`
	Courses    []StudentCourse `json:"courses"`
}

type CoursesResponse struct {
	Students []StudentWithCourses `json:"students"`
}

type CourseListOpts struct {
	PersonID   *int
	EvaltermID *int
}

// --- Payments ---

type Currency struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type BankAccount struct {
	ID                int    `json:"id"`
	BankAccountNumber string `json:"bankAccountNumber"`
}

type PaymentType struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

type Organization struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PaymentPrescription struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Description    *string `json:"description"`
	DueDate        *string `json:"due_date"`
	Amount         *string `json:"amount"`
	CurrencyID     *int    `json:"currency_id"`
	State          int     `json:"state"`
	SpecificSymbol *string `json:"specific_symbol"`
	IsCredit       bool    `json:"is_credit"`
	Direction      int     `json:"direction"`
}

type PrescriptionListOpts struct {
	DateFrom             *string
	DateTo               *string
	SchoolYear           *string
	PrescriptionID       *int
	PersonID             *int
	PaymentID            *int
	OperationIdentifier  *string
	SpecificSymbol       *string
}

// --- Statistics ---

type StudentStat struct {
	PersonAnonIdent           string  `json:"PersonAnonIdent"`
	Age                       *int    `json:"Age"`
	Gender                    *string `json:"Gender"`
	ClassName                 *string `json:"ClassName"`
	EnrolledSince             *string `json:"EnrolledSince"`
	InitialGradeNum           *int    `json:"InitialGradeNum"`
	CurrentGradeNum           *int    `json:"CurrentGradeNum"`
	OrganizationName          *string `json:"OrganizationName"`
	OrganizationIdent         *string `json:"OrganizationIdent"`
	UnenrolledSince           *string `json:"UnenrolledSince"`
	UIV_ZPUSOB                *string `json:"UIV_ZPUSOB"`
	CitizenshipCountryCode    *string `json:"CitizenshipCountryCode"`
	CitizenshipQualifierCode  *string `json:"CitizenshipQualifierCode"`
	PermanentStayCountryCode  *string `json:"PermanentStayCountryCode"`
}

type StudentStatsResponse struct {
	Students []StudentStat `json:"Students"`
}

// --- Změnový rozvrh (Schedule Changes) ---
// Served at <school>-login.edookit.net, HTTP Basic auth.

type ScheduleSlot struct {
	Courses   []string       `json:"courses"`
	Rooms     []string       `json:"rooms"`
	Teachers  []string       `json:"teachers"`
	Students  []string       `json:"students"`
	Timerange *ScheduleRange `json:"timerange"`
	Event     string         `json:"event,omitempty"`
}

type ScheduleRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ScheduleChange struct {
	Scheduled ScheduleSlot `json:"scheduled"`
	Actual    ScheduleSlot `json:"actual"`
}

type ScheduleChangesResponse struct {
	Version string           `json:"version"`
	Change  []ScheduleChange `json:"change"`
}

type ScheduleChangeOpts struct {
	From             string // YYYY-MM-DD, defaults to today
	To               string // YYYY-MM-DD, defaults to today
	TeacherNameFormat string // "code" (default), "name", or "full"
}

// --- Veřejné události (Public Events) ---
// No auth required. Served at <school>-login.edookit.net.

type PublicEvent struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Place       string `json:"place"`
	Participant string `json:"participant"`
	From        string `json:"from"`
	To          string `json:"to"`
}

type PublicEventsResponse struct {
	Version string        `json:"version"`
	Events  []PublicEvent `json:"events"`
}

type PublicEventsOpts struct {
	From string // YYYY-MM-DD
	To   string // YYYY-MM-DD
}

// --- Přímý zápis docházky (Direct Attendance) ---
// HTTP Basic. Read-only (GET) portion only.

type AttendanceLesson struct {
	PersonID         string  `json:"person_id"`
	LessonID         string  `json:"lesson_id"`
	Timerange        string  `json:"timerange"`
	RoomIDs          []int   `json:"room_ids"`
	CourseIDs        []int   `json:"course_ids"`
	RoomNames        []string `json:"room_names"`
	CourseNames      []string `json:"course_names"`
	AttendanceStatus *string `json:"attendance_status"`
	MinutesLate      *int    `json:"minutes_late"`
	MinutesEarly     *int    `json:"minutes_early"`
}

// --- Individuální cíle (Individual Goals) ---
// HTTP Basic.

type IndividualLessonContent struct {
	StudentID         int                      `json:"student_id"`
	StudentIdentifier string                   `json:"student_identifier"`
	Lessons           []IndividualLessonEntry  `json:"lessons"`
}

type IndividualLessonEntry struct {
	LessonID                  string  `json:"lesson_id"`
	ActualTimerange           *string `json:"actual_timerange"`
	StudentLessonTopic        *string `json:"student_lesson_topic"`
	StudentPlannedLessonTopic *string `json:"student_planned_lesson_topic"`
	StudentLessonNote         *string `json:"student_lesson_note"`
	SubjectName               *string `json:"subject_name"`
	CourseCode                *string `json:"course_code"`
}

type LearningAgreement struct {
	StudentID          int                   `json:"student_id"`
	StudentIdentifier  string                `json:"student_identifier"`
	LearningAgreements []LearningAgreementEntry `json:"learning_agreements"`
}

type LearningAgreementEntry struct {
	LearningAgreementID     string   `json:"learning_agreement_id"`
	CreatedByPersonID       string   `json:"created_by_person_id"`
	CreatedByPersonName     string   `json:"created_by_person_name"`
	ValidityRange           string   `json:"validity_range"`
	StudentFeedbackPositive string   `json:"student_feedback_positive"`
	StudentFeedbackNegative string   `json:"student_feedback_negative"`
	StudentFeedbackOther    string   `json:"student_feedback_other"`
	CoachFeedbackPositive   string   `json:"coach_feedback_positive"`
	CoachFeedbackNegative   string   `json:"coach_feedback_negative"`
	CoachFeedbackOther      string   `json:"coach_feedback_other"`
	NextGoals               string   `json:"next_goals"`
	NextAgreementDate       string   `json:"next_agreement_date"`
	Coaches                 []struct {
		CoachID   int    `json:"coach_id"`
		CoachName string `json:"coach_name"`
	} `json:"coaches"`
	TargetSubcompetences []struct {
		AbbreviatedIdentifier string `json:"abbreviated_identifier"`
	} `json:"target_subcompetences"`
}
