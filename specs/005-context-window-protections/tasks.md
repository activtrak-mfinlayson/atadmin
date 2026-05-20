# Tasks: Context Window Protections

**Input**: Design documents from `/specs/005-context-window-protections/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/list-flags.md ✓, quickstart.md ✓

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no shared dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Exact file paths included in all descriptions

---

## Phase 1: Setup

**Purpose**: Confirm the project builds cleanly before adding changes.

- [X] T001 Verify project builds and all existing tests pass by running `go test ./...` and `go build -o bin/atadmin ./cmd/atadmin`

---

## Phase 2: Foundational (Blocking Prerequisite)

**Purpose**: Add the two new output helpers that all list commands depend on. No command-layer work can begin until this phase is complete.

**CRITICAL**: All Phase 3–5 tasks depend on this phase completing first.

- [X] T002 Add `FilterFields(data any, fields []string) any` function to `internal/output/output.go` — handles `map[string]any`, `[]any`, and passthrough for other types as specified in plan.md §1
- [X] T003 Add `SummaryResult` struct and `JSONSummary(out io.Writer, returned int, total *int, hasMore bool) error` function to `internal/output/output.go` using `*int` for `TotalItems` with `omitempty` as specified in data-model.md
- [X] T004 Add `TestFilterFields` and `TestJSONSummary` table-driven tests to `internal/output/output_test.go` covering: single object filtering, array filtering, passthrough for non-map types, nil total omitted from JSON, has_more true/false

**Checkpoint**: `go test ./internal/output/...` passes — helpers are ready for use in command files.

---

## Phase 3: User Story 1 — Field Filtering (`--fields`) (Priority: P1) MVP

**Goal**: Every list command with `--json` support accepts `--fields id,email` (or any comma-separated top-level keys) and strips all other keys before writing to stdout.

**Independent Test**: `atadmin users list --json --fields id,email | jq '.[0] | keys'` outputs `["email","id"]` only; running with `--fields nonexistent` returns objects with zero keys and no error.

### Implementation

- [X] T005 [P] [US1] Add `--fields` flag and `output.FilterFields` call in `internal/cmd/users.go` list command — declare `var fieldsFlag string`, register `cmd.Flags().StringVar(&fieldsFlag, "fields", "", "...")`, apply `output.FilterFields(result, strings.Split(fieldsFlag, ","))` before `output.JSON()` when `fieldsFlag != ""`
- [X] T006 [P] [US1] Add `--fields` flag and `output.FilterFields` call in `internal/cmd/groups.go` list command
- [X] T007 [P] [US1] Add `--fields` flag and `output.FilterFields` call in `internal/cmd/clients.go` list command
- [X] T008 [P] [US1] Add `--fields` flag and `output.FilterFields` call in `internal/cmd/consumers.go` list command
- [X] T009 [P] [US1] Add `--fields` flag and `output.FilterFields` call in `internal/cmd/devices.go` list command
- [X] T010 [P] [US1] Add `--fields` flag and `output.FilterFields` call in `internal/cmd/alarms.go` list command
- [X] T011 [P] [US1] Add `--fields` flag and `output.FilterFields` call in `internal/cmd/signals.go` list command (no pagination on this command; `--fields` only per plan.md)
- [X] T012 [P] [US1] Add `--fields` flag and `output.FilterFields` call in `internal/cmd/schedules.go` list command (no pagination on this command; `--fields` only per plan.md)
- [X] T013 [P] [US1] Add `--fields` flag and `output.FilterFields` call in `internal/cmd/apikeys.go` list command (no pagination on this command; `--fields` only per plan.md)
- [X] T014 [P] [US1] Add `--fields` flag and `output.FilterFields` call in `internal/cmd/auditlog.go` list command
- [X] T015 [P] [US1] Add `--fields` flag and `output.FilterFields` call in `internal/cmd/agents.go` list command

**Checkpoint**: All 11 list commands accept `--fields`; `go test ./...` passes; `go build` succeeds.

---

## Phase 4: User Story 2 — Safe JSON Pagination (Priority: P2)

**Goal**: When `--json` is passed without an explicit `--limit` (or `--page-size`), the CLI automatically caps results at 50 items, preventing context window overflow for LLM agents.

**Independent Test**: `atadmin users list --json | jq length` returns ≤ 50; `atadmin users list --json --limit 200 | jq length` returns up to 200 (explicit value always respected).

**Affected commands**: Only `users list`, `agents list`, and `auditlog list` have unsafe zero defaults. The remaining paginated commands (`groups`, `clients`, `consumers`, `devices`, `alarms`) already default `--page-size` to 50 and need no change per contracts/list-flags.md.

### Implementation

- [X] T016 [P] [US2] Add safe pagination guard to `internal/cmd/users.go` list command — add `if asJSON && !cmd.Flags().Changed("limit") && limit == 0 { limit = 50 }` before the API call
- [X] T017 [P] [US2] Add safe pagination guard to `internal/cmd/agents.go` list command — same `limit` pattern as T016
- [X] T018 [P] [US2] Add safe pagination guard to `internal/cmd/auditlog.go` list command — add `if asJSON && !cmd.Flags().Changed("page-size") && pageSize == 0 { pageSize = 50 }` before the API call

**Checkpoint**: `atadmin users list --json | jq length` returns ≤ 50; explicit `--limit` values still override the default.

---

## Phase 5: User Story 3 — Summary Mode (`--summary`) (Priority: P3)

**Goal**: List commands that return paginated arrays accept `--summary` with `--json` and emit `{"returned_items": N, "has_more": bool}` instead of the full array. `--summary` short-circuits before `--fields` (summary wins when both flags are present).

**Independent Test**: `atadmin groups list --json --summary` outputs `{"returned_items": N, "has_more": false/true}`; combining `--summary --fields id` still returns summary only (not filtered data).

**Affected commands**: `users`, `groups`, `clients`, `consumers`, `devices`, `alarms`, `auditlog`, `agents` (8 total). `signals`, `schedules`, `apikeys` excluded per research.md — non-paginated results offer no meaningful `has_more`.

### Implementation

- [X] T019 [P] [US3] Add `--summary` flag and `output.JSONSummary` short-circuit to `internal/cmd/users.go` list command — declare `var summaryFlag bool`, register `cmd.Flags().BoolVar`, add `if summaryFlag && asJSON { return output.JSONSummary(...) }` before `--fields` logic; derive `hasMore` from `response.NextCursor != ""`
- [X] T020 [P] [US3] Add `--summary` flag and `output.JSONSummary` short-circuit to `internal/cmd/groups.go` list command; derive `hasMore` from returned count vs page-size
- [X] T021 [P] [US3] Add `--summary` flag and `output.JSONSummary` short-circuit to `internal/cmd/clients.go` list command
- [X] T022 [P] [US3] Add `--summary` flag and `output.JSONSummary` short-circuit to `internal/cmd/consumers.go` list command
- [X] T023 [P] [US3] Add `--summary` flag and `output.JSONSummary` short-circuit to `internal/cmd/devices.go` list command
- [X] T024 [P] [US3] Add `--summary` flag and `output.JSONSummary` short-circuit to `internal/cmd/alarms.go` list command
- [X] T025 [P] [US3] Add `--summary` flag and `output.JSONSummary` short-circuit to `internal/cmd/auditlog.go` list command; derive `hasMore` from `len(items) == pageSize`
- [X] T026 [P] [US3] Add `--summary` flag and `output.JSONSummary` short-circuit to `internal/cmd/agents.go` list command; derive `hasMore` from `response.NextCursor != ""`

**Checkpoint**: All 8 paginated list commands return aggregate stats with `--json --summary`; `go test ./...` passes.

---

## Phase 6: Polish & Validation

**Purpose**: Final build verification and acceptance criteria validation from spec.md.

- [X] T027 Run `go test ./...` — confirm all tests pass with no regressions across the entire module
- [X] T028 Run `go build -o bin/atadmin ./cmd/atadmin` — confirm binary builds cleanly
- [X] T029 Validate all three spec.md acceptance criteria using quickstart.md commands: (1) `atadmin users list --json --fields id,email | jq '.[0] | keys'` → `["email","id"]`, (2) `atadmin users list --json | jq length` → ≤ 50, (3) `atadmin groups list --json --summary` → aggregate stats object
- [X] T030 Run `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` — confirm no new vulnerabilities introduced

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — **BLOCKS all user story phases**
- **Phase 3 (US1 — `--fields`)**: Depends on Phase 2 (needs `FilterFields`); independent of US2 and US3
- **Phase 4 (US2 — Pagination)**: Depends on Phase 2 only; independent of US1 and US3
- **Phase 5 (US3 — `--summary`)**: Depends on Phase 2 (needs `JSONSummary`); independent of US1 and US2
- **Phase 6 (Polish)**: Depends on all user story phases complete

### User Story Dependencies

- **US1 (P1)**: No dependency on US2 or US3 — implement independently
- **US2 (P2)**: No dependency on US1 or US3 — implement independently
- **US3 (P3)**: No dependency on US1 or US2 — implement independently

> **Note on same-file edits**: US1, US2, and US3 all modify `users.go`, `agents.go`, and `auditlog.go`. A single-developer workflow may choose to apply all changes to one file at a time rather than making three passes. The phase structure above reflects logical story isolation; collapse per team capacity.

### Within Each Phase

- T002 → T003 → T004 are sequential (same output package files)
- T005–T015 are all [P] — 11 different files, no ordering constraint within US1
- T016–T018 are all [P] — 3 different files, no ordering constraint within US2
- T019–T026 are all [P] — 8 different files, no ordering constraint within US3
- T027 → T028 → T029 → T030 are sequential (validation pipeline)

---

## Parallel Execution Examples

### Phase 2 (Foundational — sequential, same files)

```
T002: Add FilterFields to output.go
T003: Add SummaryResult + JSONSummary to output.go   (after T002)
T004: Add TestFilterFields + TestJSONSummary         (after T003)
```

### Phase 3 (US1 — all parallel, 11 different files)

```
Launch simultaneously after Phase 2:
T005: users.go      T006: groups.go    T007: clients.go
T008: consumers.go  T009: devices.go   T010: alarms.go
T011: signals.go    T012: schedules.go T013: apikeys.go
T014: auditlog.go   T015: agents.go
```

### Phases 3 + 4 + 5 in parallel (after Phase 2, with 3 developers)

```
Developer A: T005–T015 (US1 --fields, 11 files)
Developer B: T016–T018 (US2 safe pagination, 3 files)
Developer C: T019–T026 (US3 --summary, 8 files)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. T001: Verify build
2. T002–T004: Output layer foundation
3. T005–T015: `--fields` on all 11 list commands
4. **STOP and VALIDATE**: `atadmin users list --json --fields id,email | jq '.[0] | keys'` → `["email","id"]`
5. Ship — agents immediately benefit from field filtering with no context risk

