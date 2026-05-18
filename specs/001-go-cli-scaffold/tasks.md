# Tasks: Go CLI Foundation & CI/CD Setup

**Input**: Design documents from `specs/001-go-cli-scaffold/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓, quickstart.md ✓

**Tests**: Test tasks are included as the implementation of User Story 2 (Developer Runs the Test Suite). They are not structured as TDD precursors — the spec does not require test-first order.

**Binary name**: `atadmin` | **Module**: `github.com/activtrak-mfinlayson/atadmin` | **Go**: 1.23

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize the repository structure and tooling. No user story work can begin until this is complete.

- [x] T001 Create all source directories: `cmd/atadmin/`, `internal/cmd/`, `internal/api/`, `internal/config/`, `.github/workflows/`, `bin/`
- [x] T002 Create `go.mod` with `module github.com/activtrak-mfinlayson/atadmin` and `go 1.23`
- [x] T003 [P] Create `.gitignore` ignoring `bin/`, `*.out`, `coverage.out`, `.DS_Store`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure shared by all user stories. All of Phase 1 must complete first.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T004 Add `spf13/cobra`, `spf13/viper`, and `golang.org/x/term` to `go.mod` and `go.sum` by running `go get github.com/spf13/cobra@latest github.com/spf13/viper@latest golang.org/x/term@latest`
- [x] T005 [P] Create `internal/config/config.go` — define `Config` struct with fields `Token string`, `BaseURL string`, `Format string`, `Timeout time.Duration`; implement `Load()` that initializes Viper with config file at `~/.config/atadmin/config.yaml` (XDG path) plus `~/.atadmin/config.yaml` fallback, binds env vars `ATADMIN_TOKEN`, `ATADMIN_BASE_URL`, `ATADMIN_FORMAT`, `ATADMIN_TIMEOUT`, and returns a populated `Config`
- [x] T006 [P] Create `internal/api/models.go` — package `api`; add placeholder comment block describing future request/response structs; define an empty `ErrorResponse` struct with `Message string \`json:"message"\`` for use in error parsing
- [x] T007 Create `internal/api/client.go` — define `Client` struct with fields `BaseURL *url.URL`, `HTTPClient *http.Client`, `UserAgent string`; define `authRoundTripper` struct with `token string` and `inner http.RoundTripper`; implement `authRoundTripper.RoundTrip(r *http.Request)` that clones the request, sets `Authorization: Bearer <token>`, delegates to `inner.RoundTrip`; implement `NewClient(baseURL, token, version string) (*Client, error)` that constructs the client with `authRoundTripper` as transport

**Checkpoint**: Core types and packages are in place. Run `go build ./...` — must compile with no errors before proceeding.

---

## Phase 3: User Story 1 — Developer Runs the CLI Locally (Priority: P1) 🎯 MVP

**Goal**: A developer can build `atadmin` and invoke it with `--help`, `--version`, and see proper error output for unknown commands.

**Independent Test**: Build the binary and verify:
1. `./bin/atadmin --help` → prints help with "atadmin" in the header and lists subcommands
2. `./bin/atadmin --version` → prints `atadmin version 0.1.0`
3. `./bin/atadmin unknowncmd` → exits non-zero, prints error to stderr with "Run 'atadmin --help'" hint
4. `./bin/atadmin auth --help` → prints auth subcommand help

