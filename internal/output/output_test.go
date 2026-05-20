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
