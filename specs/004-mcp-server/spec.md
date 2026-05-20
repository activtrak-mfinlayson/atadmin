# Feature Specification: Embedded MCP Server for atadmin

**Feature Branch**: `004-mcp-server`  
**Created**: 2026-05-19  
**Status**: Draft  
**Input**: User description: "Turn atadmin into its own MCP Server. Add a command `atadmin mcp serve --stdio`. Since atadmin uses Cobra, write a wrapper that automatically maps every Cobra command and its flags into an MCP Tool definition. Any MCP-compatible agent (like Craft Agent, Claude Desktop, Cursor) can instantly connect to atadmin and natively understand every command, required type, and description without running --help once."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Agent Connects and Discovers All Tools (Priority: P1)

An AI agent or developer starts `atadmin mcp serve --stdio` and connects via the MCP protocol. The agent receives a complete list of all atadmin commands (users, groups, devices, accounts, etc.) as named MCP tools, each with typed parameters and descriptions derived from the existing command definitions — without the agent reading any help text or documentation.

**Why this priority**: This is the entire value proposition. Without tool discovery working correctly, no MCP-compatible agent can use atadmin. It is the gateway requirement for all other user stories.

**Independent Test**: Can be fully tested by starting the MCP server, sending a `tools/list` request, and verifying that the response includes every top-level resource command with accurate parameter names, types, and descriptions.

**Acceptance Scenarios**:

1. **Given** atadmin is installed and configured with a valid API token, **When** an agent sends a `tools/list` MCP request to the stdio server, **Then** the response lists one MCP tool per leaf subcommand (e.g., `users_list`, `users_get`, `groups_list`) with all flags represented as typed parameters.
2. **Given** a tool definition for `users_list`, **When** the agent inspects the parameter schema, **Then** it sees named parameters (e.g., `filter`, `limit`, `cursor`) with descriptions taken from the corresponding `--flag` usage strings.
3. **Given** a new command is added to atadmin in the future, **When** the MCP server starts, **Then** the new command appears automatically in the tool list with no additional registration code.

---

### User Story 2 - Agent Executes a Tool and Gets Structured Results (Priority: P2)

An AI agent calls a specific MCP tool (e.g., `users_list` with `{"limit": 10}`) and receives the result as structured JSON content — not raw terminal output — so the agent can reason over the data without parsing.

**Why this priority**: Discovery alone is useless if execution doesn't work. Agents must be able to call tools and get machine-readable results. This is the second critical requirement for real utility.

**Independent Test**: Can be fully tested by calling `users_list` via the MCP protocol and verifying the response contains structured JSON (not a table string) with the correct shape.

**Acceptance Scenarios**:

1. **Given** the MCP server is running, **When** an agent calls `users_list` with `{"limit": 5}`, **Then** the server returns a `tools/call` response with content containing a JSON array of user objects.
2. **Given** an agent calls a tool with an invalid parameter (e.g., `limit: "not-a-number"`), **Then** the MCP server returns a structured error response with a descriptive message, not a crash or unstructured stderr dump.
3. **Given** an API call fails (e.g., network error or 401), **When** the agent called a tool, **Then** the MCP response carries an `isError: true` flag and a human-readable error message.

---

### User Story 3 - Claude Desktop / Cursor Integration (Priority: P3)

A developer adds atadmin as an MCP server in their Claude Desktop or Cursor configuration file. After adding the server entry, the AI assistant within those tools can directly call atadmin commands as part of a conversation — listing users, checking alarm status, pulling audit logs — without the developer crafting shell commands.

**Why this priority**: This is the end-user integration story that validates the feature is production-ready. It depends on P1 and P2 working correctly, and demonstrates real-world value for the target audience.

**Independent Test**: Can be fully tested by adding the server config to Claude Desktop's `mcpServers` block, restarting Claude Desktop, and asking it to list users — verifying it routes the request through atadmin without any CLI invocation by the user.

**Acceptance Scenarios**:

1. **Given** Claude Desktop config includes `{"atadmin": {"command": "atadmin", "args": ["mcp", "serve", "--stdio"]}}`, **When** Claude Desktop starts, **Then** atadmin appears as a connected MCP server with all tools available.
2. **Given** the user asks Claude "show me the first 5 users in ActivTrak", **When** Claude routes this to the `users_list` tool, **Then** Claude receives structured data and presents it in a readable format.
3. **Given** the user has not configured an API token, **When** an agent calls any tool, **Then** the MCP server returns an actionable error directing the user to run `atadmin auth login`.

---

### Edge Cases

