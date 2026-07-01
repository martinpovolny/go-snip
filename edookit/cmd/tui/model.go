package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	edookit "github.com/martinpovolny/go-snip/edookit"
)

type viewKind int

const (
	viewSchedule viewKind = iota // default — confirmed working
	viewEvents
	viewPerson
	viewStudents
	viewEmployees
	viewLessons
	viewGrades
	viewCourses
	viewPayments
	viewStats
	viewAttendance
	viewGoals
)

func (v viewKind) String() string {
	switch v {
	case viewSchedule:
		return "Schedule Changes"
	case viewEvents:
		return "Public Events"
	case viewPerson:
		return "Person"
	case viewStudents:
		return "Students"
	case viewEmployees:
		return "Employees"
	case viewLessons:
		return "Lessons"
	case viewGrades:
		return "Grades"
	case viewCourses:
		return "Courses"
	case viewPayments:
		return "Payments"
	case viewStats:
		return "Student Stats"
	case viewAttendance:
		return "Attendance"
	case viewGoals:
		return "Individual Goals"
	}
	return "Unknown"
}

type rowsLoadedMsg struct {
	rows []table.Row
	err  error
}

type Model struct {
	source   edookit.DataSource
	baseURL  string
	mockMode bool

	view           viewKind
	prevView       viewKind
	hasPrev        bool
	personID       string // active person ID for viewPerson
	gradeStudentID string // active student ID for viewGrades
	attStudentID   string // active student ID for viewAttendance
	attDate        string // date for viewAttendance (may differ from today after < / >)
	goalsStudentID string // active student ID for viewGoals
	table          table.Model
	allRows        []table.Row

	cmdMode  bool
	cmdInput textinput.Model

	filterMode  bool
	filterInput textinput.Model
	filterText  string

	status string
	err    error

	width, height int
	today         string
	weekFrom      string
	weekTo        string
	monthFrom     string
	monthTo       string
}

func newModel(source edookit.DataSource, baseURL string, mockMode bool) Model {
	cmd := textinput.New()
	cmd.Placeholder = "sc · ev · person <id> · students · emp · grades <id> · courses · payments · stats · att <id> · goals <id>"
	cmd.CharLimit = 60

	flt := textinput.New()
	flt.Placeholder = "filter..."
	flt.CharLimit = 50

	now := time.Now()
	mon := now
	for mon.Weekday() != time.Monday {
		mon = mon.AddDate(0, 0, -1)
	}
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, -1)

	m := Model{
		source:      source,
		baseURL:     baseURL,
		mockMode:    mockMode,
		view:        viewSchedule,
		cmdInput:    cmd,
		filterInput: flt,
		today:       now.Format("2006-01-02"),
		attDate:     now.Format("2006-01-02"),
		weekFrom:    mon.Format("2006-01-02"),
		weekTo:      mon.AddDate(0, 0, 6).Format("2006-01-02"),
		monthFrom:   monthStart.Format("2006-01-02"),
		monthTo:     monthEnd.Format("2006-01-02"),
		width:       80,
		height:      24,
	}
	m.table = m.buildTable(nil)
	return m
}

func (m Model) Init() tea.Cmd {
	return m.loadCmd()
}

func (m Model) loadCmd() tea.Cmd {
	return func() tea.Msg {
		rows, err := m.fetchRows()
		return rowsLoadedMsg{rows: rows, err: err}
	}
}

