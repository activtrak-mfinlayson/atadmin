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
// ListAuditLogs
// ---------------------------------------------------------------------------

func TestListAuditLogs_WithParams(t *testing.T) {
	fixture := []api.AuditLog{
		{ID: 1, Action: "login", Actor: "alice", Timestamp: "2024-01-01T10:00:00Z"},
		{ID: 2, Action: "logout", Actor: "bob", Timestamp: "2024-01-01T11:00:00Z"},
	}

	var gotFrom, gotPage string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/admin/v1/auditlog" {
			t.Errorf("path = %s, want /admin/v1/auditlog", r.URL.Path)
		}
		gotFrom = r.URL.Query().Get("FromDate")
		gotPage = r.URL.Query().Get("Page")

		// Serve actual wire format: {"data": [{auditid, eventname, user, time, ...}]}
		type wireItem struct {
			ID     int    `json:"auditid"`
			Action string `json:"eventname"`
			Actor  string `json:"user"`
			Time   string `json:"time"`
		}
		items := make([]wireItem, len(fixture))
		for i, e := range fixture {
			items[i] = wireItem{ID: e.ID, Action: e.Action, Actor: e.Actor, Time: e.Timestamp}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"totalCount": len(items), "data": items}); err != nil {
			t.Errorf("encoding fixture: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "test-token")
	got, err := client.ListAuditLogs(
		context.Background(),
		"2024-01-01", // from
		"2024-01-31", // to
		"",           // filters
		"",           // sortCol
		false,        // sortDesc
		5,            // page
		25,           // pageSize
	)
	if err != nil {
		t.Fatalf("ListAuditLogs() error: %v", err)
	}

	// FromDate and Page should be set.
	if gotFrom != "2024-01-01" {
		t.Errorf("FromDate query param = %q, want %q", gotFrom, "2024-01-01")
	}
	if gotPage != "5" {
		t.Errorf("Page query param = %q, want %q", gotPage, "5")
	}

	if len(got) != len(fixture) {
		t.Fatalf("len(logs) = %d, want %d", len(got), len(fixture))
	}
	for i, want := range fixture {
		if got[i].ID != want.ID {
			t.Errorf("logs[%d].ID = %d, want %d", i, got[i].ID, want.ID)
		}
		if got[i].Action != want.Action {
			t.Errorf("logs[%d].Action = %q, want %q", i, got[i].Action, want.Action)
		}
	}
}

func TestListAuditLogs_EmptyParams(t *testing.T) {
	var capturedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"totalCount":0,"data":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "test-token")
	_, err := client.ListAuditLogs(
		context.Background(),
		"", // from — empty, should not be set
		"", // to
		"", // filters
		"", // sortCol
		false,
		0, // page — zero, should not be set
		0, // pageSize — zero
	)
	if err != nil {
		t.Fatalf("ListAuditLogs() error: %v", err)
	}

	// When all params are empty/zero, query string should be empty.
	if capturedQuery != "" {
		t.Errorf("expected no query params when all inputs are empty, got: %q", capturedQuery)
	}
}

func TestListAuditLogs_PageSetNotFrom(t *testing.T) {
	var gotFrom, gotPage string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFrom = r.URL.Query().Get("FromDate")
		gotPage = r.URL.Query().Get("Page")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"totalCount":0,"data":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "test-token")
	_, err := client.ListAuditLogs(
		context.Background(),
		"", // from — empty
		"",
		"",
		"",
		false,
		3,  // page set
		10, // pageSize set
	)
	if err != nil {
		t.Fatalf("ListAuditLogs() error: %v", err)
	}

	// Page should be set, FromDate should not.
	if gotPage != "3" {
		t.Errorf("Page query param = %q, want %q", gotPage, "3")
	}
	if gotFrom != "" {
		t.Errorf("FromDate query param = %q, want empty (not sent)", gotFrom)
	}
}

// ---------------------------------------------------------------------------
// GetAttachment
// ---------------------------------------------------------------------------

func TestGetAttachment(t *testing.T) {
	tests := []struct {
		name         string
		attachmentID string
		content      []byte
	}{
		{name: "pdf attachment", attachmentID: "abc-123", content: []byte("%PDF-1.4 binary data")},
		{name: "another attachment", attachmentID: "xyz-789", content: []byte("raw data here")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantPath := fmt.Sprintf("/admin/v1/attachment/%s", tc.attachmentID)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				if r.URL.Path != wantPath {
					t.Errorf("path = %s, want %s", r.URL.Path, wantPath)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(tc.content)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "test-token")
			got, err := client.GetAttachment(context.Background(), tc.attachmentID)
			if err != nil {
				t.Fatalf("GetAttachment(%q) error: %v", tc.attachmentID, err)
			}
			if string(got) != string(tc.content) {
				t.Errorf("body = %q, want %q", string(got), string(tc.content))
			}
		})
	}
}
