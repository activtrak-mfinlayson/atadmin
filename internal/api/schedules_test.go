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
// ListSchedules
// ---------------------------------------------------------------------------

func TestListSchedules(t *testing.T) {
	fixture := []api.Schedule{
		{ID: 1, Name: "Standard Week", Type: "reporting", IsDefault: true},
		{ID: 2, Name: "Night Shift", Type: "shift", IsDefault: false},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/admin/v1/schedules" {
			t.Errorf("path = %s, want /admin/v1/schedules", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fixture); err != nil {
			t.Errorf("encoding fixture: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	got, err := client.ListSchedules(context.Background())
	if err != nil {
		t.Fatalf("ListSchedules() error: %v", err)
	}
	if len(got) != len(fixture) {
		t.Fatalf("len(schedules) = %d, want %d", len(got), len(fixture))
	}
	for i, want := range fixture {
		if got[i].ID != want.ID {
			t.Errorf("schedules[%d].ID = %d, want %d", i, got[i].ID, want.ID)
		}
		if got[i].Name != want.Name {
			t.Errorf("schedules[%d].Name = %q, want %q", i, got[i].Name, want.Name)
		}
		if got[i].IsDefault != want.IsDefault {
			t.Errorf("schedules[%d].IsDefault = %v, want %v", i, got[i].IsDefault, want.IsDefault)
		}
	}
}

// ---------------------------------------------------------------------------
// GetSchedule
// ---------------------------------------------------------------------------

func TestGetSchedule(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		fixture api.Schedule
	}{
		{
			name:    "existing schedule",
			id:      3,
			fixture: api.Schedule{ID: 3, Name: "Day Shift", Type: "shift", IsDefault: false},
		},
		{
			name:    "default schedule",
			id:      1,
			fixture: api.Schedule{ID: 1, Name: "Standard Week", Type: "reporting", IsDefault: true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/schedules/%d", tc.id)

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
			got, err := client.GetSchedule(context.Background(), tc.id)
			if err != nil {
				t.Fatalf("GetSchedule(%d) error: %v", tc.id, err)
			}
			if got.ID != tc.fixture.ID {
				t.Errorf("ID = %d, want %d", got.ID, tc.fixture.ID)
			}
			if got.Name != tc.fixture.Name {
				t.Errorf("Name = %q, want %q", got.Name, tc.fixture.Name)
			}
			if got.IsDefault != tc.fixture.IsDefault {
				t.Errorf("IsDefault = %v, want %v", got.IsDefault, tc.fixture.IsDefault)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SetScheduleUsers
// ---------------------------------------------------------------------------

func TestSetScheduleUsers(t *testing.T) {
	tests := []struct {
		name       string
		scheduleID int
		userIDs    []int
	}{
		{name: "assign two users", scheduleID: 5, userIDs: []int{101, 102}},
		{name: "clear users", scheduleID: 7, userIDs: []int{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/schedules/%d/users", tc.scheduleID)
			var gotBody struct {
				UserIDs []int `json:"userIds"`
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("method = %s, want PUT", r.Method)
				}
				if r.URL.Path != wantPath {
					t.Errorf("path = %s, want %s", r.URL.Path, wantPath)
				}
				if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
					t.Errorf("decoding request body: %v", err)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "tok")
			if err := client.SetScheduleUsers(context.Background(), tc.scheduleID, tc.userIDs); err != nil {
				t.Fatalf("SetScheduleUsers() error: %v", err)
			}
			if len(gotBody.UserIDs) != len(tc.userIDs) {
				t.Errorf("userIds len = %d, want %d", len(gotBody.UserIDs), len(tc.userIDs))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetUserReportingSchedule
// ---------------------------------------------------------------------------

func TestGetUserReportingSchedule(t *testing.T) {
	tests := []struct {
		name    string
		userID  int
		fixture api.Schedule
	}{
		{
			name:    "user with reporting schedule",
			userID:  42,
			fixture: api.Schedule{ID: 1, Name: "Standard Week", Type: "reporting", IsDefault: true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/user/%d/schedule/reporting", tc.userID)

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
			got, err := client.GetUserReportingSchedule(context.Background(), tc.userID)
			if err != nil {
				t.Fatalf("GetUserReportingSchedule(%d) error: %v", tc.userID, err)
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
