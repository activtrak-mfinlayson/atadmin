# Tasks: Input via Stdin (`--from-stdin`)

**Input**: Design documents from `specs/007-stdin-input/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/stdin-payloads.md ✓

**Organization**: Tasks are grouped by command group (A, B, C from plan.md) to enable independent implementation and testing of each group.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1 = Group A, US2 = Group B, US3 = Group C)
- Exact file paths included in every description

---

## Phase 1: Setup — `internal/stdin` Package

**Purpose**: Create the new transport package that all command groups depend on

- [ ] T001 Create `internal/stdin/stdin.go`: implement `ReadJSON[T any](r io.Reader) (T, error)` and `ReadRecords(r io.Reader) ([]map[string]any, error)` per `specs/007-stdin-input/data-model.md`; error messages must match contract: `--from-stdin: stdin is empty; pipe a JSON payload`, `--from-stdin: invalid JSON: <err>`, `--from-stdin: reading stdin: <err>`

---

## Phase 2: Foundational — Tests for `internal/stdin`

**Purpose**: Validate the transport helpers before any command uses them

**⚠️ CRITICAL**: No command work can begin until T001 is complete and T002 passes

- [ ] T002 Create `internal/stdin/stdin_test.go`: table-driven tests for `ReadJSON` and `ReadRecords` covering empty reader (expect error), invalid JSON syntax (expect error), valid JSON object (expect typed struct), valid JSON array (expect `[]map[string]any`), and `io.Reader` read error

**Checkpoint**: `go test ./internal/stdin/...` passes — command implementation can now begin in parallel

---

## Phase 3: User Story 1 — Group A: File-Record Commands (P1) 🎯 MVP

**Goal**: All 16 file-record commands accept `--from-stdin` as a mutually exclusive alternative to `--file`, reading a JSON array via `stdin.ReadRecords`

**Independent Test**: `echo '[{"sourceId":1,"targetId":2}]' | atadmin clients merge-bulk --from-stdin` runs successfully; `atadmin clients merge-bulk --file f.json --from-stdin` returns `Error: --file and --from-stdin are mutually exclusive`

- [ ] T003 [P] [US1] Add `--from-stdin` bool flag to `clients merge-bulk`, `clients unmerge-bulk`, `clients alias-bulk`, `clients donottrack add-bulk`, `clients donottrack remove-bulk` in `internal/cmd/clients.go`; call `stdin.ReadRecords(os.Stdin)` when set; return error if `--file` is also set
- [ ] T004 [P] [US1] Add `--from-stdin` bool flag to `consumers delete-bulk`, `consumers chrome-users import`, `consumers create`, `consumers update` in `internal/cmd/consumers.go`; call `stdin.ReadRecords(os.Stdin)` when set; return error if `--file` is also set
- [ ] T005 [P] [US1] Add `--from-stdin` bool flag to `groups members import` in `internal/cmd/groups.go`; call `stdin.ReadRecords(os.Stdin)` when set; return error if `--file` is also set
- [ ] T006 [P] [US1] Add `--from-stdin` bool flag to `hrdc import` in `internal/cmd/hrdc.go`; call `stdin.ReadRecords(os.Stdin)` when set; return error if `--file` is also set
- [ ] T007 [P] [US1] Add `--from-stdin` bool flag to `schedules create` in `internal/cmd/schedules.go`; call `stdin.ReadRecords(os.Stdin)` when set; return error if `--file` is also set
- [ ] T008 [P] [US1] Add `--from-stdin` bool flag to `signals create` and `signals update` in `internal/cmd/signals.go`; call `stdin.ReadRecords(os.Stdin)` when set; return error if `--file` is also set
- [ ] T009 [P] [US1] Add `--from-stdin` bool flag to `alarms create` and `alarms update` in `internal/cmd/alarms.go`; call `stdin.ReadRecords(os.Stdin)` when set; return error if `--file` is also set

**Checkpoint**: All Group A commands accept `--from-stdin`; mutual-exclusivity error works; `go test ./...` passes

---

## Phase 4: User Story 2 — Group B: Typed-Struct Mutation Commands (P2)

**Goal**: `users update` and all `users bulk *` commands accept a full typed JSON payload from stdin, replacing individual flags

**Independent Test**: `echo '{"displayName":"Jane"}' | atadmin users update 123 --from-stdin` patches only `displayName`; `echo '{"actions":["StartTracking"],"data":[{"entityId":1,"revision":1}]}' | atadmin users bulk start-tracking --from-stdin` executes; wrong `actions` value returns a validation error

- [ ] T010 [US2] Add `--from-stdin` bool flag to `users update` in `internal/cmd/users.go`; when set, call `stdin.ReadJSON[api.UpdateUserRequest](os.Stdin)` as the sole payload source; skip per-flag "at least one field required" validation; ignore individual field flags (`--display-name`, `--tracked`, etc.)
- [ ] T011 [US2] Add `--from-stdin` bool flag to `users bulk start-tracking`, `users bulk stop-tracking`, `users bulk delete-entity`, `users bulk delete-data` in `internal/cmd/users.go`; when set, call `stdin.ReadJSON[api.BulkActionRequest](os.Stdin)`; validate that `req.Actions[0]` matches the subcommand's expected action string (e.g. `"StartTracking"`), returning `--from-stdin: actions mismatch: expected [StartTracking]` on failure

**Checkpoint**: All Group B commands accept `--from-stdin`; individual flags ignored when `--from-stdin` set; action validation works; `go test ./...` passes

---

## Phase 5: User Story 3 — Group C: Confirmation-Bypass Commands (P3)

**Goal**: `users delete` skips the y/N interactive prompt; `consumers password set` reads a JSON password payload from stdin instead of prompting

**Independent Test**: `echo '' | atadmin users delete 123 --from-stdin` deletes without prompting; `echo '{"password":"s3cr3t"}' | atadmin consumers password set 456 --from-stdin` sets the password without interactive input; empty `password` field returns `Error: --from-stdin: "password" field is required`

- [ ] T012 [P] [US3] Add `--from-stdin` bool flag to `users delete` in `internal/cmd/users.go`; when set, assign `skipConfirmation = true` to skip the `tty.IsTerminal()` scanner prompt; do not read any data from stdin
- [ ] T013 [P] [US3] Add `--from-stdin` bool flag to `consumers password set` in `internal/cmd/consumers.go`; define unexported `passwordStdinPayload` struct (`Password string \`json:"password"\``); when set, call `stdin.ReadJSON[passwordStdinPayload](os.Stdin)`, validate `payload.Password != ""` (error: `--from-stdin: "password" field is required`), bypass `readPassword` call