### Incremental Delivery

1. T001–T004: Foundation ready
2. T005–T015 (US1): `--fields` shipped
3. T016–T018 (US2): Safe pagination shipped — agents get ≤ 50 items by default
4. T019–T026 (US3): `--summary` shipped — agents can count without loading data
5. T027–T030: Final validation and polish

### Single-Developer Shortcut (modify each file once)

Group all three user story changes per file to avoid revisiting files:

```
users.go:    T005 (--fields) + T016 (safe pagination) + T019 (--summary)
agents.go:   T015 (--fields) + T017 (safe pagination) + T026 (--summary)
auditlog.go: T014 (--fields) + T018 (safe pagination) + T025 (--summary)
groups.go:   T006 (--fields) + T020 (--summary)
clients.go:  T007 (--fields) + T021 (--summary)
consumers.go:T008 (--fields) + T022 (--summary)
devices.go:  T009 (--fields) + T023 (--summary)
alarms.go:   T010 (--fields) + T024 (--summary)
signals.go:  T011 (--fields only)
schedules.go:T012 (--fields only)
apikeys.go:  T013 (--fields only)
```

---

## Notes

- No changes to `internal/mcp/` — new flags are picked up automatically via `VisitAll` at server start (confirmed in research.md)
- `--summary` without `--json` is silently ignored (table mode passthrough)
- `--fields` with non-existent keys silently returns empty objects — no error (per contract)
- `groups`, `clients`, `consumers`, `devices`, `alarms` already default `--page-size` to 50 — no pagination guard tasks needed for those commands
- MCP `makeHandler` auto-injects `--json`, so safe pagination and all new flags work for MCP tool calls automatically
- [P] tasks = different files, no blocking dependencies within the same phase
- [Story] label maps each task to its user story for traceability
