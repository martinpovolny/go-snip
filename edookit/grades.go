package edookit

import (
	"fmt"
	"net/url"
)

func (c *Client) ListEvaluations(opts EvalListOpts) ([]Evaluation, error) {
	params := url.Values{}
	if opts.StudentID != nil {
		params.Set("student_id", fmt.Sprint(*opts.StudentID))
	}
	if opts.CourseID != nil {
		params.Set("course_id", fmt.Sprint(*opts.CourseID))
	}
	if opts.EvaltermID != nil {
		params.Set("evalterm_id", fmt.Sprint(*opts.EvaltermID))
	}
	if opts.DateFrom != nil {
		params.Set("date_from", *opts.DateFrom)
	}
	if opts.DateTo != nil {
		params.Set("date_to", *opts.DateTo)
	}

	var out []Evaluation
	if err := c.get("/api/evaluation/v1/list", params, &out); err != nil {
		return nil, err
	}
	return out, nil
}