**Checkpoint**: Both Group C commands bypass interactive prompts when `--from-stdin` is set; empty password validation works; `go test ./...` passes

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Build verification, error message consistency across all modified files

- [ ] T014 [P] Audit all `--from-stdin` error messages in `internal/cmd/clients.go`, `internal/cmd/consumers.go`, `internal/cmd/groups.go`, `internal/cmd/hrdc.go`, `internal/cmd/schedules.go`, `internal/cmd/signals.go`, `internal/cmd/alarms.go`, `internal/cmd/users.go`; confirm every error starts with `--from-stdin:` per `specs/007-stdin-input/contracts/stdin-payloads.md`
- [ ] T015 [P] Run `go build -o bin/atadmin ./cmd/atadmin` and fix any compilation errors
- [ ] T016 Run `go test ./...` and confirm all tests pass with no regressions

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on T001 — **BLOCKS all user story work**
- **Phase 3 (US1 — Group A)**: All T003–T009 depend on T001; independent of each other (different files)
- **Phase 4 (US2 — Group B)**: T010 must complete before T011 (same file: `internal/cmd/users.go`)
- **Phase 5 (US3 — Group C)**: T012 depends on Phase 4 completing (same file: `users.go`); T013 depends on T004 completing (same file: `consumers.go`)
- **Phase 6 (Polish)**: Depends on all user story phases completing

### User Story Dependencies

- **US1 (P1)**: Depends on T001 only — can start immediately after Phase 1
- **US2 (P2)**: Depends on T001 only — can start after Phase 1 (parallel with US1)
- **US3 (P3)**: Depends on Phase 4 (users.go) and T004 (consumers.go) — start after Phase 4

### File-Level Sequencing

| File | Phase 3 task | Phase 4 task | Phase 5 task |
|------|-------------|-------------|-------------|
| `internal/cmd/users.go` | — | T010 → T011 | T012 (after T011) |
| `internal/cmd/consumers.go` | T004 | — | T013 (after T004) |
| Other cmd files | T003, T005–T009 [P] | — | — |

### Within Each Phase

- Foundational (Phase 2) must complete before any user story phases
- Within Phase 4: T010 before T011 (same file)
- Within Phase 5: T013 after T004 (same file, different phases — phase ordering enforces this)

---

## Parallel Execution Examples

### Phase 3 — All Group A files in parallel

```
Parallel batch (7 tasks, 7 different files):
  T003: internal/cmd/clients.go    — 5 commands
  T004: internal/cmd/consumers.go  — 4 commands
  T005: internal/cmd/groups.go     — 1 command
  T006: internal/cmd/hrdc.go       — 1 command
  T007: internal/cmd/schedules.go  — 1 command
  T008: internal/cmd/signals.go    — 2 commands
  T009: internal/cmd/alarms.go     — 2 commands
```

### Phase 5 — Group C (two independent files)

```
Parallel batch (2 tasks, 2 different files):
  T012: internal/cmd/users.go      — users delete confirmation bypass
  T013: internal/cmd/consumers.go  — consumers password set stdin read
```

---

## Implementation Strategy

### MVP First (US1 Only)

1. Complete Phase 1: Create `internal/stdin` package (T001)
2. Complete Phase 2: Test the package (T002)
3. Complete Phase 3: All Group A file-record commands (T003–T009)
4. **STOP and VALIDATE**: Pipe a JSON array into each Group A command
5. Ship if sufficient — US2 and US3 can follow

### Incremental Delivery

1. T001 → T002 → Foundation ready
2. T003–T009 in parallel → US1 complete (16 commands) → validate and demo
3. T010–T011 → US2 complete (5 commands) → validate and demo
4. T012–T013 → US3 complete (2 commands) → validate and demo
5. T014–T016 → Polish and build verification

### Parallel Team Strategy

With multiple developers (after T001 + T002 complete):

- Developer A: T003 (clients), T004 (consumers), T007 (schedules)
- Developer B: T005 (groups), T006 (hrdc), T008 (signals), T009 (alarms)
- Developer C: T010–T011 (users Group B)

---

## Notes

- [P] tasks = different files, no blocking dependencies — safe to run concurrently
- [Story] label maps task to implementation group for traceability
- `internal/stdin` package is stateless — no global vars; all functions take `io.Reader` for testability
- `--file` and `--from-stdin` mutual exclusivity is a runtime check (not Cobra flag-level), per research.md Decision 3
- Group B commands: individual field flags are **silently ignored** (not errors) when `--from-stdin` is set
- Group C `users delete`: `--from-stdin` is a prompt-bypass signal only; stdin data is not read
- All error output goes to `os.Stderr`; only data output goes to `os.Stdout`
