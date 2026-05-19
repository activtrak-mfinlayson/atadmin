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
// ListAlarms
// ---------------------------------------------------------------------------

func TestListAlarms(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		fixture  []api.Alarm
	}{
		{
			name:     "returns decoded slice",
			page:     1,
			pageSize: 25,
			fixture: []api.Alarm{
				{ID: 10, Name: "Alarm A", Type: "activity", Enabled: true},
				{ID: 11, Name: "Alarm B", Type: "website", Enabled: false},
			},
		},
		{
			name:     "empty page",
			page:     5,
			pageSize: 10,
			fixture:  []api.Alarm{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotPage, gotPageSize string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				if r.URL.Path != "/admin/v1/alarms" {
					t.Errorf("path = %s, want /admin/v1/alarms", r.URL.Path)
				}
				gotPage = r.URL.Query().Get("page")
				gotPageSize = r.URL.Query().Get("pageSize")
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(tc.fixture); err != nil {
					t.Errorf("encoding fixture: %v", err)
				}
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "tok")
			got, err := client.ListAlarms(context.Background(), tc.page, tc.pageSize)
			if err != nil {
				t.Fatalf("ListAlarms() error: %v", err)
			}

			wantPage := fmt.Sprintf("%d", tc.page)
			wantPageSize := fmt.Sprintf("%d", tc.pageSize)
			if gotPage != wantPage {
				t.Errorf("page param = %q, want %q", gotPage, wantPage)
			}
			if gotPageSize != wantPageSize {
				t.Errorf("pageSize param = %q, want %q", gotPageSize, wantPageSize)
			}
			if len(got) != len(tc.fixture) {
				t.Fatalf("len(alarms) = %d, want %d", len(got), len(tc.fixture))
			}
			for i, want := range tc.fixture {
				if got[i].ID != want.ID {
					t.Errorf("alarms[%d].ID = %d, want %d", i, got[i].ID, want.ID)
				}
				if got[i].Name != want.Name {
					t.Errorf("alarms[%d].Name = %q, want %q", i, got[i].Name, want.Name)
				}
				if got[i].Enabled != want.Enabled {
					t.Errorf("alarms[%d].Enabled = %v, want %v", i, got[i].Enabled, want.Enabled)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetAlarm
// ---------------------------------------------------------------------------

func TestGetAlarm(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		fixture api.Alarm
	}{
		{
			name:    "existing alarm",
			id:      7,
			fixture: api.Alarm{ID: 7, Name: "After Hours", Type: "time", Enabled: true},
		},
		{
			name:    "another alarm",
			id:      99,
			fixture: api.Alarm{ID: 99, Name: "High Activity", Type: "activity", Enabled: false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/alarms/%d", tc.id)

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

			client := newTestClient(t, server.URL, "tok")
			got, err := client.GetAlarm(context.Background(), tc.id)
			if err != nil {
				t.Fatalf("GetAlarm(%d) error: %v", tc.id, err)
			}
			if got.ID != tc.fixture.ID {
				t.Errorf("ID = %d, want %d", got.ID, tc.fixture.ID)
			}
			if got.Name != tc.fixture.Name {
				t.Errorf("Name = %q, want %q", got.Name, tc.fixture.Name)
			}
			if got.Enabled != tc.fixture.Enabled {
				t.Errorf("Enabled = %v, want %v", got.Enabled, tc.fixture.Enabled)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteAlarm
// ---------------------------------------------------------------------------

func TestDeleteAlarm(t *testing.T) {
	tests := []struct {
		name string
		id   int
	}{
		{name: "delete alarm 1", id: 1},
		{name: "delete alarm 55", id: 55},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/alarms/%d", tc.id)

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
			if err := client.DeleteAlarm(context.Background(), tc.id); err != nil {
				t.Fatalf("DeleteAlarm(%d) error: %v", tc.id, err)
			}
		})
	}
}
