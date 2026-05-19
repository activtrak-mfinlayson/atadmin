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
// ListDevices
// ---------------------------------------------------------------------------

func TestListDevices(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		fixture  []api.Device
	}{
		{
			name:     "single page returns decoded slice",
			page:     1,
			pageSize: 25,
			fixture: []api.Device{
				{ID: 1, Hostname: "WORKSTATION-01", Status: "active"},
				{ID: 2, Hostname: "LAPTOP-02", Status: "inactive"},
			},
		},
		{
			name:     "empty result set",
			page:     99,
			pageSize: 10,
			fixture:  []api.Device{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotPage, gotPageSize string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				if r.URL.Path != "/admin/v1/devices" {
					t.Errorf("path = %s, want /admin/v1/devices", r.URL.Path)
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
			got, err := client.ListDevices(context.Background(), tc.page, tc.pageSize)
			if err != nil {
				t.Fatalf("ListDevices() error: %v", err)
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
				t.Fatalf("len(devices) = %d, want %d", len(got), len(tc.fixture))
			}
			for i, want := range tc.fixture {
				if got[i].ID != want.ID {
					t.Errorf("devices[%d].ID = %d, want %d", i, got[i].ID, want.ID)
				}
				if got[i].Hostname != want.Hostname {
					t.Errorf("devices[%d].Hostname = %q, want %q", i, got[i].Hostname, want.Hostname)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteDevices
// ---------------------------------------------------------------------------

func TestDeleteDevices(t *testing.T) {
	tests := []struct {
		name string
		ids  []int
	}{
		{name: "delete single device", ids: []int{1}},
		{name: "delete multiple devices", ids: []int{10, 20, 30}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("method = %s, want DELETE", r.Method)
				}
				if r.URL.Path != "/admin/v1/devices" {
					t.Errorf("path = %s, want /admin/v1/devices", r.URL.Path)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", ct)
				}

				var body struct {
					DeviceIDs []int `json:"deviceIds"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decoding request body: %v", err)
				}
				if len(body.DeviceIDs) != len(tc.ids) {
					t.Errorf("len(deviceIds) = %d, want %d", len(body.DeviceIDs), len(tc.ids))
				}
				for i, id := range tc.ids {
					if body.DeviceIDs[i] != id {
						t.Errorf("deviceIds[%d] = %d, want %d", i, body.DeviceIDs[i], id)
					}
				}

				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "test-token")
			if err := client.DeleteDevices(context.Background(), tc.ids); err != nil {
				t.Fatalf("DeleteDevices() error: %v", err)
			}
		})
	}
}
