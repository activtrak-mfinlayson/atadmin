package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/activtrak-mfinlayson/atadmin/internal/output"
)

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	output.Table(&buf, []string{"NAME", "STATUS"}, [][]string{
		{"alice", "active"},
		{"bob", "inactive"},
	})
	got := buf.String()
	if !strings.Contains(got, "NAME") || !strings.Contains(got, "alice") {
		t.Errorf("unexpected table output: %q", got)
	}
}

func TestKeyValue(t *testing.T) {
	var buf bytes.Buffer
	output.KeyValue(&buf, map[string]string{
		"username": "alice",
		"status":   "active",
	})
	got := buf.String()
	if !strings.Contains(got, "username:") || !strings.Contains(got, "alice") {
		t.Errorf("unexpected key-value output: %q", got)
	}
	if !strings.Contains(got, "status:") {
		t.Errorf("missing status line: %q", got)
	}
	// keys should be sorted alphabetically
	statusIdx := strings.Index(got, "status:")
	usernameIdx := strings.Index(got, "username:")
	if statusIdx > usernameIdx {
		t.Errorf("expected sorted order: status before username, got:\n%s", got)
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	err := output.JSON(&buf, map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("JSON error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, `"key"`) || !strings.Contains(got, `"value"`) {
		t.Errorf("unexpected JSON output: %q", got)
	}
}

func TestFilterFields(t *testing.T) {
	tests := []struct {
		name   string
		data   any
		fields []string
		check  func(t *testing.T, got any)
	}{
		{
			name:   "single object - selected keys retained",
			data:   map[string]any{"id": 1, "email": "a@b.com", "status": "active"},
			fields: []string{"id", "email"},
			check: func(t *testing.T, got any) {
				m, ok := got.(map[string]any)
				if !ok {
					t.Fatalf("expected map[string]any, got %T", got)
				}
				if m["id"] != 1 || m["email"] != "a@b.com" {
					t.Errorf("missing expected keys: %v", m)
				}
				if _, exists := m["status"]; exists {
					t.Errorf("unexpected key 'status' in result")
				}
			},
		},
		{
			name:   "array filtering",
			data:   []any{map[string]any{"id": 1, "email": "a@b.com", "status": "active"}},
			fields: []string{"id"},
			check: func(t *testing.T, got any) {
				arr, ok := got.([]any)
				if !ok {
					t.Fatalf("expected []any, got %T", got)
				}
				m := arr[0].(map[string]any)
				if m["id"] != 1 {
					t.Errorf("expected id=1, got %v", m["id"])
				}
				if _, exists := m["email"]; exists {
					t.Errorf("unexpected key 'email' in result")
				}
			},
		},
		{
			name:   "passthrough for non-map type",
			data:   "plain string",
			fields: []string{"id"},
			check: func(t *testing.T, got any) {
				if got != "plain string" {
					t.Errorf("expected passthrough, got %v", got)
				}
			},
		},
		{
			name:   "nonexistent fields produce empty object",
			data:   map[string]any{"id": 1, "email": "a@b.com"},
			fields: []string{"nonexistent"},
			check: func(t *testing.T, got any) {
				m, ok := got.(map[string]any)
				if !ok {
					t.Fatalf("expected map[string]any, got %T", got)
				}
				if len(m) != 0 {
					t.Errorf("expected empty map, got %v", m)
				}
			},
		},
		{
			name:   "fields with whitespace are trimmed",
			data:   map[string]any{"id": 1, "email": "a@b.com"},
			fields: []string{" id ", " email "},
			check: func(t *testing.T, got any) {
				m, ok := got.(map[string]any)
				if !ok {
					t.Fatalf("expected map[string]any, got %T", got)
				}
				if len(m) != 2 {
					t.Errorf("expected 2 keys, got %d: %v", len(m), m)
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := output.FilterFields(tc.data, tc.fields)
			tc.check(t, got)
		})
	}
}

func TestJSONSummary(t *testing.T) {
	tests := []struct {
		name     string
		returned int
		total    *int
		hasMore  bool
		wantKeys []string
		noKeys   []string
	}{
		{
			name:     "nil total omitted from JSON",
			returned: 10,
			total:    nil,
			hasMore:  false,
			wantKeys: []string{`"returned_items"`, `"has_more"`},
			noKeys:   []string{`"total_items"`},
		},
		{
			name:     "total included when provided",
			returned: 10,
			total:    intPtr(100),
			hasMore:  true,
			wantKeys: []string{`"returned_items"`, `"total_items"`, `"has_more": true`},
		},
		{
			name:     "has_more false",
			returned: 5,
			total:    nil,
			hasMore:  false,
			wantKeys: []string{`"has_more": false`},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := output.JSONSummary(&buf, tc.returned, tc.total, tc.hasMore); err != nil {
				t.Fatalf("JSONSummary error: %v", err)
			}
			got := buf.String()
			for _, key := range tc.wantKeys {
				if !strings.Contains(got, key) {
					t.Errorf("expected %q in output:\n%s", key, got)
				}
			}
			for _, key := range tc.noKeys {
				if strings.Contains(got, key) {
					t.Errorf("unexpected %q in output:\n%s", key, got)
				}
			}
		})
	}
}

func intPtr(n int) *int { return &n }
