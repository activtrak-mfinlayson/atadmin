package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/activtrak-mfinlayson/atadmin/internal/api"
)

func TestAuthRoundTripperInjectsToken(t *testing.T) {
	var gotAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := api.NewClient(server.URL, "mytoken", "0.1.0")
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	resp, err := client.HTTPClient.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("GET error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	want := "Bearer mytoken"
	if gotAuth != want {
		t.Errorf("Authorization = %q, want %q", gotAuth, want)
	}
}

func TestAuthRoundTripperDoesNotMutateOriginal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := api.NewClient(server.URL, "tok", "0.1.0")
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	original, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := client.HTTPClient.Do(original)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Original request must not have been mutated with auth header.
	if original.Header.Get("Authorization") != "" {
		t.Error("original request was mutated by authRoundTripper")
	}
}

func TestNewClientUserAgent(t *testing.T) {
	client, err := api.NewClient("https://api.activtrak.com", "token", "0.1.0")
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	if !strings.HasPrefix(client.UserAgent, "atadmin/") {
		t.Errorf("UserAgent = %q, want prefix %q", client.UserAgent, "atadmin/")
	}
	if client.UserAgent != "atadmin/0.1.0" {
		t.Errorf("UserAgent = %q, want %q", client.UserAgent, "atadmin/0.1.0")
	}
}

func TestNewClientInvalidURL(t *testing.T) {
	_, err := api.NewClient("://bad-url", "tok", "0.1.0")
	if err == nil {
		t.Error("NewClient() with invalid URL should return error")
	}
}
