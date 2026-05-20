# Tasks: Embedded MCP Server for atadmin

**Input**: Design documents from `specs/004-mcp-server/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓, quickstart.md ✓

**Organization**: Tasks grouped by user story. Each story is independently implementable and testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel with other [P] tasks in the same phase (different files/functions)
- **[Story]**: Maps to user story from spec.md

---

## Phase 1: Setup

**Purpose**: Add the MCP library dependency and scaffold the new package.

- [x] T001 Add `github.com/mark3labs/mcp-go` dependency by running `go get github.com/mark3labs/mcp-go` and committing changes to `go.mod` and `go.sum`
- [x] T002 Create `internal/mcp/` package with empty stub files: `mapper.go` (package mcp) and `server.go` (package mcp)

---

## Phase 2: Foundational — Cobra-to-MCP Mapper

**Purpose**: The mapper is the core engine that translates the Cobra command tree into MCP tool definitions. All three user stories depend on it.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T003 Implement `Walk(root *cobra.Command, allowMutations bool) []ToolDef` in `internal/mcp/mapper.go` — recursively calls `cmd.Commands()`, identifies leaf commands (`len(cmd.Commands()) == 0`), excludes hidden commands (`cmd.Hidden`), deprecated commands (`cmd.Deprecated != ""`), the `mcp` subtree, and `help` commands; builds `ToolDef{Name, Description, CommandPath, HasJSONFlag, IsMutation}` for each leaf
- [x] T004 [P] Implement `isMutation(cmd *cobra.Command) bool` in `internal/mcp/mapper.go` — returns true when command name is one of: `update`, `delete`, `remove`, `bulk`, `add`, `create`; used by Walk() to gate `--allow-mutations` filtering
- [x] T005 [P] Implement `toolName(path []string) string` in `internal/mcp/mapper.go` — joins command path segments with underscore, lowercases, replaces dashes with underscores (e.g., `["audit-log", "list"]` → `"audit_log_list"`)
- [x] T006 [P] Implement `extractParams(cmd *cobra.Command) []ToolParam` in `internal/mcp/mapper.go` — calls `VisitAll` on both `cmd.Flags()` and `cmd.InheritedFlags()`, maps `pflag.Flag.Value.Type()` to JSON Schema types (`"bool"` → `"boolean"`, `"int"` → `"integer"`, else `"string"`), excludes flags named `help`, `token`, `base-url`; returns `[]ToolParam{Name, Type, Description, Default}`
- [x] T007 Write mapper unit tests in `internal/mcp/mapper_test.go` — table-driven tests covering: `toolName()` with dash-containing segments, `isMutation()` for each mutation verb and a non-mutation name, `extractParams()` verifying token/base-url exclusion and type mapping, `Walk()` with a synthetic three-level cobra tree verifying hidden/deprecated/mcp exclusion and allowMutations filtering

**Checkpoint**: `go test ./internal/mcp/` passes for mapper tests.

---

## Phase 3: User Story 1 — Agent Connects and Discovers All Tools (P1) 🎯 MVP

**Goal**: `atadmin mcp serve --stdio` starts a server that responds to `tools/list` with all read-only atadmin commands as typed MCP tools.

**Independent Test**: Run `atadmin mcp serve --stdio`, send `initialize` + `tools/list` via stdin, verify response contains 15+ tools with correct names and parameter schemas.

- [x] T008 [US1] Implement `MCPServer` struct and `NewServer(root *cobra.Command, allowMutations bool) *MCPServer` in `internal/mcp/server.go` — creates a `mcp-go` server instance (`server.NewMCPServer("atadmin", version)`), calls `Walk()` to get tool defs, builds and registers each tool via `s.AddTool(tool, handler)` where the handler is a placeholder returning "not implemented"
- [x] T009 [US1] [P] Implement `setupLogger() (*os.File, error)` in `internal/mcp/server.go` — ensures `~/.config/atadmin/` directory exists (`os.MkdirAll`), opens/creates `mcp.log` with `os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)`, returns the file handle; called by `NewServer()` which sets the process logger output to this file before any other operation
- [x] T010 [US1] [P] Implement `newMCPCmd()` and `newMCPServeCmd()` in `internal/cmd/mcp.go` — `mcp` parent command (no-op, shows help) and `serve` child with `--stdio` bool flag (always required; fail with actionable message if absent) and `--allow-mutations` bool flag; `RunE` calls `mcp.NewServer(NewRootCmd(), allowMutations)` then `server.Start()`
- [x] T011 [US1] Wire `newMCPCmd()` into the root command in `internal/cmd/root.go` by adding `root.AddCommand(newMCPCmd())` alongside the other resource commands

**Checkpoint**: `atadmin mcp serve --stdio` starts, responds to `tools/list`, exits cleanly on stdin close; `go test ./internal/mcp/ ./internal/cmd/` passes.

---

## Phase 4: User Story 2 — Agent Executes a Tool and Gets Structured Results (P2)

**Goal**: Agents can call any listed tool and receive either structured JSON or plain-text output; API errors surface as `isError: true` responses rather than crashes.

**Independent Test**: Call `users_list` tool via MCP with `{"limit": 5, "json": true}`, verify response content is valid JSON matching the `UsersPage` shape; call with invalid params, verify `isError: true` response.

- [x] T012 [US2] Implement `makeHandler(cmdPath []string, hasJSONFlag bool) mcp.ToolHandlerFunc` in `internal/mcp/server.go` — closure that receives a `mcp.CallToolRequest`, calls `buildArgs()` to reconstruct the CLI arg slice, creates a fresh `cmd.NewRootCmd()`, sets `SetOut(outBuf)` and `SetErr(errBuf)`, calls `rootCmd.SetArgs(args)` then `rootCmd.ExecuteC()`, returns `mcp.NewToolResultText(outBuf.String())` on success; replace the placeholder handlers registered in T008 with `makeHandler` instances
- [x] T013 [US2] Implement `buildArgs(cmdPath []string, params map[string]any, hasJSONFlag bool) []string` in `internal/mcp/server.go` — prepends `cmdPath` segments (e.g., `["users", "list"]`), auto-injects `--json` when `hasJSONFlag` is true and params does not contain `json: false`, then appends remaining params: bool `true` → `--name`, bool `false` → omit, integer/string → `--name`, `"value"`; handles positional arg convention where a param named `"id"` is appended bare (not as `--id`) for commands like `users get`
- [x] T014 [US2] Implement error capture in `makeHandler()` in `internal/mcp/server.go` — when `ExecuteC()` returns a non-nil error or `errBuf` is non-empty, return `mcp.NewToolResultError(combinedMessage)` with `isError: true`; combined message = stderr content if non-empty, else error.Error()
- [x] T015 [US2] Write server integration tests in `internal/mcp/server_test.go` — use an in-process approach: create a `NewServer()` with a synthetic Cobra root (containing a `ping` leaf command that writes `{"ok":true}` to stdout), call `makeHandler()` directly with `{"json": false}` params, assert the result content equals `{"ok":true}` and `IsError` is false; add a test where the command returns exit error and assert `IsError` is true

**Checkpoint**: `go test ./internal/mcp/` passes; calling `users_list` via MCP returns JSON content when the API is reachable.

---

## Phase 5: User Story 3 — Claude Desktop / Cursor Integration (P3)

**Goal**: A developer configures Claude Desktop or Cursor with a one-line entry and AI assistants can list resources without manual CLI invocation.

**Independent Test**: Add atadmin to Claude Desktop `mcpServers` config, restart, ask Claude to list users — it invokes `users_list` and returns data.

- [x] T016 [US3] Implement actionable auth error detection in `makeHandler()` in `internal/mcp/server.go` — after capturing stderr, check if it contains "unauthorized" or "401"; if so, prefix the error message with: `"API authentication failed. Run 'atadmin auth login' to configure credentials, then restart the MCP server."` before returning the `isError: true` response
- [x] T017 [US3] [P] Update `specs/004-mcp-server/quickstart.md` with: (1) the log file location `~/.config/atadmin/mcp.log` for debugging, (2) the `--allow-mutations` flag example for mutation access, (3) correct note that `--token`/`--base-url` are not available as tool parameters and auth must be pre-configured via `atadmin auth login`

**Checkpoint**: Claude Desktop connects successfully and `users_list` tool returns data; `~/.config/atadmin/mcp.log` contains timestamped startup entries.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation accuracy, lint compliance, and build verification.

- [x] T018 [P] Update `specs/004-mcp-server/contracts/mcp-tools.md` to reflect the three clarified decisions: `--allow-mutations` flag, excluded parameters (`token`, `base-url`), and log file location
- [x] T019 [P] Update `specs/004-mcp-server/plan.md` Design Decisions section to add D-006 (logging to `~/.config/atadmin/mcp.log`), D-007 (read-only by default / `--allow-mutations`), D-008 (token/base-url excluded from tool params)
- [x] T020 Run `go test ./...` and fix any failures
- [x] T021 [P] Run `golangci-lint run` and fix any reported issues in `internal/mcp/`
- [x] T022 Run `go build -o bin/atadmin ./cmd/atadmin` and verify the binary builds cleanly

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user story work
- **US1 (Phase 3)**: Depends on Phase 2 (mapper must exist before server can use it)
- **US2 (Phase 4)**: Depends on Phase 3 (makeHandler replaces placeholder handlers from T008)
- **US3 (Phase 5)**: Depends on Phase 4 (auth error detection builds on isError handling from T014)
- **Polish (Phase 6)**: Depends on all story phases

### Within-Phase Parallel Opportunities

- **Phase 2**: T004, T005, T006 can be written in parallel (each is a standalone function in mapper.go with no inter-dependency)
- **Phase 3**: T009 and T010 can be written in parallel (different files: server.go vs mcp.go)
- **Phase 6**: T018, T019, T021 can be done in parallel (different files)

### Positional Argument Convention

Commands that accept a positional `<id>` argument (e.g., `users get`, `users update`, `users delete`) expose it as a string parameter named `"id"` in their MCP tool schema. `buildArgs()` detects `"id"` in the params map and appends it bare rather than as `--id`.

---

## Parallel Execution Examples

### Phase 2 Parallel (Mapper Functions)

```
Parallel:
  Task T004: isMutation() function in internal/mcp/mapper.go
  Task T005: toolName() function in internal/mcp/mapper.go
  Task T006: extractParams() function in internal/mcp/mapper.go