- [x] T008 [US1] Create `internal/cmd/root.go` — define `rootCmd` as a `*cobra.Command` with `Use: "atadmin"`, `Short: "ActivTrak admin CLI"`, and `Long` description; add `--version` flag that prints `atadmin version 0.1.0` to stdout and exits 0; add persistent flags `--format` (default `"table"`, env `ATADMIN_FORMAT`), `--token` (env `ATADMIN_TOKEN`), `--base-url` (env `ATADMIN_BASE_URL`); implement `Execute() error` that is the public entry point
- [x] T009 [US1] Create `internal/cmd/auth.go` — define `authCmd` as a `*cobra.Command` with `Use: "auth"`, `Short: "Manage authentication"`; define `authLoginCmd` as a stub with `Use: "login"`, `Short: "Authenticate with the API"` that prints `"Not yet implemented. Run 'atadmin auth --help' for usage."` to stderr and returns exit code 1; register `authLoginCmd` on `authCmd` and register `authCmd` on `rootCmd` in an `init()` function
- [x] T010 [US1] Create `cmd/atadmin/main.go` — `package main`; call `internal/cmd.Execute()`; if error is non-nil, print to stderr and `os.Exit(1)`
- [x] T011 [US1] Build the binary: `go build -o bin/atadmin ./cmd/atadmin` — must produce `bin/atadmin` with no compilation errors
- [x] T012 [US1] Verify all four acceptance scenarios from the Independent Test above pass manually; confirm stderr vs stdout separation by running `./bin/atadmin unknowncmd 2>/dev/null` (stdout empty) and `./bin/atadmin unknowncmd 2>&1 1>/dev/null` (error shown)

**Checkpoint**: `./bin/atadmin --help` works. User Story 1 is independently functional. MVP deliverable.

---

## Phase 4: User Story 2 — Developer Runs the Test Suite (Priority: P2)

**Goal**: `go test ./...` passes and produces a coverage report. Example tests cover config loading, the auth RoundTripper, and root command output.

**Independent Test**: `go test ./...` exits 0 with all tests passing; `go test -coverprofile=coverage.out ./...` produces coverage output.

- [x] T013 [P] [US2] Create `internal/config/config_test.go` — table-driven tests: (1) verify `Load()` returns defaults when no config file exists; (2) verify env var `ATADMIN_TOKEN=testtoken` sets `Config.Token` to `"testtoken"` (use `t.Setenv`); (3) verify `ATADMIN_FORMAT=json` overrides default format
- [x] T014 [P] [US2] Create `internal/api/client_test.go` — use `httptest.NewServer` to start a local server that captures request headers; create a `NewClient` pointing to the test server URL with token `"mytoken"`; make a request and assert the server received `Authorization: Bearer mytoken`; assert the `User-Agent` header contains `"atadmin/"`
- [x] T015 [P] [US2] Create `internal/cmd/root_test.go` — test `--version` flag: set `cmd.SetOut(buf)`, execute `atadmin --version`, assert buf contains `"atadmin version 0.1.0"`; test unknown command: set `cmd.SetErr(errBuf)`, execute `atadmin unknowncmd`, assert exit code non-zero and errBuf contains `"atadmin --help"`
- [x] T016 [US2] Run `go test ./...` — all tests must pass; run `go test -coverprofile=coverage.out ./...` — verify coverage output is generated; fix any test failures before marking complete

**Checkpoint**: `go test ./...` exits 0. User Story 2 independently functional.

---

## Phase 5: User Story 3 — CI Pipeline Validates Pull Requests (Priority: P3)

**Goal**: GitHub Actions CI runs on every push and PR, with lint (blocking), test, security scan, and artifact upload on main.

**Independent Test**: Push to GitHub → CI triggers → all stages pass → green status check on PR. Introduce a linting error → lint stage fails → red check.

- [x] T017 [US3] Create `.golangci.yml` — set `run.timeout: 5m`; under `linters.enable` list: `errcheck`, `gocritic`, `gosimple`, `staticcheck`, `stylecheck`, `revive`, `nolintlint`, `exhaustive`; set `linters-settings.revive.rules` to include `exported` rule requiring doc comments on exported symbols
- [x] T018 [US3] Run `golangci-lint run` locally — resolve any linting violations in the scaffold code before creating the CI workflow; all linters must pass with zero issues
- [x] T019 [US3] Create `.github/workflows/ci.yml` with the following structure:
  ```
  name: CI
  on:
    push:
      branches: ["**"]
    pull_request:
      branches: [main]
  jobs:
    ci:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v4
        - uses: actions/setup-go@v5
          with: { go-version: '1.23' }
        - name: Build
          run: go build -o bin/atadmin ./cmd/atadmin
        - uses: golangci/golangci-lint-action@v6
          with: { version: latest }
        - name: Test
          run: go test ./...
        - name: Security scan
          run: go run golang.org/x/vuln/cmd/govulncheck@latest ./...
        - uses: actions/upload-artifact@v4
          if: github.ref == 'refs/heads/main' && success()
          with:
            name: atadmin-linux-amd64
            path: bin/atadmin
  ```
