package edookit

import (
	"fmt"
	"net/url"
)

func (c *Client) ListCourses(opts CourseListOpts) ([]StudentWithCourses, error) {
	params := url.Values{}
	if opts.PersonID != nil {
		params.Set("person_id", fmt.Sprint(*opts.PersonID))
	}
	if opts.EvaltermID != nil {
		params.Set("evalterm_id", fmt.Sprint(*opts.EvaltermID))
	}

	var resp CoursesResponse
	if err := c.get("/api/course-data/v1/courses", params, &resp); err != nil {
		return nil, err
	}
	return resp.Students, nil
}
