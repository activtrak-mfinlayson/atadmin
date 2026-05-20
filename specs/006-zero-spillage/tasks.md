---

description: "Task list for Feature 006: Zero-Spillage Rule"
---

# Tasks: Zero-Spillage Rule

**Input**: Design documents from `/specs/006-zero-spillage/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/json-error.schema.json ✓, quickstart.md ✓

**Organization**: Tasks are grouped by user story.
- **US1 (P1) — Stdout Purity**: Reclassify ~21 mutation confirmation call-sites from `OutOrStdout` → `ErrOrStderr`
- **US2 (P2) — Structured JSON Errors**: Add output helpers and wire `Execute()` to emit `{"error": ..., "suggestion": ...}` to stdout when `--json` is active

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1 or US2)
- Exact file paths are included in all descriptions

---

## Phase 1: Setup (Audit)

**Purpose**: Establish the complete list of non-data stdout writes across all cmd files before making changes.

- [ ] T001 Audit all files in `internal/cmd/` not covered by data-model.md (agents.go, alarms.go, apikeys.go, auditlog.go, auth.go, clients.go, devices.go, hrdc.go, notifications.go, schedules.go) for any `fmt.Fprint*` calls sending non-data output to `cmd.OutOrStdout()` and document any additional sites

**Checkpoint**: Audit complete — scope of reclassification confirmed (≥21 known sites + any newly found)

---

## Phase 2: Foundational (New Output Helpers)

**Purpose**: Core output utilities that US2 depends on and that US1 call-sites implicitly rely on for consistent error routing. Must be complete before wiring `Execute()`.

- [ ] T002 Add `JSONError` struct, `WriteError()`, `DetectJSONMode()`, and `SuggestionFor()` functions to `internal/output/output.go` per data-model.md spec (including the suggestion-string mapping table from research.md §3)

**Checkpoint**: `output.WriteError` and `output.DetectJSONMode` are callable — US1 and US2 implementation can now proceed in parallel

---

## Phase 3: User Story 1 — Stdout Purity (Priority: P1) 🎯 MVP

**Goal**: Every mutation confirmation message ("Updated user 123", "Deleted N groups", etc.) goes to stderr unconditionally, so stdout is clean for both piped and JSON-mode callers.

**Independent Test**:
```sh
# mutation command — stderr suppressed; stdout must be empty
atadmin users update 123 --display-name "Test" 2>/dev/null
# no output to stdout; exit 0
```

### Implementation for User Story 1

- [ ] T003 [P] [US1] Reclassify 6 diagnostic messages from `cmd.OutOrStdout()` to `cmd.ErrOrStderr()` in `internal/cmd/users.go` (~lines 229 "Updated user", 269 "Delete user? [y/N]", 274 "Aborted.", 291 "Deleted user", 364 "Added group(s)", 419 "Removed group(s)")
- [ ] T004 [P] [US1] Reclassify 10 diagnostic messages from `cmd.OutOrStdout()` to `cmd.ErrOrStderr()` in `internal/cmd/groups.go` (~lines 218 "Renamed group", 249 "Deleted N groups", 251 "deleted", 371 "Added member", 405 "Removed member", 469 "Imported N records", 525 "Added N clients", 561 "Removed N clients", 617 "Added N devices", 653 "Removed N devices")
- [ ] T005 [P] [US1] Reclassify 5 diagnostic messages from `cmd.OutOrStdout()` to `cmd.ErrOrStderr()` in `internal/cmd/consumers.go` (~lines 134 "Created N consumers", 165 "Updated N consumers", 196 "Deleted N consumers", 198 "deleted", 230 "Deleted N consumers")
- [ ] T006 [US1] Fix any additional stray stdout writes found during T001 audit in remaining `internal/cmd/` files

**Checkpoint**: All mutation commands write zero non-data output to stdout; stderr carries all confirmations

---

## Phase 4: User Story 2 — Structured JSON Errors (Priority: P2)

**Goal**: Any `atadmin <cmd> --json` failure emits `{"error": "...", "suggestion": "..."}` to stdout (exit 1) instead of a plain-text error to stderr, enabling agents to parse and self-correct.

**Independent Test**:
```sh
# with an invalid/expired token
result=$(atadmin users list --json 2>/dev/null)
echo "$result" | jq -e '.error'   # must succeed
echo "$result" | jq -e '.suggestion'  # must be non-empty for auth errors
```

### Implementation for User Story 2

- [ ] T007 [US2] Modify `Execute()` in `internal/cmd/root.go` to detect JSON mode with `output.DetectJSONMode(os.Args[1:])` before calling `root.Execute()`, then route errors through `output.WriteError()` targeting stdout when `asJSON` is true — replacing the current plain `fmt.Fprintf(root.ErrOrStderr(), ...)` call per data-model.md §Modified Call-site

**Checkpoint**: `atadmin <any-cmd> --json` with a bad token outputs a valid JSON error object; non-JSON mode is unchanged

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Verify correctness across both stories and validate the quickstart.md acceptance criteria.

- [ ] T008 [P] Run `go test ./...` and confirm all existing tests pass with no regressions from the stdout→stderr reclassification
- [ ] T009 [P] Manually verify all four quickstart.md scenarios: (1) read command stdout purity, (2) mutation command stdout purity, (3) structured error in JSON mode, (4) non-JSON mode unchanged

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately; confirms final scope
- **Foundational (Phase 2)**: No dependencies — start immediately; output helpers needed before T007
- **User Story 1 (Phase 3)**: Depends only on Phase 1 (scope confirmed); T003/T004/T005 are parallel with each other; T006 follows T001
- **User Story 2 (Phase 4)**: Depends on Phase 2 (T002 must be complete before T007)
- **Polish (Phase 5)**: Depends on all prior phases complete

### User Story Dependencies

- **US1 tasks (T003–T006)**: Independent of US2; can proceed after T001
- **US2 task (T007)**: Depends on T002 (output helpers); independent of US1 changes

### Within Each Phase

- T003, T004, T005 are fully parallel (different files)
- T006 follows T001 (its scope depends on audit findings)
- T007 follows T002 (requires the new helper signatures)
- T008 and T009 can run in parallel after all implementation tasks complete

### Parallel Opportunities

```bash
# Phase 1 and 2 can start simultaneously:
Task: T001 (audit remaining cmd files)
Task: T002 (add output helpers)

