package mcp

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

// buildTestRoot constructs a minimal Cobra command tree for testing.
//
//	root
//	 ├── items
//	 │    ├── list   (flags: filter string, limit int, json bool)
//	 │    └── delete (mutation verb)
//	 ├── audit-log
//	 │    └── list
//	 ├── hidden-cmd  (Hidden: true — excluded)
//	 ├── old-cmd     (Deprecated — excluded)
//	 └── mcp
//	      └── serve  (always excluded)
func buildTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "atadmin", SilenceErrors: true, SilenceUsage: true}
	root.PersistentFlags().String("profile", "default", "Config profile")

	// Use "items" (not "users") so the test tree is independent of excludedTools.
	users := &cobra.Command{Use: "items", Short: "Manage items"}
	usersList := &cobra.Command{Use: "list", Short: "List items"}
	usersList.Flags().String("filter", "", "Filter type")
	usersList.Flags().Int("limit", 0, "Max results")
	usersList.Flags().Bool("json", false, "Output JSON")
	usersDelete := &cobra.Command{Use: "delete [id]", Short: "Delete an item"}
	users.AddCommand(usersList, usersDelete)

	auditLog := &cobra.Command{Use: "audit-log", Short: "Audit log commands"}
	auditLogList := &cobra.Command{Use: "list", Short: "List audit log entries"}
	auditLog.AddCommand(auditLogList)

	hidden := &cobra.Command{Use: "hidden-cmd", Short: "Hidden", Hidden: true}
	deprecated := &cobra.Command{Use: "old-cmd", Short: "Old", Deprecated: "use new-cmd"}

	mcpCmd := &cobra.Command{Use: "mcp", Short: "MCP server"}
	mcpServe := &cobra.Command{Use: "serve", Short: "Start MCP server"}
	mcpCmd.AddCommand(mcpServe)

	root.AddCommand(users, auditLog, hidden, deprecated, mcpCmd)
	return root
}

func TestToolName(t *testing.T) {
	tests := []struct {
		path []string
		want string
	}{
		{[]string{"users", "list"}, "users_list"},
		{[]string{"audit-log", "list"}, "audit_log_list"},
		{[]string{"api-keys", "list"}, "api_keys_list"},
		{[]string{"users", "get"}, "users_get"},
	}
	for _, tt := range tests {
		got := toolName(tt.path)
		if got != tt.want {
			t.Errorf("toolName(%v) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestIsMutation(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"delete", true},
		{"update", true},
		{"remove", true},
		{"bulk", true},
		{"add", true},
		{"create", true},
		{"list", false},
		{"get", false},
		{"serve", false},
	}
	for _, tt := range tests {
		cmd := &cobra.Command{Use: tt.name}
		if got := isMutation(cmd); got != tt.want {
			t.Errorf("isMutation(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestExtractParams_ExcludesReservedFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("token", "", "API token")
	cmd.Flags().String("base-url", "", "Base URL")
	cmd.Flags().Bool("help", false, "Show help")
	cmd.Flags().String("filter", "", "A filter")
	cmd.Flags().Bool("json", false, "JSON output")

	params := extractParams(cmd, nil)
	names := map[string]bool{}
	for _, p := range params {
		names[p.Name] = true
	}
	for _, excluded := range []string{"token", "base-url", "help"} {
		if names[excluded] {
			t.Errorf("extractParams must not include flag %q", excluded)
		}
	}
	for _, expected := range []string{"filter", "json"} {
		if !names[expected] {
			t.Errorf("extractParams must include flag %q", expected)
		}
	}
}

func TestExtractParams_TypeMapping(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "A string flag")
	cmd.Flags().Int("count", 0, "An int flag")
	cmd.Flags().Bool("verbose", false, "A bool flag")

	byName := map[string]ToolParam{}
	for _, p := range extractParams(cmd, nil) {
		byName[p.Name] = p
	}
	if got := byName["name"].JSONType; got != "string" {
		t.Errorf("'name' JSONType = %q, want 'string'", got)
	}
	if got := byName["count"].JSONType; got != "integer" {
		t.Errorf("'count' JSONType = %q, want 'integer'", got)
	}
	if got := byName["verbose"].JSONType; got != "boolean" {
		t.Errorf("'verbose' JSONType = %q, want 'boolean'", got)
	}
}

func TestWalk_ReadOnly(t *testing.T) {
	root := buildTestRoot()
	defs := Walk(root, false)

	names := map[string]bool{}
	for _, d := range defs {
		names[d.Name] = true
	}

	for _, want := range []string{"items_list", "audit_log_list"} {
		if !names[want] {
			t.Errorf("Walk(allowMutations=false): expected tool %q in list", want)
		}
	}
	for _, banned := range []string{"items_delete", "mcp_serve", "hidden_cmd", "old_cmd"} {
		if names[banned] {
			t.Errorf("Walk(allowMutations=false): %q should be excluded", banned)
		}
	}
}

func TestWalk_AllowMutations(t *testing.T) {
	root := buildTestRoot()
	defs := Walk(root, true)

	names := map[string]bool{}
	for _, d := range defs {
		names[d.Name] = true
	}
	if !names["items_delete"] {
		t.Error("Walk(allowMutations=true): items_delete should be included")
	}
	// MCP subtree still excluded regardless.
	if names["mcp_serve"] {
		t.Error("Walk: mcp_serve must always be excluded")
	}
}

func TestWalk_HasJSONFlag(t *testing.T) {
	root := buildTestRoot()
	for _, d := range Walk(root, false) {
		if d.Name == "items_list" {
			if !d.HasJSONFlag {
				t.Error("items_list should have HasJSONFlag=true")
			}
			return
		}
	}
	t.Error("items_list not found in Walk output")
}

func TestBuildInputSchema(t *testing.T) {
	params := []ToolParam{
		{Name: "filter", JSONType: "string", Description: "Filter type"},
		{Name: "limit", JSONType: "integer", Description: "Max results"},
		{Name: "json", JSONType: "boolean", Description: "JSON output"},
	}
	raw := BuildInputSchema(params)
	if len(raw) == 0 {
		t.Fatal("BuildInputSchema returned empty schema")
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("BuildInputSchema produced invalid JSON: %v", err)
	}
	if obj["type"] != "object" {
		t.Errorf("schema[type] = %v, want 'object'", obj["type"])
	}
	props, ok := obj["properties"].(map[string]any)
	if !ok {
		t.Fatal("schema missing 'properties' object")
	}
	for _, name := range []string{"filter", "limit", "json"} {
		if props[name] == nil {
			t.Errorf("schema missing property %q", name)
		}
	}
}