Then sequential:
  Task T003: Walk() integrates the above three functions
  Task T007: mapper tests
```

### Phase 3 Parallel (Server + CLI Wiring)

```
Parallel:
  Task T009: setupLogger() in internal/mcp/server.go
  Task T010: newMCPCmd() in internal/cmd/mcp.go
Then sequential:
  Task T008: NewServer() calls setupLogger() and Walk()
  Task T011: wire into root.go
```

---

## Implementation Strategy

### MVP (User Story 1 Only)

1. Phase 1: Add dependency, scaffold package
2. Phase 2: Build mapper (T003–T007)
3. Phase 3: Wire server + CLI (T008–T011)
4. **STOP and VALIDATE**: `tools/list` returns correct tool names and schemas
5. Ship as read-only tool-discovery MVP

### Full Feature Delivery

1. MVP complete (P1 done)
2. Add tool execution (P2 — T012–T015)
3. Add auth error messages + integration docs (P3 — T016–T017)
4. Polish (T018–T022)

---

## Notes

- All `internal/mcp/` code must pass `golangci-lint run` before marking complete
- `makeHandler()` must use `cmd.NewRootCmd()` (a new instance per call) — never reuse a Cobra root across tool calls
- The MCP server must NOT write anything to stdout or stderr after startup; all output goes to `~/.config/atadmin/mcp.log`
- The `--json` flag is auto-injected by `buildArgs()` when `hasJSONFlag` is true; callers do not need to pass it explicitly
