# Implementation Plan: Context Window Protections

**Branch**: `005-context-window-protections` | **Date**: 2026-05-20 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-context-window-protections/spec.md`

## Summary

Add three AI-friendly output controls to all list commands so LLM agents can consume CLI output without overflowing their context windows:

1. **`--fields id,email`** — client-side top-level key filtering applied before JSON is written to stdout.
2. **Safe JSON pagination** — when `--json` is passed without an explicit limit, default to 50 items.
3. **`--summary`** — return aggregate statistics (`{"returned_items": N, "total_items": N, "has_more": bool}`) instead of the full array.

The JSON formatting layer (`internal/output`) gains two new helpers; every Cobra list command with `--json` support gains new flags wired to those helpers. The MCP mapper picks up the new flags automatically via its existing Cobra tree walk — no changes to `internal/mcp/`.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: `spf13/cobra`, `spf13/pflag`, standard `encoding/json`, `text/tabwriter`  
**Storage**: N/A  
**Testing**: `go test ./...` with `net/http/httptest` mocks; table-driven tests  
**Target Platform**: macOS / Linux CLI  
**Project Type**: CLI tool  
**Performance Goals**: No overhead above existing `output.JSON()` call; field filtering is O(n·k) where n = items and k = number of selected fields  
**Constraints**: Client-side only — no API changes; top-level keys only for V1  
**Scale/Scope**: Applied to ~11 list commands across 8 resource files in `internal/cmd/`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test Integrity | PASS | New tests only; existing tests untouched |
| II. Scope Discipline | PASS | Changes limited to `internal/output/` and `internal/cmd/` |
| III. Commit Cadence | PASS | One commit per task at implementation time |
| IV. Stop-and-Report | PASS | Applies at implementation time |
| V. Sandbox Requirement | PASS | No `--dangerously-skip-permissions` needed |
| VI. Spec Primacy | PASS | Spec is complete; no gaps requiring amendment |

## Project Structure

### Documentation (this feature)

```text
specs/005-context-window-protections/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── list-flags.md    # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
internal/
├── output/
│   ├── output.go          # + FilterFields(), SummaryResult, JSONSummary()
│   └── output_test.go     # + TestFilterFields, TestJSONSummary
└── cmd/
    ├── users.go           # list: + --fields, --summary, safe --limit (0→50)
    ├── groups.go          # list: + --fields, --summary (--page-size already defaults to 50)
    ├── clients.go         # list: + --fields, --summary (--page-size already defaults to 50)
    ├── consumers.go       # list: + --fields, --summary (--page-size already defaults to 50)
    ├── devices.go         # list: + --fields, --summary (--page-size already defaults to 50)
    ├── alarms.go          # list: + --fields, --summary (--page-size already defaults to 50)
    ├── signals.go         # list: + --fields only (no pagination)
    ├── schedules.go       # list: + --fields only (no pagination)
    ├── apikeys.go         # list: + --fields only (no pagination)
    ├── auditlog.go        # list: + --fields, --summary, safe --page-size (0→50)
    └── agents.go          # list: + --fields, --summary, safe --limit (0→50)
