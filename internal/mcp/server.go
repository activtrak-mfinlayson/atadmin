package mcp

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

// Server wraps an mcp-go MCPServer configured from the Cobra command tree.
type Server struct {
	mcpServer *mcpserver.MCPServer
	logFile   *os.File
	newRoot   func() *cobra.Command
}

// NewServer builds an MCP server from the given Cobra root command.
// It opens ~/.config/atadmin/mcp.log for all diagnostic output, walks the
// command tree to produce tool definitions, and registers a handler for each.
func NewServer(newRoot func() *cobra.Command, version string, allowMutations bool) (*Server, error) {
	lf, err := setupLogger()
	if err != nil {
		return nil, fmt.Errorf("opening MCP log file: %w", err)
	}

	root := newRoot()
	defs := Walk(root, allowMutations)

	s := mcpserver.NewMCPServer(
		"atadmin",
		version,
		mcpserver.WithToolCapabilities(true),
	)

	srv := &Server{
		mcpServer: s,
		logFile:   lf,
		newRoot:   newRoot,
	}

	for _, def := range defs {
		tool := buildTool(def)
		handler := makeHandler(def, newRoot)
		s.AddTool(tool, handler)
	}

	log.Printf("[atadmin-mcp] server started: %d tools registered (allowMutations=%v)", len(defs), allowMutations)
	return srv, nil
}

// Start begins serving MCP over stdio.  It blocks until stdin is closed.
func (s *Server) Start() error {
	defer func() {
		if s.logFile != nil {
			_ = s.logFile.Close()
		}
	}()

	logger := log.New(s.logFile, "", log.LstdFlags)
	if err := mcpserver.ServeStdio(s.mcpServer, mcpserver.WithErrorLogger(logger)); err != nil {
		return fmt.Errorf("MCP stdio server error: %w", err)
	}
	return nil
}

// setupLogger opens (or creates) ~/.config/atadmin/mcp.log and redirects the
// default logger to it.  All subsequent log.Print* calls go to the file.
// Nothing is written to stdout or stderr during MCP operation.
func setupLogger() (*os.File, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".config", "atadmin")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("creating config dir %s: %w", dir, err)
	}
	path := filepath.Join(dir, "mcp.log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("opening log file %s: %w", path, err)
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags)
	return f, nil
}

// buildTool converts a ToolDef into an mcp-go Tool with a raw JSON Schema.
func buildTool(def ToolDef) mcpgo.Tool {
	schema := BuildInputSchema(def.Params)
	return mcpgo.NewToolWithRawSchema(def.Name, def.Description, schema)
}

// makeHandler returns a ToolHandlerFunc that invokes the corresponding Cobra
// command and captures its output.
func makeHandler(def ToolDef, newRoot func() *cobra.Command) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		args := buildArgs(def, req.GetArguments())

		root := newRoot()
		var outBuf, errBuf bytes.Buffer
		root.SetOut(&outBuf)
		root.SetErr(&errBuf)
		root.SetArgs(args)

		if err := root.ExecuteContext(ctx); err != nil {
			msg := errBuf.String()
			if msg == "" {
				msg = err.Error()
			}
			msg = strings.TrimSpace(msg)
			if isAuthError(msg) {
				msg = "API authentication failed. Run 'atadmin auth login' to configure credentials, then restart the MCP server.\n\nDetails: " + msg
			}
			log.Printf("[atadmin-mcp] tool %q error: %s", def.Name, msg)
			return mcpgo.NewToolResultError(msg), nil
		}

		result := strings.TrimRight(outBuf.String(), "\n")
		log.Printf("[atadmin-mcp] tool %q ok (%d bytes)", def.Name, len(result))
		return mcpgo.NewToolResultText(result), nil
	}
}

// buildArgs converts a tool call's parameter map into a []string CLI argument
// slice suitable for passing to cobra.Command.SetArgs.
func buildArgs(def ToolDef, params map[string]any) []string {
	args := make([]string, len(def.CommandPath))
	copy(args, def.CommandPath)

	// Auto-inject --json when the command supports it and the caller hasn't
	// explicitly set json:false.
	if def.HasJSONFlag {
		if v, ok := params["json"]; !ok || v == true {
			args = append(args, "--json")
		}
	}

	// Positional arguments — appended in declaration order, bare (not as --flag).
	positionalSet := make(map[string]bool, len(def.Positionals))
	for _, pos := range def.Positionals {
		positionalSet[pos] = true
		if v, ok := params[pos]; ok {
			args = append(args, fmt.Sprintf("%v", v))
		}
	}

	for name, val := range params {
		// "json" is handled exclusively by the HasJSONFlag branch above.
		// Positionals are already appended in order above.
		if name == "json" || positionalSet[name] {
			continue
		}
		switch v := val.(type) {
		case bool:
			if v {
				args = append(args, "--"+name)
			}
		case float64:
			// JSON numbers come through as float64.
			args = append(args, "--"+name, fmt.Sprintf("%g", v))
		default:
			args = append(args, "--"+name, fmt.Sprintf("%v", v))
		}
	}
	return args
}

// isAuthError returns true when the error message indicates an authentication
// or authorization failure.
func isAuthError(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "unauthorized") ||
		strings.Contains(lower, "401") ||
		strings.Contains(lower, "forbidden") ||
		strings.Contains(lower, "authentication")
}