func (m Model) fetchRows() ([]table.Row, error) {
	switch m.view {
	case viewSchedule:
		changes, err := m.source.ListScheduleChanges(edookit.ScheduleChangeOpts{
			From: m.weekFrom,
			To:   m.weekTo,
		})
		if err != nil {
			return nil, err
		}
		rows := make([]table.Row, len(changes))
		for i, ch := range changes {
			date, timeFrom := splitDateTime(ch.Scheduled.Timerange)
			original := joinStrings(ch.Scheduled.Courses) + " " + joinStrings(ch.Scheduled.Teachers)
			replacement := scheduleActual(ch.Actual)
			students := joinStrings(ch.Scheduled.Students)
			rows[i] = table.Row{date, timeFrom, strings.TrimSpace(original), replacement, students}
		}
		return rows, nil

	case viewEvents:
		events, err := m.source.ListPublicEvents(edookit.PublicEventsOpts{
			From: m.monthFrom,
			To:   m.monthTo,
		})
		if err != nil {
			return nil, err
		}
		rows := make([]table.Row, len(events))
		for i, ev := range events {
			date, timeFrom := splitISODateTime(ev.From)
			rows[i] = table.Row{date, timeFrom, ev.Title, ev.Place, ev.Participant}
		}
		return rows, nil

	case viewPerson:
		if m.personID == "" {
			return nil, nil
		}
		persons, err := m.source.SearchPerson(m.personID)
		if err != nil {
			return nil, err
		}
		var rows []table.Row
		for _, p := range persons {
			name := strVal(p.Firstname) + " " + strVal(p.Lastname)
			email := strVal(p.Email)
			class := strVal(p.Class)
			roles := strings.Join(p.Roles, ", ")
			rows = append(rows, table.Row{fmt.Sprint(p.EdookitID), name, roles, class, email})
			for _, ch := range p.Children {
				chName := strVal(ch.Firstname) + " " + strVal(ch.Lastname)
				chEmail := strVal(ch.Email)
				chClass := strVal(ch.Class)
				rows = append(rows, table.Row{fmt.Sprint(ch.EdookitID), "  └ " + chName, "child", chClass, chEmail})
			}
			for _, rep := range p.Representatives {
				rName := strVal(rep.Firstname) + " " + strVal(rep.Lastname)
				rEmail := strVal(rep.Email)
				rows = append(rows, table.Row{fmt.Sprint(rep.EdookitID), "  └ " + rName, "rep", "", rEmail})
			}
		}
		return rows, nil

	case viewStudents:
		students, err := m.source.ListStudents(edookit.StudentDataOpts{})
		if err != nil {
			return nil, err
		}
		rows := make([]table.Row, len(students))
		for i, s := range students {
			rows[i] = table.Row{fmt.Sprint(s.PersonID), s.Firstname + " " + s.Lastname, strVal(s.ClassName), strVal(s.PrimaryEmail)}
		}
		return rows, nil

	case viewEmployees:
		employees, err := m.source.ListEmployees(edookit.StudentDataOpts{})
		if err != nil {
			return nil, err
		}
		rows := make([]table.Row, len(employees))
		for i, e := range employees {
			rows[i] = table.Row{fmt.Sprint(e.PersonID), e.Firstname + " " + e.Lastname, strVal(e.NameAbbr), strings.Join(e.ClassNames, ", "), strVal(e.PrimaryEmail)}
		}
		return rows, nil

	case viewLessons:
		lessons, err := m.source.ListLessons(m.today, edookit.LessonListOpts{})
		if err != nil {
			return nil, err
		}
		var rows []table.Row
		for _, entry := range lessons {
			a := entry.Actual
			from := a.DatetimeFrom[11:16]
			to := a.DatetimeTo[11:16]
			rows = append(rows, table.Row{from + "–" + to, a.Name, firstTeacherAbbr(a.Teachers), firstRoomName(a.Rooms)})
		}
		return rows, nil

	case viewGrades:
		if m.gradeStudentID == "" {
			return nil, fmt.Errorf("usage: grades <student-id>")
		}
		id, err := strconv.Atoi(m.gradeStudentID)
		if err != nil {
			return nil, fmt.Errorf("invalid student id: %s", m.gradeStudentID)
		}
		evals, err := m.source.ListEvaluations(edookit.EvalListOpts{StudentID: &id})
		if err != nil {
			return nil, err
		}
		rows := make([]table.Row, len(evals))
		for i, e := range evals {
			rows[i] = table.Row{e.CourseCode, strVal(e.AchievementLevel), strVal(e.AssignmentName), strVal(e.EvaluatorFullName), strVal(e.EvaluationDate)}
		}
		return rows, nil

	case viewCourses:
		students, err := m.source.ListCourses(edookit.CourseListOpts{})
		if err != nil {
			return nil, err
		}
		var rows []table.Row
		for _, s := range students {
			id := fmt.Sprint(s.ID)
			name := s.FirstName + " " + s.LastName
			for _, c := range s.Courses {
				rows = append(rows, table.Row{id, name, c.Code, strVal(c.AttendancePercentage) + "%", firstGrade(c.FinalEvaluations)})
			}
		}
		return rows, nil

	case viewPayments:
		prescriptions, err := m.source.ListPaymentPrescriptions(edookit.PrescriptionListOpts{})
		if err != nil {
			return nil, err
		}
		rows := make([]table.Row, len(prescriptions))
		for i, p := range prescriptions {
			state := "open"
			if p.State == 2 {
				state = "closed"
			}
			rows[i] = table.Row{fmt.Sprint(p.ID), p.Name, strVal(p.DueDate), strVal(p.SpecificSymbol), state}
		}
		return rows, nil

	case viewStats:
		stats, err := m.source.ListStudentStats("")
		if err != nil {
			return nil, err
		}
		rows := make([]table.Row, len(stats))
		for i, s := range stats {
			rows[i] = table.Row{
				strVal(s.ClassName),
				strVal(s.Gender),
				fmt.Sprint(intVal(s.CurrentGradeNum)),
				fmt.Sprint(intVal(s.Age)),
				strVal(s.CitizenshipCountryCode),
			}
		}
		return rows, nil

	case viewAttendance:
		if m.attStudentID == "" {
			return nil, fmt.Errorf("usage: att <student-id>")
		}
		id, err := strconv.Atoi(m.attStudentID)
		if err != nil {
			return nil, fmt.Errorf("invalid student id: %s", m.attStudentID)
		}
		lessons, err := m.source.ListAttendanceLessons(id, m.attDate)
		if err != nil {
			return nil, err
		}
		rows := make([]table.Row, len(lessons))
		for i, l := range lessons {
			timeFrom, timeTo := parseTimerange(l.Timerange)
			status := strVal(l.AttendanceStatus)
			if status == "" {
				status = "–"
			}
			rows[i] = table.Row{timeFrom + "–" + timeTo, strings.Join(l.CourseNames, ", "), strings.Join(l.RoomNames, ", "), status}
		}
		return rows, nil

	case viewGoals:
		if m.goalsStudentID == "" {
			return nil, fmt.Errorf("usage: goals <student-id>")
		}
		id, err := strconv.Atoi(m.goalsStudentID)
		if err != nil {
			return nil, fmt.Errorf("invalid student id: %s", m.goalsStudentID)
		}
		contents, err := m.source.ListIndividualLessonContent(id, 0, m.monthFrom, m.monthTo)
		if err != nil {
			return nil, err
		}
		var rows []table.Row
		for _, c := range contents {
			for _, l := range c.Lessons {
				timeFrom, _ := parseTimerange(strValDefault(l.ActualTimerange, ""))
				rows = append(rows, table.Row{
					timeFrom,
					strVal(l.CourseCode),
					strVal(l.SubjectName),
					strVal(l.StudentLessonTopic),
				})
			}
		}
		return rows, nil
	}
	return nil, nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table = m.buildTable(m.allRows)
		return m, nil

	case rowsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = "error: " + msg.err.Error()
		} else {
			m.err = nil
			m.allRows = msg.rows
			m.status = ""
			m.applyFilter()
		}
		return m, nil

	case tea.KeyMsg:
		if m.cmdMode {
			switch msg.Type {
			case tea.KeyEsc:
				m.cmdMode = false
				m.cmdInput.SetValue("")
			case tea.KeyEnter:
				input := strings.TrimSpace(m.cmdInput.Value())
				m.cmdMode = false
				m.cmdInput.SetValue("")
				return m, m.executeCommand(input)
			default:
				var tiCmd tea.Cmd
				m.cmdInput, tiCmd = m.cmdInput.Update(msg)
				return m, tiCmd
			}
			return m, nil
		}

		if m.filterMode {
			switch msg.Type {
			case tea.KeyEsc:
				m.filterMode = false
				m.filterText = ""
				m.filterInput.SetValue("")
				m.applyFilter()
			case tea.KeyEnter:
				m.filterText = m.filterInput.Value()
				m.filterMode = false
				m.applyFilter()
			default:
				var tiCmd tea.Cmd
				m.filterInput, tiCmd = m.filterInput.Update(msg)
				m.filterText = m.filterInput.Value()
				m.applyFilter()
				return m, tiCmd
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case ":":
			m.cmdMode = true
			m.cmdInput.Focus()
			return m, textinput.Blink
		case "/":
			m.filterMode = true
			m.filterInput.SetValue("")
			m.filterInput.Focus()
			return m, textinput.Blink
		case "r":
			m.status = "loading…"
			return m, m.loadCmd()
		case "esc":
			if m.filterText != "" {
				m.filterText = ""
				m.filterInput.SetValue("")
				m.applyFilter()
			} else if m.hasPrev {
				return m, m.goBack()
			}
		case "enter":
			return m, m.handleEnter()
		case "g":
			if m.isStudentContext() {
				return m, m.navigateTo(viewGrades)
			}
			var tCmd tea.Cmd
			m.table, tCmd = m.table.Update(msg)
			return m, tCmd
		case "a":
			if m.isStudentContext() {
				return m, m.navigateTo(viewAttendance)
			}
			var tCmd tea.Cmd
			m.table, tCmd = m.table.Update(msg)
			return m, tCmd
		case "i":
			if m.isStudentContext() {
				return m, m.navigateTo(viewGoals)
			}
			var tCmd tea.Cmd
			m.table, tCmd = m.table.Update(msg)
			return m, tCmd
		case "<":
			if m.isDateNavigable() {
				m.shiftDate(-1)
				m.allRows = nil
				m.status = "loading…"
				m.table = m.buildTable(nil)
				return m, m.loadCmd()
			}
		case ">":
			if m.isDateNavigable() {
				m.shiftDate(1)
				m.allRows = nil
				m.status = "loading…"
				m.table = m.buildTable(nil)
				return m, m.loadCmd()
			}
		default:
			var tCmd tea.Cmd
			m.table, tCmd = m.table.Update(msg)
			return m, tCmd
		}
	}
	return m, nil
}

