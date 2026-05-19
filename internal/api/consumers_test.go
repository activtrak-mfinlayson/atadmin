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
// ListConsumers
// ---------------------------------------------------------------------------

func TestListConsumers(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		fixture  []api.Consumer
	}{
		{
			name:     "single page returns decoded slice",
			page:     1,
			pageSize: 25,
			fixture: []api.Consumer{
				{ID: 1, Username: "admin@example.com", Role: "admin", UseSSO: false},
				{ID: 2, Username: "viewer@example.com", Role: "viewer", UseSSO: true},
			},
		},
		{
			name:     "empty result set",
			page:     99,
			pageSize: 10,
			fixture:  []api.Consumer{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotPage, gotPageSize string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				if r.URL.Path != "/admin/v1/consumers" {
					t.Errorf("path = %s, want /admin/v1/consumers", r.URL.Path)
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
			got, err := client.ListConsumers(context.Background(), tc.page, tc.pageSize)
			if err != nil {
				t.Fatalf("ListConsumers() error: %v", err)
			}

			wantPage := fmt.Sprintf("%d", tc.page)
			wantPageSize := fmt.Sprintf("%d", tc.pageSize)
			if gotPage != wantPage {
				t.Errorf("Page query param = %q, want %q", gotPage, wantPage)
			}
			if gotPageSize != wantPageSize {
				t.Errorf("PageSize query param = %q, want %q", gotPageSize, wantPageSize)
			}

			if len(got) != len(tc.fixture) {
				t.Fatalf("len(consumers) = %d, want %d", len(got), len(tc.fixture))
			}
			for i, want := range tc.fixture {
				if got[i].ID != want.ID {
					t.Errorf("consumers[%d].ID = %d, want %d", i, got[i].ID, want.ID)
				}
				if got[i].Username != want.Username {
					t.Errorf("consumers[%d].Username = %q, want %q", i, got[i].Username, want.Username)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetConsumer
// ---------------------------------------------------------------------------

func TestGetConsumer(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		fixture api.Consumer
	}{
		{
			name:    "existing consumer",
			id:      42,
			fixture: api.Consumer{ID: 42, Username: "carol@example.com", Role: "admin", UseSSO: true},
		},
		{
			name:    "another consumer",
			id:      7,
			fixture: api.Consumer{ID: 7, Username: "dave@example.com", Role: "viewer", UseSSO: false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/consumers/%d", tc.id)

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
			got, err := client.GetConsumer(context.Background(), tc.id)
			if err != nil {
				t.Fatalf("GetConsumer(%d) error: %v", tc.id, err)
			}
			if got.ID != tc.fixture.ID {
				t.Errorf("ID = %d, want %d", got.ID, tc.fixture.ID)
			}
			if got.Username != tc.fixture.Username {
				t.Errorf("Username = %q, want %q", got.Username, tc.fixture.Username)
			}
			if got.Role != tc.fixture.Role {
				t.Errorf("Role = %q, want %q", got.Role, tc.fixture.Role)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SetConsumerRole
// ---------------------------------------------------------------------------

func TestSetConsumerRole(t *testing.T) {
	tests := []struct {
		name string
		id   int
		role string
	}{
		{name: "set admin role", id: 10, role: "admin"},
		{name: "set viewer role", id: 20, role: "viewer"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/consumers/%d/role", tc.id)

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
					Role string `json:"role"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decoding request body: %v", err)
				}
				if body.Role != tc.role {
					t.Errorf("body.role = %q, want %q", body.Role, tc.role)
				}

				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "test-token")
			if err := client.SetConsumerRole(context.Background(), tc.id, tc.role); err != nil {
				t.Fatalf("SetConsumerRole() error: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteConsumers
// ---------------------------------------------------------------------------

func TestDeleteConsumers(t *testing.T) {
	tests := []struct {
		name string
		ids  []int
	}{
		{name: "delete single consumer", ids: []int{1}},
		{name: "delete multiple consumers", ids: []int{10, 20, 30}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("method = %s, want DELETE", r.Method)
				}
				if r.URL.Path != "/admin/v1/consumers" {
					t.Errorf("path = %s, want /admin/v1/consumers", r.URL.Path)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", ct)
				}

				var body struct {
					IDs []int `json:"ids"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decoding request body: %v", err)
				}
				if len(body.IDs) != len(tc.ids) {
					t.Errorf("len(ids) = %d, want %d", len(body.IDs), len(tc.ids))
				}

				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "test-token")
			if err := client.DeleteConsumers(context.Background(), tc.ids); err != nil {
				t.Fatalf("DeleteConsumers() error: %v", err)
			}
		})
	}
}
