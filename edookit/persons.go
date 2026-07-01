package edookit

import "fmt"

// SearchPerson looks up a person by their Edookit ID or Plus4U ID.
// criterion must be a numeric Edookit ID (e.g. "237") or a Plus4U ID (e.g. "123-4432-1").
// Returns up to one person plus their children and representatives.
func (c *Client) SearchPerson(criterion string) ([]Person, error) {
	var out []Person
	if err := c.get("/api/person-search/v1/"+criterion, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPerson looks up a person by their integer Edookit ID.
func (c *Client) GetPerson(edookitID int) ([]Person, error) {
	return c.SearchPerson(fmt.Sprint(edookitID))
}