# Once T001 completes, US1 tasks run in parallel:
Task: T003 (users.go reclassification)
Task: T004 (groups.go reclassification)
Task: T005 (consumers.go reclassification)

# Once T002 completes, US2 task runs:
Task: T007 (wire Execute())
```

---

## Implementation Strategy

### MVP First (US1 Only)

1. Complete Phase 1: Audit
2. Complete Phase 2: Output helpers (required for correctness even in US1 path)
3. Complete Phase 3: US1 reclassification (T003–T006)
4. **STOP and VALIDATE**: Run `go test ./...`; test mutation commands manually
5. US1 delivers stdout purity — agents can now pipe mutation commands without corruption

### Incremental Delivery

1. Phase 1 + Phase 2 → audit complete, helpers ready
2. Phase 3 (US1) → mutation confirmations move to stderr; testable independently
3. Phase 4 (US2) → structured JSON errors enabled; testable independently
4. Phase 5 → full regression + quickstart validation

---

## Notes

- **Do not** reclassify data outputs (scalar IDs, boolean values, URL strings, "OK") — these stay on `OutOrStdout` per data-model.md
- Interactive prompts (`"Delete user %d? [y/N] "`) are diagnostic and move to `ErrOrStderr` per research.md
- The `"Run 'atadmin --help' for usage."` suffix in the current `Execute()` error is dropped in JSON mode (not machine-friendly); the suggestion field replaces it
- `SuggestionFor()` inspects `err.Error()` for substrings — no typed unwrapping required
