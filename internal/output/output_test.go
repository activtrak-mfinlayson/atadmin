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
