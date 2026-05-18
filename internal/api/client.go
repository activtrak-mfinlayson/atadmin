package api

import (
	"fmt"
	"net/http"
	"net/url"
)

// Client is the atadmin HTTP API client.
type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
	UserAgent  string
}

// authRoundTripper injects a Bearer token into every outbound request.
type authRoundTripper struct {
	token string
	inner http.RoundTripper
}

// RoundTrip clones the request, sets the Authorization header, and delegates
// to the inner transport. The original request is never mutated.
func (a *authRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	req := r.Clone(r.Context())
	req.Header.Set("Authorization", "Bearer "+a.token)
	return a.inner.RoundTrip(req)
}

// NewClient constructs a Client that authenticates every request with token.
func NewClient(baseURL, token, version string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL %q: %w", baseURL, err)
	}

	transport := &authRoundTripper{
		token: token,
		inner: http.DefaultTransport,
	}

	return &Client{
		BaseURL: u,
		HTTPClient: &http.Client{
			Transport: transport,
		},
		UserAgent: "atadmin/" + version,
	}, nil
}
