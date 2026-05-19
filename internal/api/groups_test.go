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
// ListGroups
// ---------------------------------------------------------------------------

func TestListGroups(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		fixture  []api.Group
	}{
		{
			name:     "single page returns decoded slice",
			page:     1,
			pageSize: 25,
			fixture: []api.Group{
				{ID: 1, Name: "Engineering", MemberCount: 10},
				{ID: 2, Name: "Marketing", MemberCount: 5},
			},
		},
		{
			name:     "empty result set",
			page:     99,
			pageSize: 10,
			fixture:  []api.Group{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotPage, gotPageSize string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				if r.URL.Path != "/admin/v1/groups/list" {
					t.Errorf("path = %s, want /admin/v1/groups/list", r.URL.Path)
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
			got, err := client.ListGroups(context.Background(), tc.page, tc.pageSize)
			if err != nil {
				t.Fatalf("ListGroups() error: %v", err)
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
				t.Fatalf("len(groups) = %d, want %d", len(got), len(tc.fixture))
			}
			for i, want := range tc.fixture {
				if got[i].ID != want.ID {
					t.Errorf("groups[%d].ID = %d, want %d", i, got[i].ID, want.ID)
				}
				if got[i].Name != want.Name {
					t.Errorf("groups[%d].Name = %q, want %q", i, got[i].Name, want.Name)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CreateGroup
// ---------------------------------------------------------------------------

func TestCreateGroup(t *testing.T) {
	tests := []struct {
		name    string
		group   string
		wantID  int
	}{
		{name: "creates group and returns ID", group: "NewTeam", wantID: 42},
		{name: "creates another group", group: "SalesTeam", wantID: 99},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/groups/%s", tc.group)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}
				if r.URL.Path != wantPath {
					t.Errorf("path = %s, want %s", r.URL.Path, wantPath)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				if err := json.NewEncoder(w).Encode(map[string]int{"id": tc.wantID}); err != nil {
					t.Errorf("encoding response: %v", err)
				}
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "test-token")
			gotID, err := client.CreateGroup(context.Background(), tc.group)
			if err != nil {
				t.Fatalf("CreateGroup() error: %v", err)
			}
			if gotID != tc.wantID {
				t.Errorf("CreateGroup() ID = %d, want %d", gotID, tc.wantID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AddMembers
// ---------------------------------------------------------------------------

func TestAddMembers(t *testing.T) {
	tests := []struct {
		name       string
		groupID    int
		memberID   int
		memberType string
	}{
		{name: "add client member", groupID: 10, memberID: 55, memberType: "client"},
		{name: "add device member", groupID: 20, memberID: 77, memberType: "device"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}
				if r.URL.Path != "/admin/v1/groups/members" {
					t.Errorf("path = %s, want /admin/v1/groups/members", r.URL.Path)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", ct)
				}

				var body struct {
					GroupID    int    `json:"groupId"`
					MemberID   int    `json:"memberId"`
					MemberType string `json:"memberType"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decoding body: %v", err)
				}
				if body.GroupID != tc.groupID {
					t.Errorf("groupId = %d, want %d", body.GroupID, tc.groupID)
				}
				if body.MemberID != tc.memberID {
					t.Errorf("memberId = %d, want %d", body.MemberID, tc.memberID)
				}
				if body.MemberType != tc.memberType {
					t.Errorf("memberType = %q, want %q", body.MemberType, tc.memberType)
				}

				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "test-token")
			if err := client.AddMembers(context.Background(), tc.groupID, tc.memberID, tc.memberType); err != nil {
				t.Fatalf("AddMembers() error: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ListGroupMembers
// ---------------------------------------------------------------------------

func TestListGroupMembers(t *testing.T) {
	tests := []struct {
		name    string
		groupID int
		fixture []api.GroupMember
	}{
		{
			name:    "members of group 5",
			groupID: 5,
			fixture: []api.GroupMember{
				{GroupID: 5, MemberID: 100, MemberType: "client"},
				{GroupID: 5, MemberID: 200, MemberType: "device"},
			},
		},
		{
			name:    "empty group",
			groupID: 99,
			fixture: []api.GroupMember{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/groups/%d/members", tc.groupID)

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
			got, err := client.ListGroupMembers(context.Background(), tc.groupID)
			if err != nil {
				t.Fatalf("ListGroupMembers(%d) error: %v", tc.groupID, err)
			}

			if len(got) != len(tc.fixture) {
				t.Fatalf("len(members) = %d, want %d", len(got), len(tc.fixture))
			}
			for i, want := range tc.fixture {
				if got[i].MemberID != want.MemberID {
					t.Errorf("members[%d].MemberID = %d, want %d", i, got[i].MemberID, want.MemberID)
				}
				if got[i].MemberType != want.MemberType {
					t.Errorf("members[%d].MemberType = %q, want %q", i, got[i].MemberType, want.MemberType)
				}
			}
		})
	}
}
