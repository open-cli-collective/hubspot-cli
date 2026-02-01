package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultBaseURL is the base URL for HubSpot API
	DefaultBaseURL = "https://api.hubapi.com"
)

// Client is a HubSpot API client
type Client struct {
	BaseURL     string
	AccessToken string
	HTTPClient  *http.Client
	Verbose     bool
}

// ClientConfig contains configuration for creating a new client
type ClientConfig struct {
	AccessToken string
	Verbose     bool
}

// New creates a new HubSpot API client from config
func New(cfg ClientConfig) (*Client, error) {
	if cfg.AccessToken == "" {
		return nil, ErrAccessTokenRequired
	}

	return &Client{
		BaseURL:     DefaultBaseURL,
		AccessToken: cfg.AccessToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Verbose: cfg.Verbose,
	}, nil
}

// authHeader returns the Bearer auth header value
func (c *Client) authHeader() string {
	return "Bearer " + c.AccessToken
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, urlStr string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, urlStr, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.Verbose {
		fmt.Printf("→ %s %s\n", method, urlStr)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if c.Verbose {
		fmt.Printf("← %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	if resp.StatusCode >= 400 {
		return nil, ParseAPIError(resp, respBody)
	}

	return respBody, nil
}

// get performs a GET request
func (c *Client) get(urlStr string) ([]byte, error) {
	return c.doRequest(http.MethodGet, urlStr, nil)
}

// post performs a POST request
func (c *Client) post(urlStr string, body interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPost, urlStr, body)
}

// patch performs a PATCH request
func (c *Client) patch(urlStr string, body interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPatch, urlStr, body)
}

// put performs a PUT request
func (c *Client) put(urlStr string, body interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPut, urlStr, body)
}

// delete performs a DELETE request
func (c *Client) delete(urlStr string) ([]byte, error) {
	return c.doRequest(http.MethodDelete, urlStr, nil)
}

// buildURL builds a URL with query parameters
func buildURL(base string, params map[string]string) string {
	if len(params) == 0 {
		return base
	}

	u, _ := url.Parse(base)
	q := u.Query()
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}
