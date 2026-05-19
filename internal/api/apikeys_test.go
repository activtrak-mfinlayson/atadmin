package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/activtrak-mfinlayson/atadmin/internal/api"
)

// ---------------------------------------------------------------------------
// ListAPIKeys
// ---------------------------------------------------------------------------

func TestListAPIKeys(t *testing.T) {
	fixture := []api.ApiKey{
		{ID: 1, Name: "prod-key", KeyPrefix: "at_pro", LastUsedAt: "2024-01-15"},
		{ID: 2, Name: "dev-key", KeyPrefix: "at_dev", LastUsedAt: ""},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/admin/v1/key" {
			t.Errorf("path = %s, want /admin/v1/key", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fixture); err != nil {
			t.Errorf("encoding fixture: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "test-token")
	got, err := client.ListAPIKeys(context.Background())
	if err != nil {
		t.Fatalf("ListAPIKeys() error: %v", err)
	}

	if len(got) != len(fixture) {
		t.Fatalf("len(keys) = %d, want %d", len(got), len(fixture))
	}
	for i, want := range fixture {
		if got[i].ID != want.ID {
			t.Errorf("keys[%d].ID = %d, want %d", i, got[i].ID, want.ID)
		}
		if got[i].Name != want.Name {
			t.Errorf("keys[%d].Name = %q, want %q", i, got[i].Name, want.Name)
		}
	}
}

func TestListAPIKeys_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "test-token")
	got, err := client.ListAPIKeys(context.Background())
	if err != nil {
		t.Fatalf("ListAPIKeys() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len(keys) = %d, want 0", len(got))
	}
}

// ---------------------------------------------------------------------------
// CreateAPIKey
// ---------------------------------------------------------------------------

func TestCreateAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		keyName string
		fixture api.ApiKey
	}{
		{
			name:    "creates key and returns object",
			keyName: "my-new-key",
			fixture: api.ApiKey{ID: 42, Name: "my-new-key", KeyPrefix: "at_myn", LastUsedAt: ""},
		},
		{
			name:    "empty name allowed",
			keyName: "",
			fixture: api.ApiKey{ID: 7, Name: "", KeyPrefix: "at_xxx"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}
				if r.URL.Path != "/admin/v1/key" {
					t.Errorf("path = %s, want /admin/v1/key", r.URL.Path)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", ct)
				}

				var body struct {
					Name string `json:"name"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decoding request body: %v", err)
				}
				if body.Name != tc.keyName {
					t.Errorf("body.name = %q, want %q", body.Name, tc.keyName)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				if err := json.NewEncoder(w).Encode(tc.fixture); err != nil {
					t.Errorf("encoding fixture: %v", err)
				}
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "test-token")
			got, err := client.CreateAPIKey(context.Background(), tc.keyName)
			if err != nil {
				t.Fatalf("CreateAPIKey(%q) error: %v", tc.keyName, err)
			}
			if got.ID != tc.fixture.ID {
				t.Errorf("ID = %d, want %d", got.ID, tc.fixture.ID)
			}
			if got.Name != tc.fixture.Name {
				t.Errorf("Name = %q, want %q", got.Name, tc.fixture.Name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteAPIKey
// ---------------------------------------------------------------------------

func TestDeleteAPIKey(t *testing.T) {
	tests := []struct {
		name string
		id   int
	}{
		{name: "delete existing key", id: 10},
		{name: "delete another key", id: 99},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/key/%d", tc.id)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("method = %s, want DELETE", r.Method)
				}
				if r.URL.Path != wantPath {
					t.Errorf("path = %s, want %s", r.URL.Path, wantPath)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "test-token")
			if err := client.DeleteAPIKey(context.Background(), tc.id); err != nil {
				t.Fatalf("DeleteAPIKey(%d) error: %v", tc.id, err)
			}
		})
	}
}
