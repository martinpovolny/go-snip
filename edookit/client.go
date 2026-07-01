package edookit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is an HTTP client for the Edookit API.
// It implements DataSource.
type Client struct {
	BaseURL       string
	LegacyBaseURL string // <school>-login.edookit.net for legacy modules (Změnový rozvrh)
	Username      string
	Password      string
	HTTPClient    *http.Client
}

// New creates a new Client.
// baseURL should be the school portal URL, e.g. "https://yourschool.edookit.net".
// The legacy base URL (for Změnový rozvrh) is derived automatically by inserting
// "-login" before ".edookit.net".
func New(baseURL, username, password string) *Client {
	base := strings.TrimRight(baseURL, "/")
	legacy := strings.Replace(base, ".edookit.net", "-login.edookit.net", 1)
	return &Client{
		BaseURL:       base,
		LegacyBaseURL: legacy,
		Username:      username,
		Password:      password,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// APIError represents an error returned by the Edookit API.
type APIError struct {
	Status  int
	Code    int
	Message string
}

func (e *APIError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("edookit: HTTP %d (code %d): %s", e.Status, e.Code, e.Message)
	}
	return fmt.Sprintf("edookit: HTTP %d: %s", e.Status, e.Message)
}

func (c *Client) get(path string, params url.Values, out any) error {
	return c.getFrom(c.BaseURL, path, params, out)
}

// getLegacy sends a GET to the legacy (-login) base URL used by Změnový rozvrh.
func (c *Client) getLegacy(path string, params url.Values, out any) error {
	return c.getFrom(c.LegacyBaseURL, path, params, out)
}

// getPublic sends an unauthenticated GET to the legacy domain (Veřejné události).
func (c *Client) getPublic(path string, params url.Values, out any) error {
	u := c.LegacyBaseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	return c.do(req, out)
}

func (c *Client) getFrom(base, path string, params url.Values, out any) error {
	u := base + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")

	return c.do(req, out)
}

func (c *Client) post(path string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		json.Unmarshal(body, &apiErr) //nolint:errcheck
		if apiErr.Message == "" {
			apiErr.Message = string(body)
		}
		return &APIError{Status: resp.StatusCode, Code: apiErr.Code, Message: apiErr.Message}
	}

	if out != nil && len(body) > 0 {
		return json.Unmarshal(body, out)
	}
	return nil
}
