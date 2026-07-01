package edookit

import "net/url"

// ListScheduleChanges returns timetable changes for the given date range.
// Uses the legacy -login.edookit.net base URL.
func (c *Client) ListScheduleChanges(opts ScheduleChangeOpts) ([]ScheduleChange, error) {
	params := url.Values{}
	if opts.From != "" {
		params.Set("from", opts.From)
	}
	if opts.To != "" {
		params.Set("to", opts.To)
	}
	if opts.TeacherNameFormat != "" {
		params.Set("teacherNameFormat", opts.TeacherNameFormat)
	}

	var resp ScheduleChangesResponse
	if err := c.getLegacy("/api/scheduler/v1/change", params, &resp); err != nil {
		return nil, err
	}
	return resp.Change, nil
}