- What happens when `atadmin mcp serve` is run without a configured token? → Returns a structured MCP error on every tool call rather than crashing on startup.
- How does the server handle a tool call for a command that requires interactive TTY input (e.g., delete with confirmation prompt)? → Mutation commands requiring confirmation must accept a `--yes` parameter; the MCP tool definition must expose this as a boolean parameter. Mutation commands are only reachable when `--allow-mutations` is set.
- What happens if an agent sends a malformed JSON-RPC request? → The server returns a valid JSON-RPC error response and stays alive (does not exit).
- What happens when the agent sends a `tools/call` for a non-existent tool name? → Returns a structured error: tool not found.
- How are persistent global flags (e.g., `--profile`, `--format`, `--verbose`) handled? → Global flags are exposed as optional parameters on every tool that inherits them.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST expose a subcommand `mcp serve` that starts an MCP-protocol-compliant stdio server. All diagnostic output (startup messages, errors, debug traces) MUST be written to `~/.config/atadmin/mcp.log`; nothing diagnostic may be written to stdout or stderr during MCP operation.
- **FR-002**: The MCP server MUST automatically enumerate all leaf Cobra subcommands and expose each as a named MCP tool without requiring manual registration. By default, only read-only commands are exposed; mutation commands (update, delete, bulk actions) are included only when `mcp serve --allow-mutations` is specified.
- **FR-002a**: A command is classified as a mutation if its Cobra command name is one of: `update`, `delete`, `remove`, `bulk`, `create`, or if it is annotated as a mutation via a Cobra command annotation.
- **FR-003**: Each MCP tool definition MUST include: a unique tool name (derived from command path using underscores, e.g., `users_list`), a description sourced from the Cobra command's `Short` or `Long` field, and a JSON Schema parameter definition for each flag.
- **FR-004**: Flag types MUST be mapped to JSON Schema types: boolean flags → `boolean`, integer flags → `integer`, string flags → `string`.
- **FR-005**: When an agent calls a tool, the MCP server MUST invoke the corresponding Cobra command with the provided parameters and capture its output.
- **FR-006**: Tool call results MUST be returned as structured JSON content when the underlying command supports `--json` output; fall back to plain text content for commands without JSON output.
- **FR-007**: The MCP server MUST return a valid `isError: true` response (not crash) when the underlying command fails or the API returns an error.
- **FR-008**: The global flags `--profile`, `--format`, and `--verbose` MUST be available as optional parameters on every tool. The flags `--token` and `--base-url` MUST NOT be exposed as tool parameters to prevent credential leakage through MCP request logs; authentication must be configured via a profile before starting the server.
- **FR-009**: The server MUST correctly implement the MCP `initialize` handshake, including protocol version negotiation and capability advertisement.
- **FR-010**: The server MUST stay alive and continue serving after individual tool call errors.
- **FR-011**: Commands that support `--yes` for non-interactive confirmation MUST expose `yes` as a boolean parameter in their MCP tool definition.

### Key Entities

- **MCP Tool**: A named capability exposed over the MCP protocol. Corresponds 1:1 with a leaf Cobra subcommand. Has a name, description, and JSON Schema input definition.
- **MCP Server**: The long-running process started by `atadmin mcp serve --stdio`. Reads JSON-RPC 2.0 messages from stdin, writes responses to stdout.
- **Tool Name**: A string derived from the Cobra command path by joining path segments with underscores and lowercasing (e.g., `atadmin users list` → `users_list`).
- **Parameter**: An MCP tool input field. Derived from a Cobra command flag. Has a name, type, description, and required/optional designation.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All atadmin leaf commands (currently 20+) appear in the MCP tool list without any manual registration — validated by counting tools in `tools/list` response.
- **SC-002**: An agent can successfully call any read-only tool and receive a structured response within the same latency budget as the equivalent CLI invocation (no more than 2× overhead).
- **SC-003**: The MCP server handles 100 consecutive tool calls without exiting or requiring restart.
- **SC-004**: Integration with Claude Desktop requires zero code changes to Claude Desktop — only a one-line config entry.
- **SC-005**: A developer can add atadmin as an MCP server and have an AI assistant successfully list resources in under 5 minutes from first configuration.

## Clarifications

### Session 2026-05-19

- Q: Where should diagnostic logging go when running as an MCP stdio server? → A: Log to file at `~/.config/atadmin/mcp.log`
- Q: Should mutation commands (update/delete/bulk) be exposed as MCP tools? → A: Opt-in via `--allow-mutations` flag on `mcp serve`; read-only by default
- Q: Should `--token` and `--base-url` be exposed as per-tool parameters? → A: No — exclude both to prevent token leakage; expose only `--profile`, `--format`, `--verbose`

## Assumptions

- The existing atadmin Cobra command tree is the single source of truth for tool definitions; no separate registration file is required.
- `--stdio` is the only transport mode required for v1; HTTP/SSE transport is out of scope.
- The MCP protocol version targeted is the current stable version (2024-11-05 or later).
- Global flags inherited by all subcommands are exposed as optional parameters on every tool.
- The MCP server process exits cleanly when stdin is closed (the agent disconnects).
- All diagnostic logging is written to `~/.config/atadmin/mcp.log`; no diagnostic output goes to stdout (which is reserved for JSON-RPC) or stderr.
- Authentication is handled via the profile configured before server startup (e.g., `atadmin auth login` then `atadmin mcp serve`). The `--token` and `--base-url` flags are intentionally excluded from MCP tool parameters to prevent credentials from appearing in agent request logs.
- Commands that produce table output will return that output as a plain-text MCP content block when `--json` is not applicable.
- The `mcp` subcommand itself and its `serve` child are excluded from the tool list (a server should not advertise itself as a tool).
- By default (`mcp serve` without `--allow-mutations`), mutation commands are excluded from the tool list. The flag `--allow-mutations` must be explicitly passed to include them.
