package edookit

import (
	"fmt"
	"net/url"
)

// ListAttendanceLessons returns lessons and their attendance state for a student on a given date.
// date is YYYY-MM-DD; omit to use today.
func (c *Client) ListAttendanceLessons(studentID int, date string) ([]AttendanceLesson, error) {
	params := url.Values{"student_id": {fmt.Sprint(studentID)}}
	if date != "" {
		params.Set("date", date)
	}
	var out []AttendanceLesson
	if err := c.get("/api/attendance/v3/lessons", params, &out); err != nil {
		return nil, err
	}
	return out, nil
}