func (m *Model) executeCommand(input string) tea.Cmd {
	cmd, arg, _ := strings.Cut(input, " ")
	arg = strings.TrimSpace(arg)

	switch cmd {
	case "sc", "schedule":
		m.switchView(viewSchedule)
	case "ev", "events":
		m.switchView(viewEvents)
	case "person", "p":
		if arg == "" {
			m.status = "usage: person <edookit-id>"
			return nil
		}
		m.personID = arg
		m.switchView(viewPerson)
	case "s", "students":
		m.switchView(viewStudents)
	case "emp", "employees":
		m.switchView(viewEmployees)
	case "l", "lessons":
		m.switchView(viewLessons)
	case "g", "grades":
		if arg == "" {
			m.status = "usage: grades <student-id>"
			return nil
		}
		m.gradeStudentID = arg
		m.switchView(viewGrades)
	case "c", "courses":
		m.switchView(viewCourses)
	case "pay", "payments":
		m.switchView(viewPayments)
	case "stats":
		m.switchView(viewStats)
	case "att", "attendance":
		if arg == "" {
			m.status = "usage: att <student-id>"
			return nil
		}
		m.attStudentID = arg
		m.switchView(viewAttendance)
	case "goals":
		if arg == "" {
			m.status = "usage: goals <student-id>"
			return nil
		}
		m.goalsStudentID = arg
		m.switchView(viewGoals)
	case "q", "quit":
		return tea.Quit
	default:
		m.status = fmt.Sprintf("unknown: %q  (sc · ev · person <id> · students · emp · grades <id> · courses · payments · stats · att <id> · goals <id>)", cmd)
		return nil
	}
	return m.loadCmd()
}

