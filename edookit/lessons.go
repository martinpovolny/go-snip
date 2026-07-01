package edookit

import (
	"fmt"
	"net/url"
)

func (c *Client) ListLessons(date string, opts LessonListOpts) (map[string]LessonEntry, error) {
	params := url.Values{"date": {date}}
	if opts.CourseID != nil {
		params.Set("course_id", fmt.Sprint(*opts.CourseID))
	}
	if opts.CourseTypeID != nil {
		params.Set("course_type_id", fmt.Sprint(*opts.CourseTypeID))
	}
	if opts.RoomID != nil {
		params.Set("room_id", fmt.Sprint(*opts.RoomID))
	}
	if opts.StudentPersonID != nil {
		params.Set("student_person_id", fmt.Sprint(*opts.StudentPersonID))
	}
	if opts.TeacherPersonID != nil {
		params.Set("teacher_person_id", fmt.Sprint(*opts.TeacherPersonID))
	}
	if opts.WorkTypeID != nil {
		params.Set("work_type_id", fmt.Sprint(*opts.WorkTypeID))
	}

	var resp LessonsResponse
	if err := c.get("/api/lesson/v2/list-lessons", params, &resp); err != nil {
		return nil, err
	}
	return resp.Lessons, nil
}

func (c *Client) ListRooms() (map[string]Room, error) {
	var resp RoomsResponse
	if err := c.get("/api/lesson/v2/lists/rooms", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Rooms, nil
}

func (c *Client) ListCourseTypes() (map[string]CourseType, error) {
	var resp CourseTypesResponse
	if err := c.get("/api/lesson/v2/lists/types/course", nil, &resp); err != nil {
		return nil, err
	}
	return resp.CourseTypes, nil
}

func (c *Client) ListWorkTypes() (map[string]WorkType, error) {
	var resp WorkTypesResponse
	if err := c.get("/api/lesson/v2/lists/types/work", nil, &resp); err != nil {
		return nil, err
	}
	return resp.WorkTypes, nil
}
