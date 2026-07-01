package edookit

func (c *Client) ListStudentStats(date string) ([]StudentStat, error) {
	path := "/api/stats/v1/students"
	if date != "" {
		path += "/" + date
	}
	var resp StudentStatsResponse
	if err := c.get(path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Students, nil
}
