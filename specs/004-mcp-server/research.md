# Research: Embedded MCP Server for atadmin

**Feature**: 004-mcp-server  
**Date**: 2026-05-19

---

## R-001: Go MCP Library Selection

**Decision**: Use `github.com/mark3labs/mcp-go`

**Rationale**: Most mature and widely-deployed Go MCP library. Supports stdio transport, MCP protocol version 2024-11-05 through 2025-11-25, and programmatic tool registration at runtime. The newer official `github.com/modelcontextprotocol/go-sdk` (maintained by the MCP project) is an alternative but is newer and less battle-tested in production as of this writing. `mcp-go` has the largest community adoption, extensive documentation, and confirmed support for the dynamic tool registration pattern this feature requires.

**Alternatives considered**:
- `github.com/modelcontextprotocol/go-sdk` — official SDK, also supports stdio and dynamic registration; viable alternative if mcp-go has gaps
- Custom JSON-RPC 2.0 implementation — full control but significant boilerplate; rejected as unnecessary given quality libraries exist

---

## R-002: Cobra Command Tree Enumeration

**Decision**: Walk `cmd.Commands()` recursively; identify leaf commands by `len(cmd.Commands()) == 0`

**Rationale**: Cobra exposes the full command tree via `cmd.Commands()`. Walking depth-first and collecting nodes where `len(cmd.Commands()) == 0` gives all executable leaf commands. This is the standard Cobra pattern and is confirmed to work with Cobra v1.

**Edge cases to handle**:
- **Hidden commands**: Check `cmd.Hidden` — exclude hidden commands from tool list
- **Deprecated commands**: Check `cmd.Deprecated != ""` — exclude deprecated commands
- **The `mcp` command itself**: Exclude explicitly to prevent circular self-reference
- **Help commands**: Cobra auto-adds a `help` subcommand — exclude by name

**Alternatives considered**:
- Manual tool registration — defeats the purpose of automatic mapping; rejected
- Reflection — not needed; Cobra provides explicit tree traversal methods

---

## R-003: Flag Extraction and Type Mapping

**Decision**: Use `pflag.FlagSet.VisitAll` on both `cmd.Flags()` and `cmd.InheritedFlags()` to collect all applicable flags; derive type from `flag.Value.Type()`

**Rationale**: Each `*pflag.Flag` exposes `Name`, `Usage`, `DefValue`, and `Value.Type()`. The `Value.Type()` string maps directly to JSON Schema types: `"bool"` → `boolean`, `"int"` → `integer`, everything else → `string`. `cmd.InheritedFlags()` surfaces persistent flags (e.g., `--profile`, `--format`) that the command inherits from ancestors, avoiding duplication while ensuring completeness.

**Flags to always exclude from tool parameters**: `help` (auto-added by Cobra, not meaningful for agents)

---

## R-004: Tool Name Derivation

**Decision**: Join the command path segments (excluding the root `atadmin`) with underscores, lowercase: `atadmin users list` → `users_list`

**Rationale**: MCP tool names must be unique string identifiers. Underscores are conventional in the MCP ecosystem and produce valid identifiers. The root command name is omitted (it's constant and adds no disambiguation value). This produces names like `users_list`, `users_get`, `groups_list`, `devices_list`, `alarms_list`, `audit_log_list`.

**Alternatives considered**:
- Slash-separated (`users/list`) — not conventional for MCP tool names
- Dash-separated (`users-list`) — valid but underscores are more common

---

## R-005: Command Execution and Output Capture

**Decision**: Create a fresh `NewRootCmd()` instance per tool call; set `cmd.SetOut(buf)` and `cmd.SetErr(errbuf)` before executing; pass tool parameters as reconstructed `args []string`

**Rationale**: The existing `NewRootCmd()` factory is safe to call per-request (confirmed in existing tests). Injecting a `bytes.Buffer` via `SetOut`/`SetErr` captures output without any OS pipe complexity. Constructing `args` from MCP parameters (flag name → `--flag-name value`) is straightforward. For boolean flags, emit `--flag` when `true`, omit when `false`.

**JSON output**: Before executing, check if the command has a `--json` flag registered (`cmd.Flags().Lookup("json") != nil`). If yes, prepend `--json` to args (or append `--format=json`) to get machine-readable output. The JSON content is returned as the MCP text content block.

**Alternatives considered**:
- `exec.Command` subprocess — adds process spawn overhead; loses in-process context; rejected
- OS pipe redirection — more complex; requires goroutines; no advantage over buffer injection

---

## R-006: MCP Protocol Version

**Decision**: Target `2024-11-05`; advertise capability for `2025-03-26` if the client requests it

**Rationale**: `2024-11-05` is universally supported by Claude Desktop, Cursor, and all major MCP clients as of mid-2025. `mark3labs/mcp-go` handles protocol version negotiation automatically during the `initialize` handshake, so the server will negotiate the highest mutually supported version without any application code.

---

## R-007: Error Handling Strategy

**Decision**: Return `isError: true` in the MCP tool call response (not a JSON-RPC error) for command failures; reserve JSON-RPC errors for protocol-level failures

**Rationale**: MCP distinguishes between tool execution errors (expected failures — e.g., API 404, auth error) and protocol errors (malformed request). Tool execution errors should use the `isError: true` content block so the agent receives descriptive text and can react intelligently. Protocol errors use the JSON-RPC error response. This matches the MCP spec and the `mcp-go` library's conventions.
