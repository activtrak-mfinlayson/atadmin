# Data Model: Embedded MCP Server for atadmin

**Feature**: 004-mcp-server  
**Date**: 2026-05-19

This feature has no persistent storage. All entities are in-memory constructs that exist for the lifetime of the `atadmin mcp serve` process.

---

## Entities

### MCPTool

Represents a single tool exposed over the MCP protocol. Maps 1:1 to a leaf Cobra subcommand.

| Field        | Type            | Description                                                              |
|--------------|-----------------|--------------------------------------------------------------------------|
| Name         | string          | Unique identifier. Derived from command path: `users_list`, `groups_get` |
| Description  | string          | Human-readable summary. Sourced from `cobra.Command.Short`               |
| InputSchema  | JSONSchema       | Describes accepted parameters. Generated from command flags.             |
| CommandPath  | []string        | The Cobra command path segments: `["users", "list"]`                     |
| HasJSONFlag  | bool            | Whether the underlying command supports `--json` output.                 |

**Derivation rules**:
- `Name` = strings.Join(commandPath, "_"), lowercased, excluding root command name
- `Description` = `cmd.Short` (falls back to first line of `cmd.Long` if `Short` is empty)
- `InputSchema` = JSON Schema `object` with one property per flag (see `ToolParameter`)
- `HasJSONFlag` = `cmd.Flags().Lookup("json") != nil`

---

### ToolParameter

One parameter within an `MCPTool`'s input schema. Maps 1:1 to a Cobra flag.

| Field       | Type    | Description                                              |
|-------------|---------|----------------------------------------------------------|
| Name        | string  | Flag name without leading dashes (e.g., `limit`)         |
| Type        | string  | JSON Schema type: `"boolean"`, `"integer"`, or `"string"` |
| Description | string  | Sourced from `pflag.Flag.Usage`                          |
| Default     | string  | Sourced from `pflag.Flag.DefValue`; empty string if none |
| Required    | bool    | Always `false` — all MCP parameters are optional         |

**Type mapping**:
- `pflag.Flag.Value.Type() == "bool"` → `"boolean"`
- `pflag.Flag.Value.Type() == "int"` → `"integer"`
- Any other type → `"string"`

**Excluded flags**: `help` (auto-added by Cobra; not meaningful for agent callers)

---

### ToolCallRequest

The parsed input from an agent's `tools/call` MCP request.

| Field      | Type              | Description                                      |
|------------|-------------------|--------------------------------------------------|
| ToolName   | string            | The MCP tool name (e.g., `users_list`)           |
| Parameters | map[string]any    | The agent-supplied parameter values               |

---

### ToolCallResult

The response produced after executing the underlying Cobra command.

| Field    | Type    | Description                                                                 |
|----------|---------|-----------------------------------------------------------------------------|
| Content  | string  | The captured stdout of the command (JSON or plain text)                     |
| IsError  | bool    | `true` if the command returned a non-zero exit code or the API returned an error |
| MIMEType | string  | `"application/json"` when `HasJSONFlag` was true and command succeeded; `"text/plain"` otherwise |

---

### MCPServer

The long-running process state. Instantiated once at `atadmin mcp serve --stdio` startup.

| Field      | Type         | Description                                          |
|------------|--------------|------------------------------------------------------|
| Tools      | []MCPTool    | The complete tool list, built once from the Cobra tree at startup |
| Transport  | stdio        | Reads JSON-RPC from stdin; writes to stdout          |
| Version    | string       | MCP protocol version negotiated during `initialize`  |

**Lifecycle**:
1. On startup: walk Cobra tree → build `Tools` list
2. On `initialize`: respond with server info and tool capability
3. On `tools/list`: return `Tools`
4. On `tools/call`: find tool by name → reconstruct args → execute → return result
5. On stdin close: exit cleanly

---

## Relationships

```
MCPServer
  └── has many → MCPTool
                    └── has many → ToolParameter

tools/call request → ToolCallRequest → MCPServer executes → ToolCallResult
```

---

## Invariants

- Tool names are unique within a server instance (enforced by Cobra's own uniqueness constraint on command names)
- The `mcp` command subtree is always excluded from the tool list
- Hidden and deprecated Cobra commands are always excluded
- The `help` flag is never exposed as a tool parameter
