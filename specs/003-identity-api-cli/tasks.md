# Tasks: Identity API CLI Commands

**Input**: Design documents from `specs/003-identity-api-cli/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/commands.md ✓

**Tests**: Included per project convention (`CLAUDE.md` mandates `httptest`-based API tests for every new client method).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to ([US1]–[US6])

---

## Phase 1: Setup

No new project initialization needed — the Go module, directory structure, and toolchain are already established. Proceed directly to Foundational.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared types and error handling that every Identity command depends on. Must be complete before any user story implementation begins.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T001 Add Identity type family to `internal/api/models.go`: `IdentityField`, `IdentityAgent`, `IdentityGroupRef`, `Identity`, `UsersPage`, `IdentityListParams`, `UpdateUserRequest`, `BulkActionRequest`, `BulkEntityData`, `BulkActionResponse`, `BulkEntitySuccess`, `BulkEntityFailure`
- [x] T002 Extend `checkResponse` in `internal/api/helpers.go` with an explicit `http.StatusConflict` (409) case that returns an actionable error: `"conflict (409): the entity was modified concurrently. Re-fetch with 'atadmin users get <id>' and retry"`

**Checkpoint**: Foundation ready — user story phases can now begin.

---

## Phase 3: User Story 1 — List and Search Users (Priority: P1) 🎯 MVP

**Goal**: `atadmin users list` fetches and displays identity entities with filtering, searching, and pagination support. `--json` outputs raw `UsersPage` JSON.

**Independent Test**: Run `atadmin users list` against a mocked API server; verify table with ID, DISPLAY NAME, STATUS, GROUPS, TRACKED columns. Run `atadmin users list --filter tracked --json` and pipe to `jq`; verify valid JSON array.

### Implementation for User Story 1

- [x] T003 [P] [US1] Implement `ListUsers(ctx, IdentityListParams) (UsersPage, error)` in `internal/api/identity.go`. Maps `IdentityListParams` to `GET /identity/v1/entities` query parameters (search, searchType, filters, sort, sortDirection, count, cursor). Omits zero-value fields.
- [x] T004 [P] [US1] Write `TestListUsers` and `TestListUsers_Empty` in `internal/api/identity_test.go`. Use `httptest.NewServer` serving `{"results":[...],"cursor":"","totalCount":N}`. Assert method=GET, path, and decoded `UsersPage.Results` length and field values. Use `newTestClient` helper already present in the package.
- [x] T005 [US1] Create `internal/cmd/users.go` with `newUsersCmd(state *appState)` and `newUsersListCmd(state *appState)`. Table output: columns ID, DISPLAY NAME, STATUS, GROUPS, TRACKED (use `fieldValue()` helper for pointer fields). Support `--filter`, `--search`, `--search-type`, `--sort`, `--sort-dir`, `--limit`, `--cursor`, `--json` flags per `contracts/commands.md`.
- [x] T006 [US1] Register `newUsersCmd(&state)` in `newRootCmd()` in `internal/cmd/root.go` alongside existing resource commands.

**Checkpoint**: `atadmin users list` is fully functional and testable independently.

---

## Phase 4: User Story 2 — Inspect a Single User (Priority: P1)

**Goal**: `atadmin users get <id>` fetches and displays all fields of a single identity entity, including the revision number needed for mutations.

**Independent Test**: Run `atadmin users get 12345` against a mocked server; verify key-value output contains `id`, `revision`, `displayName`, `tracked`, `groups`, and `status` keys. Run with `--json`; verify output is valid `IdentityDetailsResponse` JSON.

**Note**: `GetUser` is also the foundation for revision auto-fetch in all mutating commands (US3, US4, US5). Complete this phase before implementing those.

### Implementation for User Story 2

- [x] T007 [P] [US2] Implement `GetUser(ctx, id int64) (*Identity, error)` in `internal/api/identity.go`. Calls `GET /identity/v1/entities/{id}`. Decodes the `IdentityDetailsResponse` JSON directly into `*Identity` (no wire wrapper needed for single-entity responses).
- [x] T008 [P] [US2] Write `TestGetUser` and `TestGetUser_NotFound` in `internal/api/identity_test.go`. Assert path `/identity/v1/entities/12345`, decoded `Identity.ID`, `Identity.Revision`, and `Identity.Tracked`. Assert 404 → error.
- [x] T009 [US2] Add `newUsersGetCmd(state *appState)` to `internal/cmd/users.go`. Accepts `<id>` positional arg (int64). Key-value output per contract: id, revision, displayName, firstName, lastName, emails, upns, groups, primaryGroup, tracked, status, timezone, agents, created, updated. Support `--json`.
- [x] T010 [US2] Add `users.AddCommand(newUsersGetCmd(state))` call in `newUsersCmd()` in `internal/cmd/users.go`.

**Checkpoint**: `atadmin users get <id>` is fully functional. Revision auto-fetch pattern is available for US3+.

---

## Phase 5: User Story 3 — Manage User Group Memberships (Priority: P2)

**Goal**: `atadmin users groups add <userId> <groupId>` and `atadmin users groups remove <userId> <groupId>` add or remove groups from an identity. Revision is auto-fetched via `GetUser`.

**Independent Test**: Run `atadmin users groups add 12345 42` against a mocked server that handles both the GET revision fetch and the POST groups call; verify the POST body contains `{"groupIds":[42],"revision":N}`.

**Dependency**: Requires T007 (`GetUser`) to be implemented first.

### Implementation for User Story 3

- [x] T011 [P] [US3] Implement `AddUserGroups(ctx, userID int64, groupIDs []int, revision int64) (*Identity, error)` in `internal/api/identity.go`. Calls `POST /identity/v1/entities/{id}/groups` with body `{"groupIds":[...],"revision":N}`. Returns updated `*Identity`.
- [x] T012 [P] [US3] Implement `RemoveUserGroups(ctx, userID int64, groupIDs []int, revision int64) (*Identity, error)` in `internal/api/identity.go`. Calls `DELETE /identity/v1/entities/{id}/groups` with body `{"groupIds":[...],"revision":N}`.
- [x] T013 [US3] Write `TestAddUserGroups` and `TestRemoveUserGroups` in `internal/api/identity_test.go`. Assert correct HTTP method, path, and request body JSON for each. Assert decoded response is a valid `*Identity`.
- [x] T014 [US3] Add `newUsersGroupsCmd`, `newUsersGroupsAddCmd`, `newUsersGroupsRemoveCmd` to `internal/cmd/users.go`. `groups add` and `groups remove` accept positional `<userId> <groupId>` or `<userId> --group-ids 1,2,3`. Both auto-fetch revision via `GetUser` when `--revision` is 0. Wire `groups` under `users` in `newUsersCmd`.

**Checkpoint**: `atadmin users groups add/remove` is fully functional. Group membership management is independently testable.

---

## Phase 6: User Story 4 — Update User Fields (Priority: P2)

**Goal**: `atadmin users update <id>` patches scalar fields (display name, first/last name, timezone, tracked) using `PATCH /identity/v1/entities/{id}/revision/{revision}`. `atadmin users delete <id>` deletes an entity using `DELETE /identity/v1/entities/{id}?revision=N`.

**Independent Test**: Run `atadmin users update 12345 --display-name "Alice Smith"` against a mocked server that handles the GET revision fetch and the PATCH; verify PATCH path contains `/revision/N` and body contains `{"displayName":{"value":"Alice Smith"}}`.

**Dependency**: Requires T007 (`GetUser`) to be implemented first.

### Implementation for User Story 4

- [x] T015 [P] [US4] Implement `UpdateUser(ctx, id, revision int64, req UpdateUserRequest) (*Identity, error)` in `internal/api/identity.go`. Calls `PATCH /identity/v1/entities/{id}/revision/{revision}`. Constructs a partial `FieldEditsDto`-shaped map from non-nil fields in `UpdateUserRequest`: string fields become `{"value":"..."}` objects; `Tracked` becomes a bare `"tracking": bool`.
- [x] T016 [P] [US4] Implement `DeleteUser(ctx, id, revision int64) error` in `internal/api/identity.go`. Calls `DELETE /identity/v1/entities/{id}?revision={revision}`. Expects 204 No Content. Uses `checkResponse`.
- [x] T017 [US4] Write `TestUpdateUser`, `TestUpdateUser_Conflict`, `TestDeleteUser` in `internal/api/identity_test.go`. For update: assert PATCH method, path includes `/revision/3`, body has correct field shape. For conflict: mock server returns 409, assert error message contains "Re-fetch". For delete: assert DELETE method and query param `revision`.
- [x] T018 [US4] Add `newUsersUpdateCmd(state *appState)` to `internal/cmd/users.go`. Accepts `<id>` arg plus `--display-name`, `--first-name`, `--last-name`, `--timezone`, `--tracked`, `--revision`, `--json` flags. Validates at least one update flag is provided. Auto-fetches revision when `--revision=0`. On success: prints `Updated user <id>`. Wire into `newUsersCmd`.
- [x] T019 [US4] Add `newUsersDeleteCmd(state *appState)` to `internal/cmd/users.go`. Accepts `<id>` arg plus `--revision`, `--yes` flags. In TTY mode without `--yes`: prompts `Delete user <id>? [y/N]` using `tty.IsTerminal`. In non-TTY mode without `--yes`: fails with error. On success: prints `Deleted user <id>`. Wire into `newUsersCmd`.

**Checkpoint**: `atadmin users update` and `atadmin users delete` are fully functional.

---

## Phase 7: User Story 5 — Bulk Actions on Multiple Users (Priority: P3)

**Goal**: `atadmin users bulk <action> --ids 1,2,3` applies start-tracking, stop-tracking, delete-entity, or delete-data to multiple identities in one API call. Revisions are pre-fetched concurrently.

**Independent Test**: Run `atadmin users bulk stop-tracking --ids 12345,12346` against a mocked server that handles two concurrent GET requests and one POST to `/identity/v1/entities/bulk`; verify bulk request body has correct `actions` and `data` arrays, and output table shows per-entity success/failure.

**Dependency**: Requires T007 (`GetUser`) to be implemented first.

### Implementation for User Story 5

- [x] T020 [P] [US5] Implement `BulkAction(ctx, req BulkActionRequest) (*BulkActionResponse, error)` in `internal/api/identity.go`. Calls `POST /identity/v1/entities/bulk` with JSON body. Decodes `BulkActionResponse`. Uses `checkResponse`.
- [x] T021 [P] [US5] Write `TestBulkAction_AllSuccess` and `TestBulkAction_PartialFailure` in `internal/api/identity_test.go`. Assert POST method, path, and request body contains correct `actions` and `data` arrays. Assert decoded response `Successful` and `Failures` slices.
- [x] T022 [US5] Add `newUsersBulkCmd(state *appState)` to `internal/cmd/users.go` with four subcommands: `start-tracking`, `stop-tracking`, `delete-entity`, `delete-data`. Each accepts `--ids` (comma-separated int64 list) and `--json`. Implements concurrent revision pre-fetch (up to 10 goroutines via `sync.WaitGroup` + buffered channel). Prints per-entity success/failure table. Exits with code 1 if any failures. Wire `bulk` under `users` in `newUsersCmd`.

**Checkpoint**: `atadmin users bulk` is fully functional with partial-failure reporting.

---

## Phase 8: User Story 6 — List and Inspect Agent Entities (Priority: P3)

**Goal**: `atadmin agents list` fetches and displays agent (device) entities across the account using the same `UsersPage` shape as users.

**Independent Test**: Run `atadmin agents list` against a mocked server at `/identity/v1/agents`; verify table with USER ID, USERNAME, DOMAIN, ALIAS, LICENSE, LAST LOG columns. Run `atadmin agents list --json`; verify valid JSON with `results` array containing nested `agents` data.

**Dependency**: Foundational types (T001) required. Independent of US2–US5.

### Implementation for User Story 6

- [x] T023 [P] [US6] Implement `ListAgents(ctx, IdentityListParams) (UsersPage, error)` in `internal/api/identity.go`. Same parameter handling as `ListUsers` but targets `GET /identity/v1/agents`.
- [x] T024 [P] [US6] Write `TestListAgents` and `TestListAgents_Empty` in `internal/api/identity_test.go`. Assert path `/identity/v1/agents`. Same response shape as `TestListUsers`.
- [x] T025 [US6] Create `internal/cmd/agents.go` with `newAgentsCmd(state *appState)` and `newAgentsListCmd(state *appState)`. Table output: USER ID, USERNAME, DOMAIN, ALIAS, LICENSE, LAST LOG (extract from `Identity.Agents[0]` or show entity-level data when no agents). Support same flags as `users list`.
- [x] T026 [US6] Register `newAgentsCmd(&state)` in `newRootCmd()` in `internal/cmd/root.go`.

**Checkpoint**: `atadmin agents list` is fully functional and independently testable.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Validation, linting, and correctness pass across all stories.

- [x] T027 Run `go test ./...` from repo root; fix any test failures in `internal/api/identity_test.go` or `internal/cmd/` before proceeding
- [x] T028 [P] Run `golangci-lint run`; fix all reported issues (common: unused variables, unhandled `fmt.Fprint*` returns, errcheck violations)
- [x] T029 [P] Validate `quickstart.md` scenarios against the running CLI — smoke-test each command example in `specs/003-identity-api-cli/quickstart.md` with `--help` at minimum; flag any commands that error or produce wrong output
- [x] T030 Verify `atadmin --help` and `atadmin users --help` list all new subcommands with descriptions

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: No dependencies — start immediately
- **US1 (Phase 3)**: Depends on Foundational (T001, T002)
- **US2 (Phase 4)**: Depends on Foundational (T001, T002)
- **US3 (Phase 5)**: Depends on Foundational AND US2 (T007 `GetUser` required for revision auto-fetch)
- **US4 (Phase 6)**: Depends on Foundational AND US2 (T007 `GetUser` required)
- **US5 (Phase 7)**: Depends on Foundational AND US2 (T007 `GetUser` required)
- **US6 (Phase 8)**: Depends on Foundational only — independent of US2–US5
- **Polish (Phase 9)**: Depends on all desired stories being complete

### User Story Dependencies

```
Foundational (T001, T002)
    ├── US1 (T003–T006)    ← independent
    ├── US2 (T007–T010)    ← independent; T007 (GetUser) gates US3/US4/US5
    │     └── US3 (T011–T014)
    │     └── US4 (T015–T019)
    │     └── US5 (T020–T022)
    └── US6 (T023–T026)    ← independent of US2–US5