```

**Structure Decision**: Single-project, existing layout. No new packages. Two helper functions and one struct are added to `internal/output/output.go` alongside the existing `Table`, `KeyValue`, and `JSON` functions.

## Context Files

All files an implementation agent may touch for this feature:

- `internal/output/output.go`
- `internal/output/output_test.go`
- `internal/cmd/users.go`
- `internal/cmd/groups.go`
- `internal/cmd/clients.go`
- `internal/cmd/consumers.go`
- `internal/cmd/devices.go`
- `internal/cmd/alarms.go`
- `internal/cmd/signals.go`
- `internal/cmd/schedules.go`
- `internal/cmd/apikeys.go`
- `internal/cmd/auditlog.go`
- `internal/cmd/agents.go`

## Design Details

### 1. Output Layer (`internal/output/output.go`)

Two additions and no changes to existing functions:

**`FilterFields(data any, fields []string) any`**

Strips non-selected top-level keys before JSON serialization. Handles three shapes:
- `map[string]any` → returns new map containing only the requested keys
- `[]any` → iterates elements; applies map filter to any `map[string]any` elements, passes others through
- Anything else → returned unchanged (passthrough)

```go
func FilterFields(data any, fields []string) any {
    allowed := make(map[string]bool, len(fields))
    for _, f := range fields {
        allowed[strings.TrimSpace(f)] = true
    }
    switch v := data.(type) {
    case map[string]any:
        out := make(map[string]any, len(allowed))
        for k, val := range v {
            if allowed[k] { out[k] = val }
        }
        return out
    case []any:
        result := make([]any, len(v))
        for i, elem := range v {
            result[i] = FilterFields(elem, fields)
        }
        return result
    }
    return data
}
```

Called by list commands before `output.JSON()`:
```go
var data any = apiResult
if fieldsFlag != "" {
    data = output.FilterFields(apiResult, strings.Split(fieldsFlag, ","))
}
return output.JSON(cmd.OutOrStdout(), data)
```

**`SummaryResult` and `JSONSummary`**

```go
type SummaryResult struct {
    ReturnedItems int  `json:"returned_items"`
    TotalItems    *int `json:"total_items,omitempty"`
    HasMore       bool `json:"has_more"`
}

func JSONSummary(out io.Writer, returned int, total *int, hasMore bool) error {
    return JSON(out, SummaryResult{ReturnedItems: returned, TotalItems: total, HasMore: hasMore})
}
```

`total` is a pointer — commands whose API response doesn't include a total count pass `nil`, which omits the field via `omitempty`.

### 2. Safe JSON Pagination (Cobra command layer)

Two pagination patterns exist in the codebase:

| Pattern | Commands | Default | Action |
|---------|----------|---------|--------|
| `--limit int` (cursor) | `users list`, `agents list` | `0` (server default) | If `asJSON && !cmd.Flags().Changed("limit") && limit == 0` → set `limit = 50` |
| `--page-size int` (offset), default 50 | `groups`, `clients`, `consumers`, `devices`, `alarms` | `50` | Already safe — no change needed |
| `--page-size int` (offset), default 0 | `auditlog` | `0` | If `asJSON && !cmd.Flags().Changed("page-size") && pageSize == 0` → set `pageSize = 50` |
| No pagination | `signals`, `schedules`, `apikeys` | N/A | No change |

The `cmd.Flags().Changed("limit")` check ensures explicit user values are never overridden.

### 3. `--fields` Flag (Cobra command layer)

Added to every list command with a `--json` flag. The variable binds to a `string`; the value is a comma-separated list of top-level key names (e.g., `"id,email"`).

```go
var fieldsFlag string
cmd.Flags().StringVar(&fieldsFlag, "fields", "", "Comma-separated top-level JSON keys to include (e.g. id,email)")
```

The MCP mapper's `extractParams` will automatically expose `fields` as a `string` parameter on each MCP tool because it calls `cmd.Flags().VisitAll()`.

### 4. `--summary` Flag (Cobra command layer)

Added to list commands that return arrays (all except `signals`, `schedules`, `apikeys`). The flag short-circuits before `--fields`.

```go
var summaryFlag bool
cmd.Flags().BoolVar(&summaryFlag, "summary", false, "Return aggregate statistics instead of full results")

// In RunE, before --fields logic:
if summaryFlag && asJSON {
    hasMore := false  // derive from NextCursor or pagination metadata when available
    return output.JSONSummary(cmd.OutOrStdout(), len(items), nil, hasMore)
}
```

For cursor-based commands (`users list`, `agents list`) the page wrapper contains a `NextCursor` or equivalent field — use that to set `hasMore`.

### 5. MCP Compatibility

**No changes to `internal/mcp/`.** The mapper's `Walk` calls `extractParams` which uses `cmd.Flags().VisitAll()` to enumerate all flags at server startup. Once `--fields` and `--summary` are registered on each list command, they appear automatically in the tool definitions on the next `atadmin mcp serve` start.

The `makeHandler` in `server.go` already auto-injects `--json` for commands where `HasJSONFlag` is true, so the safe pagination behavior also fires automatically for MCP tool calls without any MCP-layer changes.

## Complexity Tracking

No constitution violations. No extra complexity justified.
