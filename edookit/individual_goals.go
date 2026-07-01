package edookit

import (
	"fmt"
	"net/url"
)

// ListIndividualLessonContent returns per-lesson individual curriculum for students.
// At least one of studentID (>0), lessonID (>0), or dateFrom must be set.
func (c *Client) ListIndividualLessonContent(studentID, lessonID int, dateFrom, dateTo string) ([]IndividualLessonContent, error) {
	params := url.Values{}
	if studentID > 0 {
		params.Set("student_id", fmt.Sprint(studentID))
	}
	if lessonID > 0 {
		params.Set("lesson_id", fmt.Sprint(lessonID))
	}
	if dateFrom != "" {
		params.Set("date_from", dateFrom)
	}
	if dateTo != "" {
		params.Set("date_to", dateTo)
	}
	var out []IndividualLessonContent
	if err := c.get("/api/individual-goals/v1/individual-lesson-content", params, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListLearningAgreements returns student learning agreements (výukové plány).
func (c *Client) ListLearningAgreements(studentID int, dateFrom, dateTo string) ([]LearningAgreement, error) {
	params := url.Values{}
	if studentID > 0 {
		params.Set("student_id", fmt.Sprint(studentID))
	}
	if dateFrom != "" {
		params.Set("date_from", dateFrom)
	}
	if dateTo != "" {
		params.Set("date_to", dateTo)
	}
	var out []LearningAgreement
	if err := c.get("/api/individual-goals/v1/learning-agreements", params, &out); err != nil {
		return nil, err
	}
	return out, nil
}
