package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/activtrak-mfinlayson/atadmin/internal/cmd"
)

func TestRootVersion(t *testing.T) {
	root := cmd.NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"--version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "atadmin version "+cmd.Version) {
		t.Errorf("version output = %q, want it to contain %q", got, "atadmin version "+cmd.Version)
	}
}

func TestRootHelp(t *testing.T) {
	root := cmd.NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "atadmin") {
		t.Errorf("help output = %q, does not contain %q", got, "atadmin")
	}
	if !strings.Contains(got, "auth") {
		t.Errorf("help output = %q, does not list 'auth' subcommand", got)
	}
}

func TestRootUnknownCommand(t *testing.T) {
	root := cmd.NewRootCmd()
	root.SetArgs([]string{"unknowncmd"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for unknown command, got nil")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("error = %q, want to contain 'unknown command'", err.Error())
	}
}

func TestRootNoArgs(t *testing.T) {
	root := cmd.NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{})

	// No args should show help (not error).
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() with no args returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "atadmin") {
		t.Error("no-args output does not contain 'atadmin'")
	}
}

// ---------------------------------------------------------------------------
// T014: CLI-level dry-run output contract
// ---------------------------------------------------------------------------

func TestDryRunCLIOutputContract(t *testing.T) {
	var buf bytes.Buffer
	root := cmd.NewTestRootWithDryRun(&buf)
	root.SetArgs([]string{"groups", "rename", "42", "--name", "Engineering"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	var got struct {
		Action  string          `json:"action"`
		Target  string          `json:"target"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.NewDecoder(&buf).Decode(&got); err != nil {
		t.Fatalf("dry-run output is not valid JSON: %v\nraw: %q", err, buf.String())
	}

	if got.Action != "update" {
		t.Errorf("action = %q, want %q", got.Action, "update")
	}
	if got.Target != "/admin/v1/groups/42" {
		t.Errorf("target = %q, want %q", got.Target, "/admin/v1/groups/42")
	}
	if len(got.Payload) == 0 {
		t.Error("payload field is missing or empty")
	}
	// Payload should contain the name field
	var payload map[string]any
	if err := json.Unmarshal(got.Payload, &payload); err != nil {
		t.Fatalf("payload is not valid JSON object: %v\nraw: %s", err, string(got.Payload))
	}
	if payload["name"] != "Engineering" {
		t.Errorf("payload[name] = %v, want %q", payload["name"], "Engineering")
	}
}
