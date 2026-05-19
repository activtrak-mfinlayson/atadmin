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
// ListSignals
// ---------------------------------------------------------------------------

func TestListSignals(t *testing.T) {
	fixture := []api.Signal{
		{ID: 1, Name: "High CPU", Type: "system", Enabled: true},
		{ID: 2, Name: "After Hours", Type: "time", Enabled: false},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/admin/v1/signals" {
			t.Errorf("path = %s, want /admin/v1/signals", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fixture); err != nil {
			t.Errorf("encoding fixture: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	got, err := client.ListSignals(context.Background())
	if err != nil {
		t.Fatalf("ListSignals() error: %v", err)
	}
	if len(got) != len(fixture) {
		t.Fatalf("len(signals) = %d, want %d", len(got), len(fixture))
	}
	for i, want := range fixture {
		if got[i].ID != want.ID {
			t.Errorf("signals[%d].ID = %d, want %d", i, got[i].ID, want.ID)
		}
		if got[i].Name != want.Name {
			t.Errorf("signals[%d].Name = %q, want %q", i, got[i].Name, want.Name)
		}
		if got[i].Enabled != want.Enabled {
			t.Errorf("signals[%d].Enabled = %v, want %v", i, got[i].Enabled, want.Enabled)
		}
	}
}

func TestListSignals_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]api.Signal{}); err != nil {
			t.Errorf("encoding fixture: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	got, err := client.ListSignals(context.Background())
	if err != nil {
		t.Fatalf("ListSignals() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d signals", len(got))
	}
}

// ---------------------------------------------------------------------------
// CreateSignal
// ---------------------------------------------------------------------------

func TestCreateSignal(t *testing.T) {
	tests := []struct {
		name   string
		body   map[string]any
		wantID int
	}{
		{
			name:   "creates and returns id",
			body:   map[string]any{"name": "New Signal", "type": "system", "enabled": true},
			wantID: 42,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotBody map[string]any

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}
				if r.URL.Path != "/admin/v1/signal" {
					t.Errorf("path = %s, want /admin/v1/signal", r.URL.Path)
				}
				if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
					t.Errorf("decoding request body: %v", err)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				if err := json.NewEncoder(w).Encode(map[string]any{"id": tc.wantID}); err != nil {
					t.Errorf("encoding response: %v", err)
				}
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "tok")
			id, err := client.CreateSignal(context.Background(), tc.body)
			if err != nil {
				t.Fatalf("CreateSignal() error: %v", err)
			}
			if id != tc.wantID {
				t.Errorf("id = %d, want %d", id, tc.wantID)
			}
			if gotBody["name"] != tc.body["name"] {
				t.Errorf("request body name = %v, want %v", gotBody["name"], tc.body["name"])
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteSignal
// ---------------------------------------------------------------------------

func TestDeleteSignal(t *testing.T) {
	tests := []struct {
		name string
		id   int
	}{
		{name: "delete signal 1", id: 1},
		{name: "delete signal 99", id: 99},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/signals/%d", tc.id)

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

			client := newTestClient(t, server.URL, "tok")
			if err := client.DeleteSignal(context.Background(), tc.id); err != nil {
				t.Fatalf("DeleteSignal(%d) error: %v", tc.id, err)
			}
		})
	}
}
