package output_test

import (
	"bytes"
	"errors"
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

func TestDetectJSONMode(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"empty", []string{}, false},
		{"--json flag", []string{"users", "list", "--json"}, true},
		{"--format=json", []string{"users", "list", "--format=json"}, true},
		{"-f=json", []string{"users", "list", "-f=json"}, true},
		{"--format json (space-separated)", []string{"users", "--format", "json"}, true},
		{"-f json (space-separated)", []string{"users", "-f", "json"}, true},
		{"--format table (not json)", []string{"users", "--format", "table"}, false},
		{"--format at end (no value)", []string{"users", "--format"}, false},
		{"unrelated flags", []string{"users", "--verbose", "--limit", "10"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := output.DetectJSONMode(tt.args); got != tt.want {
				t.Errorf("DetectJSONMode(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func intPtr(n int) *int { return &n }

func TestToGeneric(t *testing.T) {
	type Inner struct {
		Value string `json:"value"`
	}
	type Sample struct {
		ID       int     `json:"id"`
		Name     string  `json:"name"`
		Hidden   string  `json:"-"`
		Internal string  `json:"-"`
		Optional *string `json:"opt,omitempty"`
		Sub      *Inner  `json:"sub,omitempty"`
	}

	t.Run("flat struct to map", func(t *testing.T) {
		s := Sample{ID: 42, Name: "alice"}
		got, err := output.ToGeneric(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", got)
		}
		if m["id"] != 42 || m["name"] != "alice" {
			t.Errorf("unexpected map contents: %v", m)
		}
		if _, exists := m["Hidden"]; exists {
			t.Error("json:\"-\" field should be excluded")
		}
		if _, exists := m["opt"]; exists {
			t.Error("nil omitempty pointer should be excluded")
		}
	})

	t.Run("slice of structs to []any", func(t *testing.T) {
		items := []Sample{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
		got, err := output.ToGeneric(items)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr, ok := got.([]any)
		if !ok {
			t.Fatalf("expected []any, got %T", got)
		}
		if len(arr) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(arr))
		}
		m := arr[0].(map[string]any)
		if m["id"] != 1 {
			t.Errorf("expected id=1, got %v", m["id"])
		}
	})

	t.Run("non-nil pointer field included", func(t *testing.T) {
		str := "present"
		s := Sample{ID: 1, Optional: &str, Sub: &Inner{Value: "x"}}
		got, err := output.ToGeneric(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m := got.(map[string]any)
		if m["opt"] != "present" {
			t.Errorf("expected opt=present, got %v", m["opt"])
		}
		sub, ok := m["sub"].(map[string]any)
		if !ok || sub["value"] != "x" {
			t.Errorf("expected nested sub.value=x, got %v", m["sub"])
		}
	})

	t.Run("nil pointer to struct returns nil", func(t *testing.T) {
		var p *Sample
		got, err := output.ToGeneric(p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("omitempty slice excluded when nil", func(t *testing.T) {
		type WithSlice struct {
			Items []string `json:"items,omitempty"`
		}
		got, err := output.ToGeneric(WithSlice{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m := got.(map[string]any)
		if _, exists := m["items"]; exists {
			t.Error("nil omitempty slice should be excluded")
		}
	})
}

func TestSuggestionFor(t *testing.T) {
	tests := []struct {
		name    string
		errMsg  string
		wantSub string
	}{
		{"loading profile error", `loading profile "default": token not configured`, "auth login"},
		{"401 status", "users list: 401 Unauthorized", "auth login"},
		{"unauthorized lowercase", "request failed: unauthorized", "auth login"},
		{"404 status", "users get: 404 Not Found", "Check the resource ID"},
		{"not found lowercase", "resource not found", "Check the resource ID"},
		{"500 status", "users list: 500 Internal Server Error", "Try again later"},
		{"server error text", "server error occurred", "Try again later"},
		{"generic error", "some unknown failure", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := output.SuggestionFor(errors.New(tt.errMsg))
			if tt.wantSub == "" {
				if got != "" {
					t.Errorf("SuggestionFor(%q) = %q, want empty", tt.errMsg, got)
				}
			} else if !strings.Contains(got, tt.wantSub) {
				t.Errorf("SuggestionFor(%q) = %q, want substring %q", tt.errMsg, got, tt.wantSub)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	t.Run("json mode", func(t *testing.T) {
		var buf bytes.Buffer
		output.WriteError(&buf, errors.New("something failed"), "Run 'atadmin auth login'", true)
		got := buf.String()
		if !strings.Contains(got, `"error"`) || !strings.Contains(got, "something failed") {
			t.Errorf("WriteError JSON output missing fields: %q", got)
		}
		if !strings.Contains(got, `"suggestion"`) {
			t.Errorf("WriteError JSON output missing suggestion: %q", got)
		}
	})
	t.Run("text mode", func(t *testing.T) {
		var buf bytes.Buffer
		output.WriteError(&buf, errors.New("something failed"), "", false)
		got := buf.String()
		if !strings.HasPrefix(got, "Error:") {
			t.Errorf("WriteError text output should start with 'Error:': %q", got)
		}
	})
	t.Run("json omits empty suggestion", func(t *testing.T) {
		var buf bytes.Buffer
		output.WriteError(&buf, errors.New("oops"), "", true)
		got := buf.String()
		if strings.Contains(got, `"suggestion"`) {
			t.Errorf("WriteError should omit empty suggestion in JSON: %q", got)
		}
	})
}
