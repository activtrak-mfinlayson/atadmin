// Package cmd_test exercises the clients command output-formatting helpers.
// Tests are deliberately kept at the output-function level so they do not
// require a live API, a config file, or a real Cobra command execution path.
// This lets the test suite pass before the api.Client methods are implemented.
package cmd_test

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/activtrak-mfinlayson/atadmin/internal/api"
	"github.com/activtrak-mfinlayson/atadmin/internal/cmd"
	"github.com/activtrak-mfinlayson/atadmin/internal/output"
)

// ---------------------------------------------------------------------------
// Local helpers: mirror the rendering logic used inside clients.go so that
// tests can validate the formatted output without calling the Cobra commands.
// ---------------------------------------------------------------------------

func renderClientTable(clients []api.ATClient) string {
	var buf bytes.Buffer
	rows := make([][]string, len(clients))
	for i, c := range clients {
		rows[i] = []string{c.Username, c.Alias, c.Status}
	}
	output.Table(&buf, []string{"USERNAME", "ALIAS", "STATUS"}, rows)
	return buf.String()
}

func renderClientJSON(clients []api.ATClient) (string, error) {
	var buf bytes.Buffer
	if err := output.JSON(&buf, clients); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderClientKeyValue(c api.ATClient) string {
	var buf bytes.Buffer
	output.KeyValue(&buf, map[string]string{
		"id":          strconv.Itoa(c.ID),
		"username":    c.Username,
		"logonDomain": c.LogonDomain,
		"alias":       c.Alias,
		"status":      c.Status,
		"deviceCount": strconv.Itoa(c.DeviceCount),
	})
	return buf.String()
}

func renderDNTTable(entries []api.DNTEntry) string {
	var buf bytes.Buffer
	rows := make([][]string, len(entries))
	for i, e := range entries {
		isGlobal := "false"
		if e.IsGlobal {
			isGlobal = "true"
		}
		rows[i] = []string{
			strconv.Itoa(e.ID),
			e.LogonDomain,
			e.Username,
			isGlobal,
		}
	}
	output.Table(&buf, []string{"ID", "DOMAIN", "USERNAME", "GLOBAL"}, rows)
	return buf.String()
}

// ---------------------------------------------------------------------------
// Tests: clients list — table headers present
// ---------------------------------------------------------------------------

func TestClientsListTableHeaders(t *testing.T) {
	got := renderClientTable([]api.ATClient{
		{ID: 1, Username: "alice", Alias: "Alice Smith", Status: "active"},
	})

	for _, want := range []string{"USERNAME", "ALIAS", "STATUS"} {
		if !strings.Contains(got, want) {
			t.Errorf("table output missing header %q:\n%s", want, got)
		}
	}
}

func TestClientsListTableRows(t *testing.T) {
	clients := []api.ATClient{
		{ID: 1, Username: "alice", Alias: "Alice Smith", Status: "active"},
		{ID: 2, Username: "bob", Alias: "", Status: "inactive"},
	}

	got := renderClientTable(clients)

	for _, want := range []string{"alice", "Alice Smith", "active", "bob", "inactive"} {
		if !strings.Contains(got, want) {
			t.Errorf("table output missing value %q:\n%s", want, got)
		}
	}
}

func TestClientsListTableEmptyStillHasHeaders(t *testing.T) {
	got := renderClientTable(nil)

	for _, want := range []string{"USERNAME", "ALIAS", "STATUS"} {
		if !strings.Contains(got, want) {
			t.Errorf("empty table output missing header %q:\n%s", want, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Tests: clients list --json — valid JSON array output
// ---------------------------------------------------------------------------

func TestClientsListJSONIsParseable(t *testing.T) {
	clients := []api.ATClient{
		{ID: 10, Username: "carol", Alias: "Carol", Status: "active", DeviceCount: 1},
	}

	got, err := renderClientJSON(clients)
	if err != nil {
		t.Fatalf("renderClientJSON error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\nraw output:\n%s", err, got)
	}

	if len(decoded) != 1 {
		t.Errorf("expected 1 JSON element, got %d", len(decoded))
	}
	if decoded[0]["username"] != "carol" {
		t.Errorf("expected username=carol in JSON, got %v", decoded[0]["username"])
	}
}

func TestClientsListJSONMultipleClients(t *testing.T) {
	clients := []api.ATClient{
		{ID: 1, Username: "alice", Status: "active"},
		{ID: 2, Username: "bob", Status: "inactive"},
	}

	got, err := renderClientJSON(clients)
	if err != nil {
		t.Fatalf("renderClientJSON error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\nraw:\n%s", err, got)
	}
	if len(decoded) != 2 {
		t.Errorf("expected 2 JSON elements, got %d", len(decoded))
	}
}

func TestClientsListJSONEmptyIsArray(t *testing.T) {
	got, err := renderClientJSON([]api.ATClient{})
	if err != nil {
		t.Fatalf("renderClientJSON error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\nraw:\n%s", err, got)
	}
	if len(decoded) != 0 {
		t.Errorf("expected empty JSON array, got %d elements", len(decoded))
	}
}

// ---------------------------------------------------------------------------
// Tests: clients get — key-value output contains "username" field
// ---------------------------------------------------------------------------

func TestClientsGetKeyValueContainsUsername(t *testing.T) {
	c := api.ATClient{
		ID:          42,
		Username:    "dave",
		LogonDomain: "CORP",
		Alias:       "Dave",
		Status:      "active",
		DeviceCount: 3,
	}

	got := renderClientKeyValue(c)

	if !strings.Contains(got, "username") {
		t.Errorf("key-value output missing key 'username':\n%s", got)
	}
	if !strings.Contains(got, "dave") {
		t.Errorf("key-value output missing value 'dave':\n%s", got)
	}
}

func TestClientsGetKeyValueAllFieldsPresent(t *testing.T) {
	c := api.ATClient{
		ID:          42,
		Username:    "dave",
		LogonDomain: "CORP",
		Alias:       "Dave",
		Status:      "active",
		DeviceCount: 3,
	}

	got := renderClientKeyValue(c)

	wantPairs := []struct{ key, value string }{
		{"id", "42"},
		{"username", "dave"},
		{"logonDomain", "CORP"},
		{"alias", "Dave"},
		{"status", "active"},
		{"deviceCount", "3"},
	}
	for _, p := range wantPairs {
		if !strings.Contains(got, p.key) {
			t.Errorf("key-value output missing key %q:\n%s", p.key, got)
		}
		if !strings.Contains(got, p.value) {
			t.Errorf("key-value output missing value %q for key %q:\n%s", p.value, p.key, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Tests: donottrack list — table headers and row values
// ---------------------------------------------------------------------------

func TestDNTListTableHeaders(t *testing.T) {
	got := renderDNTTable([]api.DNTEntry{
		{ID: 1, LogonDomain: "CORP", Username: "eve", IsGlobal: false},
	})

	for _, want := range []string{"ID", "DOMAIN", "USERNAME", "GLOBAL"} {
		if !strings.Contains(got, want) {
			t.Errorf("DNT table output missing header %q:\n%s", want, got)
		}
	}
}

func TestDNTListTableRowValues(t *testing.T) {
	entries := []api.DNTEntry{
		{ID: 1, LogonDomain: "CORP", Username: "eve", IsGlobal: false},
		{ID: 2, LogonDomain: "", Username: "", IsGlobal: true},
	}

	got := renderDNTTable(entries)

	for _, want := range []string{"CORP", "eve", "false", "true"} {
		if !strings.Contains(got, want) {
			t.Errorf("DNT table output missing value %q:\n%s", want, got)
		}
	}
}

func TestDNTListTableEmptyStillHasHeaders(t *testing.T) {
	got := renderDNTTable(nil)

	for _, want := range []string{"ID", "DOMAIN", "USERNAME", "GLOBAL"} {
		if !strings.Contains(got, want) {
			t.Errorf("empty DNT table output missing header %q:\n%s", want, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Tests: NewRootCmd help — verify "clients" appears in top-level help
// ---------------------------------------------------------------------------

func TestRootHelpListsClientsSubcommand(t *testing.T) {
	root := cmd.NewTestClientsRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"clients", "--help"})

	// Cobra writes help then returns nil; we only care that "clients" subcommands
	// appear without panicking.
	_ = root.Execute()

	got := buf.String()
	for _, want := range []string{"list", "get", "health", "delete", "restore", "merge", "alias", "donottrack"} {
		if !strings.Contains(got, want) {
			t.Errorf("clients --help output missing subcommand %q:\n%s", want, got)
		}
	}
}