```

### Within Each User Story

- API method implementation → API test → cmd implementation (sequential within story)
- API method + API test can be written in parallel (different concerns, same file — coordinate)

### Parallel Opportunities

| When | What can run in parallel |
|---|---|
| After T001+T002 | T003+T004 (US1 API), T007+T008 (US2 API), T023+T024 (US6 API) |
| After T007 | T011+T012 (US3), T015+T016 (US4), T020+T021 (US5) can all start |
| After all stories | T027+T028+T029 (Polish) |

---

## Parallel Example: Foundational → US1 + US2 + US6 simultaneously

```text
# Once T001 and T002 are done, launch in parallel:
Task T003: Implement ListUsers() in internal/api/identity.go
Task T004: Write TestListUsers in internal/api/identity_test.go
Task T007: Implement GetUser() in internal/api/identity.go
Task T008: Write TestGetUser in internal/api/identity_test.go
Task T023: Implement ListAgents() in internal/api/identity.go
Task T024: Write TestListAgents in internal/api/identity_test.go
```

Note: T003, T007, and T023 all write to `internal/api/identity.go` — coordinate to avoid merge conflicts (e.g., implement sequentially or split into separate files per concern).

---

## Implementation Strategy

### MVP First (User Story 1 Only — Read-Only Listing)

1. Complete Phase 2: Foundational (T001, T002)
2. Complete Phase 3: US1 — `atadmin users list` (T003–T006)
3. **STOP and VALIDATE**: `go test ./...` passes; `atadmin users list --json` works against live API
4. Demo or deploy

### Recommended Incremental Delivery

1. Foundation (T001–T002) → 2 tasks
2. US1 `users list` (T003–T006) → **MVP, demo-able**
3. US2 `users get` (T007–T010) → enables all mutation stories
4. US3 `users groups` (T011–T014) → group management
5. US4 `users update/delete` (T015–T019) → full single-entity lifecycle
6. US5 `users bulk` (T020–T022) → batch operations
7. US6 `agents list` (T023–T026) → device inventory
8. Polish (T027–T030)

### Full Parallel Strategy (Solo Developer)

Because `identity.go` is a single file that US1, US2, and US6 all write to, the recommended approach for a solo developer is sequential within each file but parallel across different files (e.g., write API method then immediately write its test before moving to the next method).

---

## Notes

- All new API methods go in `internal/api/identity.go`; all tests go in `internal/api/identity_test.go`
- Use the existing `newTestClient(t, server.URL, "tok")` helper for all API tests
- Follow the `unexported wire struct → public model` pattern from `groups.go` and `signals.go` for any wire format mismatches discovered at implementation time
- Commit after each phase checkpoint (or use `/speckit-git-commit`)
- If a 409 test in T017 fails, verify T002 (`checkResponse` 409 case) was applied first
- The `fieldValue()` helper from `data-model.md` should be added to `internal/api/identity.go` as an unexported function (not exported from the package)
