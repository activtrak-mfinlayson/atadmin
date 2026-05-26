# Tasks: Safe Exploration (--dry-run)

**Input**: Design documents from `/specs/008-dry-run/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/dry-run-output.md ✓

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2)
- Exact file paths included in every implementation task

---

## Phase 1: Setup

**Purpose**: Establish a clean baseline before modifying shared API client code.

- [X] T001 Run `go build -o bin/atadmin ./cmd/atadmin` and `go test ./...` to confirm baseline compiles and all existing tests pass before changes

**Checkpoint**: Baseline green — safe to proceed

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add shared types and helper functions that both user stories depend on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T002 Add `DryRunOutput` struct with `Action string`, `Target string`, and `Payload json.RawMessage` JSON fields to `internal/api/client.go`
- [X] T003 [P] Add `isMutatingMethod(method string) bool` helper (returns true for POST, PUT, PATCH, DELETE) to `internal/api/helpers.go`
- [X] T004 [P] Add `httpMethodToAction(method string) string` helper (POST→`"create"`, PUT/PATCH→`"update"`, DELETE→`"delete"`) to `internal/api/helpers.go`

**Checkpoint**: Foundation ready — US1 and US2 can proceed

---

## Phase 3: User Story 1 — Dry Run Flag & Interception (Priority: P1) 🎯 MVP

**Goal**: `--dry-run` is available as a persistent root-level flag; when set, every mutating command (`POST`, `PUT`, `PATCH`, `DELETE`) makes zero HTTP requests to the remote API.

**Independent Test**:
```bash
atadmin clients alias set <id> test-alias --dry-run
```
Verify: exits 0, no network traffic, stdout contains a JSON line.

### Implementation for User Story 1

- [X] T005 [US1] Add `DryRun bool` and `Out io.Writer` fields to the `Client` struct in `internal/api/client.go`
- [X] T006 [US1] Append `dryRun bool, out io.Writer` parameters to `NewClient()` and assign them to `client.DryRun` and `client.Out` in `internal/api/client.go`
- [X] T007 [US1] Add `var dryRunFlag bool` and `root.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Preview the action without executing it (prints JSON to stdout)")` inside `NewRootCmd()` in `internal/cmd/root.go`
- [X] T008 [US1] Pass `dryRunFlag, os.Stdout` as the final two arguments to the `api.NewClient()` call inside `PersistentPreRunE` in `internal/cmd/root.go`
- [X] T009 [US1] Add dry-run short-circuit block in `doRequest()` in `internal/api/helpers.go`: if `c.DryRun && isMutatingMethod(method)`, marshal `body` as `json.RawMessage` (use `json.RawMessage("null")` when `body == nil`), encode a `DryRunOutput{Action, Target, Payload}` to `c.Out`, and return `&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(nil))}, nil`

### Tests for User Story 1

- [X] T010 [P] [US1] Write httptest table-driven test in `internal/api/client_test.go` asserting the test server receives **zero requests** for POST, PUT, PATCH, DELETE when `client.DryRun = true`; server should record all incoming requests and assert count == 0

**Checkpoint**: User Story 1 complete — `--dry-run` suppresses all mutating HTTP calls and exits 0

---

## Phase 4: User Story 2 — JSON Output Contract (Priority: P2)

**Goal**: When `--dry-run` intercepts a mutating request, it writes exactly one valid JSON line to stdout matching the schema `{"action":"...","target":"...","payload":...}` per `contracts/dry-run-output.md`.

**Independent Test**:
```bash
atadmin groups rename 42 "Engineering" --dry-run | jq .
```
Verify: `{"action":"update","target":"/admin/v1/groups/42","payload":{"name":"Engineering"}}`.

### Tests for User Story 2

- [X] T011 [P] [US2] Write httptest table-driven test in `internal/api/client_test.go` verifying that `DryRunOutput` JSON written to `client.Out` has correct `action` field for all four mutating methods (POST→`"create"`, PUT→`"update"`, PATCH→`"update"`, DELETE→`"delete"`) and correct `target` field matching the request path
- [X] T012 [P] [US2] Write httptest test in `internal/api/client_test.go` verifying `payload` encodes as JSON `null` when `body == nil` (e.g. a DELETE with no request body)
- [X] T013 [US2] Write httptest test in `internal/api/client_test.go` verifying that a `GET` request is **not** intercepted when `client.DryRun = true` — the test server must still receive the request
- [X] T014 [US2] Write CLI-level test in `internal/cmd/root_test.go` that sets `--dry-run` on a mutating cobra command, captures stdout via `cmd.SetOut()`, and asserts parsed JSON has `action`, `target`, and `payload` fields per the output contract