func (m *Model) switchView(v viewKind) {
	m.view = v
	m.allRows = nil
	m.filterText = ""
	m.filterInput.SetValue("")
	m.status = "loading…"
	m.table = m.buildTable(nil)
}

func (m *Model) savePrev() {
	m.prevView = m.view
	m.hasPrev = true
}

func (m *Model) goBack() tea.Cmd {
	m.view = m.prevView
	m.hasPrev = false
	m.allRows = nil
	m.filterText = ""
	m.filterInput.SetValue("")
	m.status = "loading…"
	m.table = m.buildTable(nil)
	return m.loadCmd()
}

// isDateNavigable reports whether the current view supports < / > date navigation.
func (m *Model) isDateNavigable() bool {
	switch m.view {
	case viewSchedule, viewEvents, viewGoals, viewAttendance:
		return true
	}
	return false
}

// isStudentContext reports whether the current view supports per-student navigation.
func (m *Model) isStudentContext() bool {
	switch m.view {
	case viewStudents, viewEmployees, viewPerson, viewCourses,
		viewGrades, viewAttendance, viewGoals:
		return true
	}
	return false
}

// selectedPersonID returns the student ID for the current context: from the
// selected row (col 0) for list views, or from the stored ID for detail views.
func (m *Model) selectedPersonID() string {
	switch m.view {
	case viewPerson:
		return m.personID
	case viewGrades:
		return m.gradeStudentID
	case viewAttendance:
		return m.attStudentID
	case viewGoals:
		return m.goalsStudentID
	}
	// viewStudents, viewEmployees, viewCourses — ID is in column 0
	row := m.table.SelectedRow()
	if len(row) == 0 {
		return ""
	}
	return row[0]
}

