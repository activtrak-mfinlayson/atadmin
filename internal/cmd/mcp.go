package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	internalmcp "github.com/activtrak-mfinlayson/atadmin/internal/mcp"
)

// newMCPCmd returns the "mcp" parent command.
func newMCPCmd() *cobra.Command {
	mcp := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol server commands",
		Long:  `Commands for running atadmin as an MCP server for AI agents.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	mcp.AddCommand(newMCPServeCmd())
	return mcp
}

// newMCPServeCmd returns the "mcp serve" command.
func newMCPServeCmd() *cobra.Command {
	var (
		allowMutations bool
	)

	cmd := &cobra.Command{
		Use:   "serve --stdio",
		Short: "Start an MCP stdio server exposing all atadmin commands as tools",
		Long: `Start a Model Context Protocol server on stdin/stdout.

Any MCP-compatible client (Claude Desktop, Cursor, Craft Agent) can connect
and discover all atadmin commands as named tools with typed parameters.

By default only read-only commands are exposed. Pass --allow-mutations to
also include commands that modify data (update, delete, bulk actions).

Diagnostic output is written to ~/.config/atadmin/mcp.log.

Claude Desktop configuration (~/.config/Claude/claude_desktop_config.json):
  {
    "mcpServers": {
      "atadmin": {
        "command": "atadmin",
        "args": ["mcp", "serve", "--stdio"]
      }
    }
  }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// PersistentPreRunE on root normally builds the API client, but
			// the MCP server does not need it at startup — each tool call
			// creates its own root instance.  We skip PersistentPreRunE by
			// constructing a fresh root via NewRootCmd inside the server.
			srv, err := internalmcp.NewServer(NewRootCmd, Version, allowMutations)
			if err != nil {
				return fmt.Errorf("initializing MCP server: %w", err)
			}
			return srv.Start()
		},
	}

	cmd.Flags().BoolVar(&allowMutations, "allow-mutations", false,
		"Also expose mutation commands (update, delete, bulk actions)")
	// --stdio is the canonical flag documented in MCP tooling; we accept it
	// but it has no effect since stdio is the only transport.
	cmd.Flags().Bool("stdio", true, "Use stdio transport (always true; provided for compatibility)")
	_ = cmd.Flags().MarkHidden("stdio")

	return cmd
}