**Checkpoint**: All acceptance criteria from spec.md satisfied — no-op interception (US1) and correct JSON output (US2) both verified

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Ensure project-wide quality and validate real-world usage per quickstart.

- [X] T015 [P] Run `go test ./...` and fix any compilation errors or test failures introduced by the new parameters to `NewClient()`
- [X] T016 [P] Run `golangci-lint run` and fix any lint warnings (unused imports, shadow variables, missing error checks)
- [X] T017 Validate quickstart.md scenario manually: `atadmin groups rename 42 "New Name" --dry-run` outputs `{"action":"update","target":"/admin/v1/groups/42","payload":{"name":"New Name"}}` and makes no network call
- [X] T018 Validate quickstart.md scenario manually: `atadmin clients list --dry-run` and `atadmin groups get 42 --dry-run` execute normally with no dry-run JSON output (read-only commands unaffected)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS both user stories
- **US1 (Phase 3)**: Depends on Foundational — T005 requires T002 (DryRunOutput type); T009 requires T003 + T004 (helpers)
- **US2 (Phase 4)**: T011–T013 depend on US1 complete (need the short-circuit in `doRequest()` to exist); T014 depends on T007–T008 (CLI flag wired up)
- **Polish (Phase 5)**: Depends on US1 and US2 complete

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 2 only — no dependency on US2
- **US2 (P2)**: T011–T013 require US1 Phase 3 to be complete; they test behavior introduced by T009

### Within User Story 1

- T005 → T006 (need fields before updating constructor)
- T006 → T008 (need new signature before calling it in root.go)
- T007 can be done in parallel with T005–T006 (different file: `root.go`)
- T009 depends on T005 (needs `c.DryRun`, `c.Out`) and T003–T004 (helpers)
- T010 (test) can be written as a skeleton before T009 — TDD approach

### Parallel Opportunities

- T003 and T004 (Phase 2): different functions, both in `helpers.go` — can write simultaneously
- T005 and T007: different files (`client.go` vs `root.go`) — can work in parallel
- T010 (US1 test) and T011–T013 (US2 tests): different test assertions, all in `client_test.go` — can be written in parallel once T009 is in place
- T015 and T016 (Polish): independent tools, fully parallel

---

## Parallel Example: User Story 1

```bash
# Phase 2 parallel work (different functions, same file):
Task T003: Add isMutatingMethod() in internal/api/helpers.go
Task T004: Add httpMethodToAction() in internal/api/helpers.go

# US1 parallel start (different files):
Task T005: Add DryRun/Out fields to Client struct in internal/api/client.go
Task T007: Add --dry-run persistent flag in internal/cmd/root.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Baseline green
2. Complete Phase 2: DryRunOutput type + helpers
3. Complete Phase 3: US1 flag + interception
4. **STOP and VALIDATE**: Run `go test ./...`; manually run `atadmin clients alias set <id> test --dry-run` and confirm no HTTP call and JSON output
5. Demo or merge if MVP accepted

### Incremental Delivery

1. Setup + Foundational → types and helpers ready
2. US1 complete → mutating commands suppressed, JSON output emitted → validate independently
3. US2 complete → output contract tests pass → all acceptance criteria met
4. Polish → lint clean, quickstart validated, full test suite green

---

## Notes

- [P] tasks = different files or independently writable sections, no blocking dependencies
- [Story] label maps each task to its acceptance criterion in `specs/008-dry-run/spec.md`
- The synthetic `&http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(nil))}` returned by the dry-run block must pass the existing `checkResponse()` check without any changes to caller code
- `NewClient()` signature change in T006 will require updating **all call sites** — search `internal/cmd/` for `api.NewClient(` before implementing
- US1 and US2 share the same code path in `doRequest()` but are independently testable via `client.Out` capture
