package edookit

import "net/url"

func (c *Client) ListStudents(opts StudentDataOpts) ([]Student, error) {
	path := "/api/student-data/v1/list"
	if opts.Date != "" {
		path += "/" + opts.Date
	}
	params := url.Values{}
	if opts.IncludeInactiveSince != "" {
		params.Set("include-inactive-since", opts.IncludeInactiveSince)
	}
	var resp StudentsResponse
	if err := c.get(path, params, &resp); err != nil {
		return nil, err
	}
	return resp.Students, nil
}

func (c *Client) ListEmployees(opts StudentDataOpts) ([]Employee, error) {
	path := "/api/employee-data/v1/list"
	if opts.Date != "" {
		path += "/" + opts.Date
	}
	params := url.Values{}
	if opts.IncludeInactiveSince != "" {
		params.Set("include-inactive-since", opts.IncludeInactiveSince)
	}
	var resp EmployeesResponse
	if err := c.get(path, params, &resp); err != nil {
		return nil, err
	}
	return resp.Employees, nil
}
