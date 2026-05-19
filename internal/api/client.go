package api

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"
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

func (a *authRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	req := r.Clone(r.Context())
	req.Header.Set("Authorization", "Bearer "+a.token)
	return a.inner.RoundTrip(req)
}

// retryRoundTripper retries on HTTP 429 with exponential backoff.
type retryRoundTripper struct {
	inner    http.RoundTripper
	maxRetry int
}

func (r *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)
	for attempt := 0; attempt <= r.maxRetry; attempt++ {
		resp, err = r.inner.RoundTrip(req)
		if err != nil || resp.StatusCode != http.StatusTooManyRequests {
			return resp, err
		}
		if attempt < r.maxRetry {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			time.Sleep(backoff)
		}
	}
	return resp, err
}

// verboseRoundTripper logs request/response info to a writer when enabled.
type verboseRoundTripper struct {
	inner   http.RoundTripper
	out     io.Writer
	enabled bool
}

func (v *verboseRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if v.enabled {
		_, _ = fmt.Fprintf(v.out, "> %s %s\n", req.Method, req.URL)
	}
	resp, err := v.inner.RoundTrip(req)
	if v.enabled && err == nil {
		_, _ = fmt.Fprintf(v.out, "< %s\n", resp.Status)
	}
	return resp, err
}

// NewClient constructs a Client with auth injection, retry-on-429, and optional
// verbose logging to stderr.
func NewClient(baseURL, token, version string, verbose bool, verboseOut io.Writer) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL %q: %w", baseURL, err)
	}

	auth := &authRoundTripper{token: token, inner: http.DefaultTransport}
	retry := &retryRoundTripper{inner: auth, maxRetry: 3}
	verbose_ := &verboseRoundTripper{inner: retry, out: verboseOut, enabled: verbose}

	return &Client{
		BaseURL:  u,
		HTTPClient: &http.Client{Transport: verbose_},
		UserAgent: "atadmin/" + version,
	}, nil
}