func (m *Model) handleEnter() tea.Cmd {
	switch m.view {
	case viewStudents, viewEmployees, viewCourses:
		id := m.selectedPersonID()
		if id == "" {
			return nil
		}
		m.savePrev()
		m.personID = id
		m.switchView(viewPerson)
		return m.loadCmd()
	case viewPerson:
		// Drill into a child / representative row.
		row := m.table.SelectedRow()
		if len(row) == 0 {
			return nil
		}
		id := row[0]
		if id == m.personID {
			return nil // top-level student row — already shown
		}
		m.savePrev()
		m.personID = id
		m.switchView(viewPerson)
		return m.loadCmd()
	}
	return nil
}

func (m *Model) navigateTo(v viewKind) tea.Cmd {
	id := m.selectedPersonID()
	if id == "" {
		return nil
	}
	m.savePrev()
	switch v {
	case viewGrades:
		m.gradeStudentID = id
	case viewAttendance:
		m.attStudentID = id
	case viewGoals:
		m.goalsStudentID = id
	}
	m.switchView(v)
	return m.loadCmd()
}

// shiftDate moves the active date range for the current view by delta units.
// Schedule shifts by weeks; Events and Goals by months; Attendance by days.
func (m *Model) shiftDate(delta int) {
	switch m.view {
	case viewSchedule:
		wf, _ := time.Parse("2006-01-02", m.weekFrom)
		wf = wf.AddDate(0, 0, delta*7)
		m.weekFrom = wf.Format("2006-01-02")
		m.weekTo = wf.AddDate(0, 0, 6).Format("2006-01-02")
	case viewEvents, viewGoals:
		mf, _ := time.Parse("2006-01-02", m.monthFrom)
		mf = mf.AddDate(0, delta, 0)
		ms := time.Date(mf.Year(), mf.Month(), 1, 0, 0, 0, 0, mf.Location())
		m.monthFrom = ms.Format("2006-01-02")
		m.monthTo = ms.AddDate(0, 1, -1).Format("2006-01-02")
	case viewAttendance:
		d, _ := time.Parse("2006-01-02", m.attDate)
		m.attDate = d.AddDate(0, 0, delta).Format("2006-01-02")
	}
}

