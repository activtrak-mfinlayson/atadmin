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
// ListUsers
// ---------------------------------------------------------------------------

func TestListUsers(t *testing.T) {
	fixture := api.UsersPage{
		Results: []api.Identity{
			{ID: 1, Revision: 2, Tracked: true, Status: "active"},
			{ID: 2, Revision: 1, Tracked: false, Status: "inactive"},
		},
		TotalCount: 2,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/identity/v1/entities" {
			t.Errorf("path = %s, want /identity/v1/entities", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fixture); err != nil {
			t.Errorf("encoding fixture: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	got, err := client.ListUsers(context.Background(), api.IdentityListParams{})
	if err != nil {
		t.Fatalf("ListUsers() error: %v", err)
	}
	if got.TotalCount != fixture.TotalCount {
		t.Errorf("TotalCount = %d, want %d", got.TotalCount, fixture.TotalCount)
	}
	if len(got.Results) != len(fixture.Results) {
		t.Fatalf("len(Results) = %d, want %d", len(got.Results), len(fixture.Results))
	}
	if got.Results[0].ID != 1 {
		t.Errorf("Results[0].ID = %d, want 1", got.Results[0].ID)
	}
	if got.Results[0].Tracked != true {
		t.Errorf("Results[0].Tracked = %v, want true", got.Results[0].Tracked)
	}
}

func TestListUsers_WithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("search") != "alice" {
			t.Errorf("search = %q, want %q", q.Get("search"), "alice")
		}
		if q.Get("searchType") != "email" {
			t.Errorf("searchType = %q, want %q", q.Get("searchType"), "email")
		}
		if q.Get("sortDirection") != "asc" {
			t.Errorf("sortDirection = %q, want %q", q.Get("sortDirection"), "asc")
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(api.UsersPage{}); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	_, err := client.ListUsers(context.Background(), api.IdentityListParams{
		Search:     "alice",
		SearchType: "email",
		SortDir:    "asc",
	})
	if err != nil {
		t.Fatalf("ListUsers() error: %v", err)
	}
}

func TestListUsers_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(api.UsersPage{Results: []api.Identity{}}); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	got, err := client.ListUsers(context.Background(), api.IdentityListParams{})
	if err != nil {
		t.Fatalf("ListUsers() error: %v", err)
	}
	if len(got.Results) != 0 {
		t.Errorf("expected empty Results, got %d", len(got.Results))
	}
}

// ---------------------------------------------------------------------------
// GetUser
// ---------------------------------------------------------------------------

func TestGetUser(t *testing.T) {
	displayName := api.IdentityField{Value: "Alice Smith", ID: "dn-1", Source: "Admin"}
	fixture := api.Identity{
		ID:          12345,
		Revision:    3,
		DisplayName: &displayName,
		Tracked:     true,
		Status:      "active",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/identity/v1/entities/12345" {
			t.Errorf("path = %s, want /identity/v1/entities/12345", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fixture); err != nil {
			t.Errorf("encoding fixture: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	got, err := client.GetUser(context.Background(), 12345)
	if err != nil {
		t.Fatalf("GetUser() error: %v", err)
	}
	if got.ID != 12345 {
		t.Errorf("ID = %d, want 12345", got.ID)
	}
	if got.Revision != 3 {
		t.Errorf("Revision = %d, want 3", got.Revision)
	}
	if got.DisplayName == nil || got.DisplayName.Value != "Alice Smith" {
		t.Errorf("DisplayName = %v, want Alice Smith", got.DisplayName)
	}
	if !got.Tracked {
		t.Errorf("Tracked = false, want true")
	}
}

func TestGetUser_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]any{"message": "not found"}); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	_, err := client.GetUser(context.Background(), 99999)
	if err == nil {
		t.Fatal("GetUser() expected error for 404, got nil")
	}
}

// ---------------------------------------------------------------------------
// UpdateUser
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/identity/v1/entities/12345/revision/3" {
			t.Errorf("path = %s, want /identity/v1/entities/12345/revision/3", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		updated := api.Identity{ID: 12345, Revision: 4}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(updated); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	name := "Alice Updated"
	client := newTestClient(t, server.URL, "tok")
	got, err := client.UpdateUser(context.Background(), 12345, 3, api.UpdateUserRequest{
		DisplayName: &name,
	})
	if err != nil {
		t.Fatalf("UpdateUser() error: %v", err)
	}
	if got.Revision != 4 {
		t.Errorf("Revision = %d, want 4", got.Revision)
	}
	dn, ok := gotBody["displayName"].(map[string]any)
	if !ok || dn["value"] != "Alice Updated" {
		t.Errorf("request body displayName = %v, want {value: Alice Updated}", gotBody["displayName"])
	}
}

func TestUpdateUser_Conflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		if err := json.NewEncoder(w).Encode(map[string]any{"message": "revision conflict"}); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	name := "X"
	client := newTestClient(t, server.URL, "tok")
	_, err := client.UpdateUser(context.Background(), 1, 0, api.UpdateUserRequest{DisplayName: &name})
	if err == nil {
		t.Fatal("UpdateUser() expected 409 error, got nil")
	}
	if got := err.Error(); len(got) == 0 {
		t.Error("error message is empty")
	}
}

// ---------------------------------------------------------------------------
// DeleteUser
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/identity/v1/entities/12345" {
			t.Errorf("path = %s, want /identity/v1/entities/12345", r.URL.Path)
		}
		if r.URL.Query().Get("revision") != "3" {
			t.Errorf("revision query param = %q, want %q", r.URL.Query().Get("revision"), "3")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	if err := client.DeleteUser(context.Background(), 12345, 3); err != nil {
		t.Fatalf("DeleteUser() error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// AddUserGroups / RemoveUserGroups
// ---------------------------------------------------------------------------

func TestAddUserGroups(t *testing.T) {
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/identity/v1/entities/12345/groups" {
			t.Errorf("path = %s, want /identity/v1/entities/12345/groups", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		updated := api.Identity{ID: 12345, Revision: 4}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(updated); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	got, err := client.AddUserGroups(context.Background(), 12345, []int{42, 43}, 3)
	if err != nil {
		t.Fatalf("AddUserGroups() error: %v", err)
	}
	if got.Revision != 4 {
		t.Errorf("Revision = %d, want 4", got.Revision)
	}
	if gotBody["revision"] != float64(3) {
		t.Errorf("revision in body = %v, want 3", gotBody["revision"])
	}
}

func TestRemoveUserGroups(t *testing.T) {
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/identity/v1/entities/12345/groups" {
			t.Errorf("path = %s, want /identity/v1/entities/12345/groups", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		updated := api.Identity{ID: 12345, Revision: 4}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(updated); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	got, err := client.RemoveUserGroups(context.Background(), 12345, []int{42}, 3)
	if err != nil {
		t.Fatalf("RemoveUserGroups() error: %v", err)
	}
	if got.Revision != 4 {
		t.Errorf("Revision = %d, want 4", got.Revision)
	}
	if gotBody["revision"] != float64(3) {
		t.Errorf("revision in body = %v, want 3", gotBody["revision"])
	}
}

// ---------------------------------------------------------------------------
// BulkAction
// ---------------------------------------------------------------------------

func TestBulkAction_AllSuccess(t *testing.T) {
	var gotBody api.BulkActionRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/identity/v1/entities/bulk" {
			t.Errorf("path = %s, want /identity/v1/entities/bulk", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		result := api.BulkActionResponse{
			Successful: []api.BulkEntitySuccess{{EntityID: 1}, {EntityID: 2}},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	req := api.BulkActionRequest{
		Actions: []string{"StopTracking"},
		Data: []api.BulkEntityData{
			{EntityID: 1, Revision: 2},
			{EntityID: 2, Revision: 5},
		},
	}
	client := newTestClient(t, server.URL, "tok")
	got, err := client.BulkAction(context.Background(), req)
	if err != nil {
		t.Fatalf("BulkAction() error: %v", err)
	}
	if len(got.Successful) != 2 {
		t.Errorf("Successful = %d, want 2", len(got.Successful))
	}
	if len(got.Failures) != 0 {
		t.Errorf("Failures = %d, want 0", len(got.Failures))
	}
	if len(gotBody.Actions) != 1 || gotBody.Actions[0] != "StopTracking" {
		t.Errorf("actions = %v, want [StopTracking]", gotBody.Actions)
	}
}

func TestBulkAction_PartialFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := api.BulkActionResponse{
			Successful: []api.BulkEntitySuccess{{EntityID: 1}},
			Failures:   []api.BulkEntityFailure{{EntityID: 2, Message: "revision conflict"}},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	req := api.BulkActionRequest{
		Actions: []string{"StartTracking"},
		Data: []api.BulkEntityData{
			{EntityID: 1, Revision: 1},
			{EntityID: 2, Revision: 99},
		},
	}
	client := newTestClient(t, server.URL, "tok")
	got, err := client.BulkAction(context.Background(), req)
	if err != nil {
		t.Fatalf("BulkAction() error: %v", err)
	}
	if len(got.Successful) != 1 {
		t.Errorf("Successful = %d, want 1", len(got.Successful))
	}
	if len(got.Failures) != 1 {
		t.Errorf("Failures = %d, want 1", len(got.Failures))
	}
	if got.Failures[0].EntityID != 2 {
		t.Errorf("Failures[0].EntityID = %d, want 2", got.Failures[0].EntityID)
	}
}

// ---------------------------------------------------------------------------
// ListAgents
// ---------------------------------------------------------------------------

func TestListAgents(t *testing.T) {
	fixture := api.UsersPage{
		Results:    []api.Identity{{ID: 99, Tracked: true}},
		TotalCount: 1,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/identity/v1/agents" {
			t.Errorf("path = %s, want /identity/v1/agents", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fixture); err != nil {
			t.Errorf("encoding fixture: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	got, err := client.ListAgents(context.Background(), api.IdentityListParams{})
	if err != nil {
		t.Fatalf("ListAgents() error: %v", err)
	}
	if got.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", got.TotalCount)
	}
}

func TestListAgents_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(api.UsersPage{Results: []api.Identity{}}); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	got, err := client.ListAgents(context.Background(), api.IdentityListParams{})
	if err != nil {
		t.Fatalf("ListAgents() error: %v", err)
	}
	if len(got.Results) != 0 {
		t.Errorf("expected empty Results, got %d", len(got.Results))
	}
}

// ---------------------------------------------------------------------------
// FetchRevisions helper
// ---------------------------------------------------------------------------

func TestFetchRevisions(t *testing.T) {
	entities := map[int64]int64{1: 5, 2: 3}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var id int64
		if _, err := fmt.Sscanf(r.URL.Path, "/identity/v1/entities/%d", &id); err != nil {
			t.Errorf("parsing path %s: %v", r.URL.Path, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		rev, ok := entities[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		identity := api.Identity{ID: id, Revision: rev}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(identity); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	revs, errs := client.FetchRevisions(context.Background(), []int64{1, 2}, 5)
	if len(errs) != 0 {
		t.Fatalf("FetchRevisions() errors: %v", errs)
	}
	if revs[1] != 5 {
		t.Errorf("revision for id=1 = %d, want 5", revs[1])
	}
	if revs[2] != 3 {
		t.Errorf("revision for id=2 = %d, want 3", revs[2])
	}
}
