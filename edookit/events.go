package edookit

import "net/url"

// ListPublicEvents returns school public events for the given date range.
// No authentication required — served at the legacy -login.edookit.net domain.
func (c *Client) ListPublicEvents(opts PublicEventsOpts) ([]PublicEvent, error) {
	params := url.Values{}
	if opts.From != "" {
		params.Set("from", opts.From)
	}
	if opts.To != "" {
		params.Set("to", opts.To)
	}

	var resp PublicEventsResponse
	if err := c.getPublic("/api/public/v1/events", params, &resp); err != nil {
		return nil, err
	}
	return resp.Events, nil
}
