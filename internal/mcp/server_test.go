package mcp

import (
	"context"
	"fmt"
	"strings"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
)

// newTestRoot returns a minimal Cobra root for server tests.
func newTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "atadmin", SilenceErrors: true, SilenceUsage: true}
	root.PersistentFlags().String("profile", "default", "Config profile")

	// "ping" writes a fixed JSON body to stdout.
	ping := &cobra.Command{
		Use:   "ping",
		Short: "Ping the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _ = cmd.OutOrStdout().Write([]byte(`{"ok":true}`))
			return nil
		},
	}
	ping.Flags().Bool("json", false, "JSON output")

	// "fail" always exits with an unauthorized error.
	fail := &cobra.Command{
		Use:   "fail",
		Short: "Always returns an error",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _ = cmd.ErrOrStderr().Write([]byte("unauthorized: token invalid"))
			return fmt.Errorf("unauthorized: token invalid")
		},
	}

	root.AddCommand(ping, fail)
	return root
}

// callRequest builds a CallToolRequest with the given arguments.
func callRequest(params map[string]any) mcpgo.CallToolRequest {
	req := mcpgo.CallToolRequest{}
	req.Params.Arguments = params
	return req
}

// textContent extracts the text from the first content block of a result.
func textContent(result *mcpgo.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}
	if tc, ok := result.Content[0].(mcpgo.TextContent); ok {
		return tc.Text
	}
	return ""
}

// --- buildArgs tests ---

func TestBuildArgs_JSONAutoInjected(t *testing.T) {
	def := ToolDef{CommandPath: []string{"users", "list"}, HasJSONFlag: true}
	args := buildArgs(def, map[string]any{})
	for _, a := range args {
		if a == "--json" {
			return
		}
	}
	t.Errorf("expected --json in args when hasJSONFlag=true; got %v", args)
}

func TestBuildArgs_JSONSuppressedWhenFalse(t *testing.T) {
	def := ToolDef{CommandPath: []string{"users", "list"}, HasJSONFlag: true}
	args := buildArgs(def, map[string]any{"json": false})
	for _, a := range args {
		if a == "--json" {
			t.Errorf("--json must be absent when json=false; got %v", args)
		}
	}
}

func TestBuildArgs_BoolFlagTrue(t *testing.T) {
	def := ToolDef{CommandPath: []string{"users", "list"}}
	args := buildArgs(def, map[string]any{"verbose": true, "json": false})
	for _, a := range args {
		if a == "--verbose" {
			return
		}
	}
	t.Errorf("expected --verbose in args; got %v", args)
}

func TestBuildArgs_BoolFlagFalse(t *testing.T) {
	def := ToolDef{CommandPath: []string{"users", "list"}}
	args := buildArgs(def, map[string]any{"verbose": false, "json": false})
	for _, a := range args {
		if a == "--verbose" {
			t.Errorf("--verbose must be absent when false; got %v", args)
		}
	}
}

func TestBuildArgs_StringFlag(t *testing.T) {
	def := ToolDef{CommandPath: []string{"users", "list"}}
	args := buildArgs(def, map[string]any{"filter": "tracked", "json": false})
	for i, a := range args {
		if a == "--filter" && i+1 < len(args) && args[i+1] == "tracked" {
			return
		}
	}
	t.Errorf("expected --filter tracked in args; got %v", args)
}

func TestBuildArgs_IntFlag(t *testing.T) {
	// JSON numbers arrive as float64.
	def := ToolDef{CommandPath: []string{"users", "list"}}
	args := buildArgs(def, map[string]any{"limit": float64(10), "json": false})
	for i, a := range args {
		if a == "--limit" && i+1 < len(args) && args[i+1] == "10" {
			return
		}
	}
	t.Errorf("expected --limit 10 in args; got %v", args)
}

func TestBuildArgs_PositionalID(t *testing.T) {
	def := ToolDef{CommandPath: []string{"users", "get"}, Positionals: []string{"id"}}
	args := buildArgs(def, map[string]any{"id": "42", "json": false})
	for _, a := range args {
		if a == "--id" {
			t.Fatalf("id should be positional, not --id; got %v", args)
		}
	}
	for _, a := range args {
		if a == "42" {
			return
		}
	}
	t.Errorf("expected positional '42' in args; got %v", args)
}

// TestBuildArgs_JSONNotInjectedOnCommandWithoutJSONFlag verifies that passing
// json:true to a command that has no --json flag does not append --json.
func TestBuildArgs_JSONNotInjectedOnCommandWithoutJSONFlag(t *testing.T) {
	def := ToolDef{CommandPath: []string{"groups", "get"}, HasJSONFlag: false, Positionals: []string{"id"}}
	args := buildArgs(def, map[string]any{"id": "706"})
	for _, a := range args {
		if a == "--json" {
			t.Errorf("--json must not appear when HasJSONFlag=false; got %v", args)
		}
	}
	found := false
	for _, a := range args {
		if a == "706" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected positional '706' in args; got %v", args)
	}
}

// --- makeHandler tests ---

func TestMakeHandler_Success(t *testing.T) {
	def := ToolDef{
		Name:        "ping",
		Description: "Ping",
		CommandPath: []string{"ping"},
		HasJSONFlag: true,
	}
	handler := makeHandler(def, newTestRoot)
	result, err := handler(context.Background(), callRequest(map[string]any{"json": false}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected IsError=false; content=%q", textContent(result))
	}
	if got := textContent(result); got != `{"ok":true}` {
		t.Errorf("content = %q, want %q", got, `{"ok":true}`)
	}
}

func TestMakeHandler_CommandError(t *testing.T) {
	def := ToolDef{
		Name:        "fail",
		Description: "Fail",
		CommandPath: []string{"fail"},
	}
	handler := makeHandler(def, newTestRoot)
	result, err := handler(context.Background(), callRequest(nil))
	if err != nil {
		t.Fatalf("handler must not return a Go error; got %v", err)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true for a failing command")
	}
}

func TestMakeHandler_AuthErrorHint(t *testing.T) {
	def := ToolDef{
		Name:        "fail",
		Description: "Fail",
		CommandPath: []string{"fail"},
	}
	handler := makeHandler(def, newTestRoot)
	result, _ := handler(context.Background(), callRequest(nil))
	if !result.IsError {
		t.Fatal("expected IsError=true")
	}
	content := textContent(result)
	const hint = "atadmin auth login"
	if !strings.Contains(content, hint) {
		t.Errorf("expected auth hint %q in error message; got %q", hint, content)
	}
}

// --- isAuthError tests ---

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"unauthorized: invalid token", true},
		{"401 Unauthorized", true},
		{"authentication required", true},
		{"forbidden", true},
		{"resource not found", false},
		{"timeout", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isAuthError(tt.msg); got != tt.want {
			t.Errorf("isAuthError(%q) = %v, want %v", tt.msg, got, tt.want)
		}
	}
}