func (m *Model) applyFilter() {
	if m.filterText == "" {
		m.table = m.buildTable(m.allRows)
		return
	}
	lower := strings.ToLower(m.filterText)
	var filtered []table.Row
	for _, row := range m.allRows {
		for _, cell := range row {
			if strings.Contains(strings.ToLower(cell), lower) {
				filtered = append(filtered, row)
				break
			}
		}
	}
	m.table = m.buildTable(filtered)
}

func (m Model) buildTable(rows []table.Row) table.Model {
	cols := m.columnsFor(m.view)
	tableHeight := m.height - 5
	if tableHeight < 3 {
		tableHeight = 3
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(colorPurpleDim)).
		BorderBottom(true).
		Background(lipgloss.Color(colorBg)).
		Foreground(lipgloss.Color(colorLavender)).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color(colorText)).
		Background(lipgloss.Color(colorSelected)).
		Bold(false)
	t.SetStyles(s)
	return t
}

func (m Model) columnsFor(v viewKind) []table.Column {
	w := m.width
	if w < 40 {
		w = 40
	}
	switch v {
	case viewSchedule:
		origW := (w - 12 - 8 - 10 - 6) / 2
		return []table.Column{
			{Title: "Date", Width: 7},
			{Title: "Time", Width: 6},
			{Title: "Original", Width: max(origW, 14)},
			{Title: "Change", Width: max(origW, 14)},
			{Title: "Class", Width: 10},
		}
	case viewEvents:
		titleW := w - 7 - 6 - 20 - 20 - 6
		return []table.Column{
			{Title: "Date", Width: 7},
			{Title: "Time", Width: 6},
			{Title: "Title", Width: max(titleW, 16)},
			{Title: "Place", Width: 20},
			{Title: "Classes", Width: 20},
		}
	case viewPerson:
		nameW := w - 6 - 10 - 8 - 28 - 6
		return []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Name", Width: max(nameW, 20)},
			{Title: "Role", Width: 10},
			{Title: "Class", Width: 8},
			{Title: "Email", Width: 28},
		}
	case viewStudents:
		nameW := w - 6 - 8 - 30 - 6
		return []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Name", Width: max(nameW, 15)},
			{Title: "Class", Width: 8},
			{Title: "Email", Width: 30},
		}
	case viewEmployees:
		nameW := w - 6 - 5 - 10 - 30 - 6
		return []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Name", Width: max(nameW, 15)},
			{Title: "Abbr", Width: 5},
			{Title: "Class(es)", Width: 10},
			{Title: "Email", Width: 30},
		}
	case viewLessons:
		nameW := w - 14 - 8 - 20 - 6
		return []table.Column{
			{Title: "Time", Width: 14},
			{Title: "Subject", Width: max(nameW, 15)},
			{Title: "Teacher", Width: 8},
			{Title: "Room", Width: 20},
		}
	case viewGrades:
		assignW := w - 14 - 6 - 22 - 12 - 6
		return []table.Column{
			{Title: "Course", Width: 14},
			{Title: "Grade", Width: 6},
			{Title: "Assignment", Width: max(assignW, 16)},
			{Title: "Evaluator", Width: 22},
			{Title: "Date", Width: 12},
		}
	case viewCourses:
		nameW := w - 5 - 20 - 6 - 6 - 6
		return []table.Column{
			{Title: "ID", Width: 5},
			{Title: "Student", Width: max(nameW, 14)},
			{Title: "Course", Width: 20},
			{Title: "Att%", Width: 6},
			{Title: "Grade", Width: 6},
		}
	case viewPayments:
		nameW := w - 6 - 12 - 14 - 8 - 6
		return []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Name", Width: max(nameW, 12)},
			{Title: "Due Date", Width: 12},
			{Title: "Spec.Symbol", Width: 14},
			{Title: "State", Width: 8},
		}
	case viewStats:
		return []table.Column{
			{Title: "Class", Width: 8},
			{Title: "Gender", Width: 7},
			{Title: "Grade", Width: 6},
			{Title: "Age", Width: 4},
			{Title: "Citizenship", Width: 12},
		}
	case viewAttendance:
		subjectW := w - 14 - 20 - 8 - 6
		return []table.Column{
			{Title: "Time", Width: 14},
			{Title: "Course", Width: max(subjectW, 12)},
			{Title: "Room", Width: 20},
			{Title: "Status", Width: 8},
		}
	case viewGoals:
		topicW := w - 10 - 12 - 20 - 6
		return []table.Column{
			{Title: "Time", Width: 10},
			{Title: "Course", Width: 12},
			{Title: "Subject", Width: 20},
			{Title: "Topic", Width: max(topicW, 16)},
		}
	}
	return nil
}

