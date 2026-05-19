package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// AccountPing
// ---------------------------------------------------------------------------

func TestAccountPing(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{name: "success 200", statusCode: http.StatusOK, wantErr: false},
		{name: "success 204", statusCode: http.StatusNoContent, wantErr: false},
		{name: "unauthorized", statusCode: http.StatusUnauthorized, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "tok")
			err := client.AccountPing(context.Background())

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.wantErr && gotPath != "/admin/v1/accounts/ping" {
				t.Errorf("path = %q, want /admin/v1/accounts/ping", gotPath)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetPrivacySettings
// ---------------------------------------------------------------------------

func TestGetPrivacySettings(t *testing.T) {
	fixture := map[string]any{
		"screenshotEnabled": true,
		"videoEnabled":      false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/admin/v1/accountsettings/privacy" {
			t.Errorf("path = %s, want /admin/v1/accountsettings/privacy", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fixture); err != nil {
			t.Errorf("encoding fixture: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	got, err := client.GetPrivacySettings(context.Background())
	if err != nil {
		t.Fatalf("GetPrivacySettings() error: %v", err)
	}
	if got["screenshotEnabled"] != true {
		t.Errorf("screenshotEnabled = %v, want true", got["screenshotEnabled"])
	}
	if got["videoEnabled"] != false {
		t.Errorf("videoEnabled = %v, want false", got["videoEnabled"])
	}
}

// ---------------------------------------------------------------------------
// UpdatePrivacySettings
// ---------------------------------------------------------------------------

func TestUpdatePrivacySettings(t *testing.T) {
	body := map[string]any{"screenshotEnabled": false}

	var gotMethod string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		if r.URL.Path != "/admin/v1/accountsettings/privacy" {
			t.Errorf("path = %s, want /admin/v1/accountsettings/privacy", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	if err := client.UpdatePrivacySettings(context.Background(), body); err != nil {
		t.Fatalf("UpdatePrivacySettings() error: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method = %s, want PUT", gotMethod)
	}
	if gotBody["screenshotEnabled"] != false {
		t.Errorf("screenshotEnabled = %v, want false", gotBody["screenshotEnabled"])
	}
}

// ---------------------------------------------------------------------------
// GetSSOEnabled
// ---------------------------------------------------------------------------

func TestGetSSOEnabled(t *testing.T) {
	tests := []struct {
		name    string
		fixture bool
	}{
		{name: "enabled true", fixture: true},
		{name: "enabled false", fixture: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				if r.URL.Path != "/admin/v1/accountsettings/sso/enabled" {
					t.Errorf("path = %s, want /admin/v1/accountsettings/sso/enabled", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(tc.fixture); err != nil {
					t.Errorf("encoding fixture: %v", err)
				}
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, "tok")
			got, err := client.GetSSOEnabled(context.Background())
			if err != nil {
				t.Fatalf("GetSSOEnabled() error: %v", err)
			}
			if got != tc.fixture {
				t.Errorf("GetSSOEnabled() = %v, want %v", got, tc.fixture)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetRoleAccess
// ---------------------------------------------------------------------------

func TestGetRoleAccess(t *testing.T) {
	fixture := []map[string]any{
		{"resource": "screenshots", "roles": []any{"admin", "supervisor"}},
		{"resource": "reports", "roles": []any{"admin"}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/admin/v1/accountsettings/roleaccess" {
			t.Errorf("path = %s, want /admin/v1/accountsettings/roleaccess", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fixture); err != nil {
			t.Errorf("encoding fixture: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "tok")
	got, err := client.GetRoleAccess(context.Background())
	if err != nil {
		t.Fatalf("GetRoleAccess() error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0]["resource"] != "screenshots" {
		t.Errorf("got[0].resource = %v, want screenshots", got[0]["resource"])
	}
}
