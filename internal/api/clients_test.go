package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/activtrak-mfinlayson/atadmin/internal/api"
)

// ---------------------------------------------------------------------------
// ListClients
// ---------------------------------------------------------------------------

func TestListClients(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		fixture  []api.ATClient
	}{
		{
			name:     "single page returns decoded slice",
			page:     1,
			pageSize: 25,
			fixture: []api.ATClient{
				{ID: 1, Username: "alice", LogonDomain: "corp", Alias: "Alice A", Status: "active", DeviceCount: 2},
				{ID: 2, Username: "bob", LogonDomain: "corp", Alias: "Bob B", Status: "inactive", DeviceCount: 0},
			},
		},
		{
			name:     "empty result set",
			page:     99,
			pageSize: 10,
			fixture:  []api.ATClient{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotPage, gotPageSize string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				if r.URL.Path != "/admin/v1/clients" {
					t.Errorf("path = %s, want /admin/v1/clients", r.URL.Path)
				}
				gotPage = r.URL.Query().Get("Page")
				gotPageSize = r.URL.Query().Get("PageSize")

				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(tc.fixture); err != nil {
					t.Errorf("encoding fixture: %v", err)
				}
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "test-token")
			got, err := client.ListClients(context.Background(), tc.page, tc.pageSize)
			if err != nil {
				t.Fatalf("ListClients() error: %v", err)
			}

			// Verify query parameters were forwarded.
			wantPage := fmt.Sprintf("%d", tc.page)
			wantPageSize := fmt.Sprintf("%d", tc.pageSize)
			if gotPage != wantPage {
				t.Errorf("Page query param = %q, want %q", gotPage, wantPage)
			}
			if gotPageSize != wantPageSize {
				t.Errorf("PageSize query param = %q, want %q", gotPageSize, wantPageSize)
			}

			// Verify decoded length and content.
			if len(got) != len(tc.fixture) {
				t.Fatalf("len(clients) = %d, want %d", len(got), len(tc.fixture))
			}
			for i, want := range tc.fixture {
				if got[i].ID != want.ID {
					t.Errorf("clients[%d].ID = %d, want %d", i, got[i].ID, want.ID)
				}
				if got[i].Username != want.Username {
					t.Errorf("clients[%d].Username = %q, want %q", i, got[i].Username, want.Username)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetClientByID
// ---------------------------------------------------------------------------

func TestGetClientByID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		fixture api.ATClient
	}{
		{
			name:    "existing client",
			id:      42,
			fixture: api.ATClient{ID: 42, Username: "carol", LogonDomain: "acme", Status: "active", DeviceCount: 3},
		},
		{
			name:    "another client",
			id:      7,
			fixture: api.ATClient{ID: 7, Username: "dave", LogonDomain: "acme", Status: "inactive"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/clients/%d", tc.id)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				if r.URL.Path != wantPath {
					t.Errorf("path = %s, want %s", r.URL.Path, wantPath)
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(tc.fixture); err != nil {
					t.Errorf("encoding fixture: %v", err)
				}
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "test-token")
			got, err := client.GetClientByID(context.Background(), tc.id)
			if err != nil {
				t.Fatalf("GetClientByID(%d) error: %v", tc.id, err)
			}
			if got.ID != tc.fixture.ID {
				t.Errorf("ID = %d, want %d", got.ID, tc.fixture.ID)
			}
			if got.Username != tc.fixture.Username {
				t.Errorf("Username = %q, want %q", got.Username, tc.fixture.Username)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateClientAlias
// ---------------------------------------------------------------------------

func TestUpdateClientAlias(t *testing.T) {
	tests := []struct {
		name  string
		id    int
		alias string
	}{
		{name: "set alias", id: 10, alias: "New Alias"},
		{name: "clear alias", id: 20, alias: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/clients/%d", tc.id)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("method = %s, want PUT", r.Method)
				}
				if r.URL.Path != wantPath {
					t.Errorf("path = %s, want %s", r.URL.Path, wantPath)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", ct)
				}

				var body struct {
					Alias string `json:"alias"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decoding request body: %v", err)
				}
				if body.Alias != tc.alias {
					t.Errorf("body.alias = %q, want %q", body.Alias, tc.alias)
				}

				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "test-token")
			if err := client.UpdateClientAlias(context.Background(), tc.id, tc.alias); err != nil {
				t.Fatalf("UpdateClientAlias() error: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ListDoNotTrack
// ---------------------------------------------------------------------------

func TestListDoNotTrack(t *testing.T) {
	fixture := []api.DNTEntry{
		{ID: 1, LogonDomain: "corp", Username: "eve", IsGlobal: false},
		{ID: 2, LogonDomain: "", Username: "", IsGlobal: true},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/admin/v1/clients/donottrack" {
			t.Errorf("path = %s, want /admin/v1/clients/donottrack", r.URL.Path)
		}
		// Serve actual wire format: {"records": [...]} with logondomain/globaluser fields.
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"records": fixture}); err != nil {
			t.Errorf("encoding fixture: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "test-token")
	got, err := client.ListDoNotTrack(context.Background())
	if err != nil {
		t.Fatalf("ListDoNotTrack() error: %v", err)
	}
	if len(got) != len(fixture) {
		t.Fatalf("len(entries) = %d, want %d", len(got), len(fixture))
	}
	for i, want := range fixture {
		if got[i].ID != want.ID {
			t.Errorf("entries[%d].ID = %d, want %d", i, got[i].ID, want.ID)
		}
		if got[i].IsGlobal != want.IsGlobal {
			t.Errorf("entries[%d].IsGlobal = %v, want %v", i, got[i].IsGlobal, want.IsGlobal)
		}
	}
}

// ---------------------------------------------------------------------------
// AddDoNotTrack
// ---------------------------------------------------------------------------

func TestAddDoNotTrack(t *testing.T) {
	tests := []struct {
		name   string
		domain string
		user   string
	}{
		{name: "domain and user", domain: "corp.example.com", user: "frank"},
		{name: "user only", domain: "", user: "grace"},
		{name: "domain only", domain: "example.com", user: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}
				if r.URL.Path != "/admin/v1/clients/donottrack" {
					t.Errorf("path = %s, want /admin/v1/clients/donottrack", r.URL.Path)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", ct)
				}

				var body struct {
					LogonDomain string `json:"logonDomain"`
					Username    string `json:"username"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decoding body: %v", err)
				}
				if body.LogonDomain != tc.domain {
					t.Errorf("logonDomain = %q, want %q", body.LogonDomain, tc.domain)
				}
				if body.Username != tc.user {
					t.Errorf("username = %q, want %q", body.Username, tc.user)
				}

				w.WriteHeader(http.StatusCreated)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "test-token")
			if err := client.AddDoNotTrack(context.Background(), tc.domain, tc.user); err != nil {
				t.Fatalf("AddDoNotTrack() error: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// checkResponse error mapping
// ---------------------------------------------------------------------------

func TestCheckResponseErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantSubstr string
	}{
		{
			name:       "400 with API message",
			statusCode: http.StatusBadRequest,
			body:       `{"message":"invalid input"}`,
			wantSubstr: "invalid input",
		},
		{
			name:       "401 unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{}`,
			wantSubstr: "atadmin auth login",
		},
		{
			name:       "403 forbidden",
			statusCode: http.StatusForbidden,
			body:       `{}`,
			wantSubstr: "forbidden",
		},
		{
			name:       "404 not found",
			statusCode: http.StatusNotFound,
			body:       `{}`,
			wantSubstr: "not found",
		},
		{
			name:       "429 rate limited",
			statusCode: http.StatusTooManyRequests,
			body:       `{}`,
			wantSubstr: "rate limited",
		},
		{
			name:       "503 server error",
			statusCode: http.StatusServiceUnavailable,
			body:       `{}`,
			wantSubstr: "server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "test-token")
			_, err := client.ListClients(context.Background(), 1, 10)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantSubstr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tc.wantSubstr)
			}
		})
	}
}