func (m Model) View() string {
	var b strings.Builder

	mode := "live"
	if m.mockMode {
		mode = "mock"
	}
	leftHeader := headerStyle.Render(" edookit ") +
		headerDimStyle.Render("┃") +
		headerDimStyle.Render(" "+m.baseURL+" ") +
		headerDimStyle.Render("┃") +
		headerDimStyle.Render(" "+mode+" ")
	version := headerDimStyle.Render("v0.1.0 ")
	pad := m.width - lipgloss.Width(leftHeader) - lipgloss.Width(version)
	if pad < 0 {
		pad = 0
	}
	b.WriteString(leftHeader + headerDimStyle.Render(strings.Repeat(" ", pad)) + version)
	b.WriteString("\n")

	viewTitle := m.view.String()
	if m.view == viewPerson && m.personID != "" {
		viewTitle += " #" + m.personID
	} else if m.view == viewGrades && m.gradeStudentID != "" {
		viewTitle += " student #" + m.gradeStudentID
	} else if m.view == viewAttendance && m.attStudentID != "" {
		viewTitle += " student #" + m.attStudentID + "  " + m.attDate
	} else if m.view == viewGoals && m.goalsStudentID != "" {
		viewTitle += " student #" + m.goalsStudentID
	} else if m.view == viewSchedule {
		viewTitle += "  " + m.weekFrom + " – " + m.weekTo
	} else if m.view == viewEvents {
		viewTitle += "  " + m.monthFrom + " – " + m.monthTo
	}
	count := fmt.Sprintf("[%d]", len(m.table.Rows()))
	filterHint := ""
	if m.filterText != "" {
		filterHint = titleCountStyle.Render("  /" + m.filterText)
	}
	titleLeft := titleBarStyle.Render(" "+viewTitle) + titleCountStyle.Render(" "+count) + filterHint
	titlePad := m.width - lipgloss.Width(titleLeft)
	if titlePad < 0 {
		titlePad = 0
	}
	b.WriteString(titleLeft + titleBarStyle.Render(strings.Repeat(" ", titlePad)))
	b.WriteString("\n")

	b.WriteString(m.table.View())
	b.WriteString("\n")

	if m.cmdMode {
		b.WriteString(cmdPromptStyle.Render(":") + " " + m.cmdInput.View())
	} else if m.filterMode {
		b.WriteString(cmdPromptStyle.Render("/") + " " + m.filterInput.View())
	} else if m.status != "" {
		if m.err != nil {
			b.WriteString(statusErrStyle.Render("  " + m.status))
		} else {
			b.WriteString(statusOkStyle.Render("  " + m.status))
		}
	} else {
		parts := []string{
			hintsStyle.Render("  "),
			hint("j/k", "move"),
		}
		switch m.view {
		case viewStudents, viewEmployees, viewCourses:
			parts = append(parts,
				hintsStyle.Render("  "), hint("enter", "person"),
				hintsStyle.Render("  "), hint("g", "grades"),
				hintsStyle.Render("  "), hint("a", "att"),
				hintsStyle.Render("  "), hint("i", "goals"),
			)
		case viewPerson:
			parts = append(parts,
				hintsStyle.Render("  "), hint("enter", "drill in"),
				hintsStyle.Render("  "), hint("g", "grades"),
				hintsStyle.Render("  "), hint("a", "att"),
				hintsStyle.Render("  "), hint("i", "goals"),
			)
		case viewGrades:
			parts = append(parts,
				hintsStyle.Render("  "), hint("a", "att"),
				hintsStyle.Render("  "), hint("i", "goals"),
			)
		case viewAttendance:
			parts = append(parts,
				hintsStyle.Render("  "), hint("< >", "prev/next day"),
				hintsStyle.Render("  "), hint("i", "goals"),
				hintsStyle.Render("  "), hint("g", "grades"),
			)
		case viewGoals:
			parts = append(parts,
				hintsStyle.Render("  "), hint("< >", "prev/next month"),
				hintsStyle.Render("  "), hint("a", "att"),
				hintsStyle.Render("  "), hint("g", "grades"),
			)
		case viewSchedule:
			parts = append(parts,
				hintsStyle.Render("  "), hint("< >", "prev/next week"),
			)
		case viewEvents:
			parts = append(parts,
				hintsStyle.Render("  "), hint("< >", "prev/next month"),
			)
		}
		if m.hasPrev {
			parts = append(parts, hintsStyle.Render("  "), hint("esc", "back"))
		}
		parts = append(parts,
			hintsStyle.Render("  "), hint("r", "refresh"),
			hintsStyle.Render("  "), hint("/", "filter"),
			hintsStyle.Render("  "), hint(":", "cmd"),
			hintsStyle.Render("  "), hint("q", "quit"),
		)
		b.WriteString(strings.Join(parts, ""))
	}

	return b.String()
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func firstTeacherAbbr(teachers map[string]edookit.TeacherRef) string {
	for _, t := range teachers {
		return t.PersonAbbr
	}
	return ""
}

func firstRoomName(rooms map[string]edookit.RoomRef) string {
	for _, r := range rooms {
		return r.RoomName
	}
	return ""
}

func firstGrade(evals []edookit.CourseFinalEvaluation) string {
	for _, e := range evals {
		if e.Grade != nil {
			return *e.Grade
		}
	}
	return "–"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func splitDateTime(r *edookit.ScheduleRange) (date, timeFrom string) {
	if r == nil {
		return "–", "–"
	}
	s := r.From
	if len(s) >= 16 {
		date = s[8:10] + "." + s[5:7]
		timeFrom = s[11:16]
	}
	return
}

// splitISODateTime parses "2026-05-21T08:00:00+02:00" → "21.05", "08:00"
func splitISODateTime(s string) (date, timeFrom string) {
	if len(s) < 16 {
		return s, ""
	}
	date = s[8:10] + "." + s[5:7]
	timeFrom = s[11:16]
	return
}

func joinStrings(ss []string) string {
	return strings.Join(ss, ", ")
}

func intVal(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func strValDefault(s *string, def string) string {
	if s == nil {
		return def
	}
	return *s
}

// parseTimerange extracts HH:MM start and end from a PostgreSQL range string
// like `["2026-05-21 08:00:00","2026-05-21 08:45:00")` or an ISO string.
func parseTimerange(s string) (from, to string) {
	if len(s) < 20 {
		return s, ""
	}
	// PostgreSQL range format: ["2026-05-21 08:00:00","2026-05-21 08:45:00")
	if s[0] == '[' || s[0] == '(' {
		s = s[1:] // strip leading bracket
		// find the comma separating the two timestamps
		// timestamps are quoted: "2026-05-21 08:00:00"
		if s[0] == '"' {
			s = s[1:]
			if len(s) >= 16 {
				from = s[11:16]
			}
			if i := strings.Index(s, `","`); i >= 0 {
				rest := s[i+3:]
				if len(rest) >= 16 {
					to = rest[11:16]
				}
			}
		}
		return
	}
	// ISO format fallback
	if len(s) >= 16 {
		from = s[11:16]
	}
	return
}

func scheduleActual(slot edookit.ScheduleSlot) string {
	if slot.Event != "" {
		return "→ " + slot.Event
	}
	if slot.Timerange == nil {
		return "cancelled"
	}
	parts := joinStrings(slot.Courses)
	if len(slot.Teachers) > 0 {
		parts += " " + joinStrings(slot.Teachers)
	}
	if len(slot.Rooms) > 0 {
		parts += " " + joinStrings(slot.Rooms)
	}
	return parts
}
