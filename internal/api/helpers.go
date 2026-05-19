package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// doRequest constructs and executes an HTTP request against the client's BaseURL.
// If body is non-nil it is marshalled to JSON and sent as the request body with
// Content-Type: application/json. The caller is responsible for closing the
// response body.
func (c *Client) doRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshalling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	// Parse the path (which may include a query string) relative to BaseURL.
	rel, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("parsing path %q: %w", path, err)
	}
	u := c.BaseURL.ResolveReference(rel)
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request %s %s: %w", method, path, err)
	}
	return resp, nil
}

// checkResponse examines the HTTP status code and returns an actionable error
// for any non-2xx response. It reads (and closes) the body to extract an API
// message when available.
func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Attempt to parse an API-provided message.
	var apiMsg string
	if resp.Body != nil {
		defer func() { _ = resp.Body.Close() }()
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Message != "" {
			apiMsg = errResp.Message
		}
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		if apiMsg != "" {
			return fmt.Errorf("bad request: %s", apiMsg)
		}
		return fmt.Errorf("bad request: the server rejected the request")
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized: your token may have expired. Try running 'atadmin auth login'")
	case http.StatusForbidden:
		return fmt.Errorf("forbidden: your account role may not have permission for this operation")
	case http.StatusNotFound:
		return fmt.Errorf("not found: the requested resource does not exist")
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limited: retried multiple times. Wait a moment and try again")
	default:
		if resp.StatusCode >= 500 {
			return fmt.Errorf("server error (%d): try again later", resp.StatusCode)
		}
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}