- [ ] T020 [US3] Create the GitHub remote repository (if not already done) and push the branch; verify CI triggers and all stages pass; confirm artifact appears under the Actions run summary on a main-branch push
- [ ] T021 [US3] Verify PR gate: open a PR against main; confirm the CI status check appears; deliberately introduce and then revert a linting error to confirm the lint stage blocks the PR when it fails

**Checkpoint**: CI pipeline fully operational. All three user stories independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final cleanup, documentation, and acceptance validation.

- [x] T022 [P] Create root-level `README.md` with: project description ("ActivTrak admin CLI"), prerequisites (Go 1.23, golangci-lint), quick build/test instructions (`go build -o bin/atadmin ./cmd/atadmin`, `go test ./...`), link to `specs/001-go-cli-scaffold/quickstart.md` for full guide
- [x] T023 [P] Update `CLAUDE.md` (symlink target `AGENTS.md`) — replace the generic build command `go build -o bin/app ./cmd/appname` with `go build -o bin/atadmin ./cmd/atadmin` to reflect the actual binary name
- [x] T024 Run the full acceptance check from `quickstart.md`: build → `atadmin --help` → `atadmin --version` → `go test ./...` → `golangci-lint run` → `govulncheck ./...` → all pass with zero errors; record any failures and fix before marking complete

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 completion
- **Phase 3 (US1 — Local Build)**: Depends on Phase 2 — ⚠️ BLOCKS US2 and US3
- **Phase 4 (US2 — Test Suite)**: Depends on Phase 3 (needs implementation to test)
- **Phase 5 (US3 — CI Pipeline)**: Depends on Phase 4 (needs passing tests for CI to validate)
- **Phase 6 (Polish)**: Depends on all phases complete

### User Story Dependencies

- **US1 (P1)**: Independent after Phase 2. MVP deliverable on its own.
- **US2 (P2)**: Depends on US1 implementation existing; test files are written against it.
- **US3 (P3)**: Depends on US2 (CI must run and pass the test suite).

### Within Each Phase — Parallel Opportunities

- T003, T005, T006 can run in parallel (different files, no shared state)
- T013, T014, T015 can run in parallel (separate test files, no inter-dependency)
- T022, T023 can run in parallel (separate files)

---

## Parallel Execution Example: Phase 2 Foundational

```
# Start these concurrently:
Task T005: internal/config/config.go
Task T006: internal/api/models.go

# After both complete:
Task T007: internal/api/client.go (uses Config + models)
```

## Parallel Execution Example: Phase 4 (US2 Test Suite)

```
# Start these concurrently:
Task T013: internal/config/config_test.go
Task T014: internal/api/client_test.go
Task T015: internal/cmd/root_test.go

# After all complete:
Task T016: go test ./... (validates all three)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (`go build ./...` must pass)
3. Complete Phase 3: User Story 1 (build + verify --help, --version, error output)
4. **STOP AND VALIDATE**: `./bin/atadmin --help` works — MVP deliverable
5. Proceed to Phase 4 when ready

### Incremental Delivery

1. Phase 1 + 2 → Foundation ready
2. Phase 3 (US1) → Working binary ← **Demo here**
3. Phase 4 (US2) → Test suite passes ← **Demo here**
4. Phase 5 (US3) → CI green ← **Demo here**
5. Phase 6 → Polished and documented

### Single-Developer Sequential Order

T001 → T002 → T003 → T004 → T005 → T006 → T007 → T008 → T009 → T010 → T011 → T012 → T013 → T014 → T015 → T016 → T017 → T018 → T019 → T020 → T021 → T022 → T023 → T024

---

## Notes

- `[P]` = different files, no incomplete dependencies — safe to run concurrently
- `[USn]` maps each task to its user story for traceability
- Each phase checkpoint is a demo-able increment — stop and validate before proceeding
- Never write API credentials or tokens in code, logs, or test output
- Commit after each phase checkpoint at minimum; commit after each task when possible
