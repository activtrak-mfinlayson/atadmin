# Implementation Plan: Input via Stdin (`--from-stdin`)

**Branch**: `007-stdin-input` | **Date**: 2026-05-20 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/007-stdin-input/spec.md`

## Summary

Add a `--from-stdin` flag to all mutating commands so that LLMs and scripts can pipe JSON payloads directly — bypassing shell escaping complexity and interactive confirmation prompts. The flag has three behavioral profiles depending on the command: (A) replace `--file` with stdin for bulk/record commands, (B) replace individual flags with a typed JSON payload for struct-based mutation commands, and (C) bypass interactive confirmation prompts for delete/password commands.

## Technical Context

**Language/Version**: Go 1.25.5 (module `go 1.25.5`)
**Primary Dependencies**: `spf13/cobra` v1, `encoding/json` (stdlib), `io` (stdlib) — no new external dependencies
**Storage**: N/A
**Testing**: `go test ./...`; table-driven tests; existing `httptest` patterns in `internal/api/`
**Target Platform**: Linux x86_64 / macOS arm64 (development)
**Project Type**: CLI tool
**Performance Goals**: No new latency requirements; stdin reading is synchronous and bounded by payload size
**Constraints**: JSON-only on stdin (CSV not supported via stdin, per spec); `--file` and `--from-stdin` are mutually exclusive
**Scale/Scope**: Internal operator tool; single invocation; payloads expected to be <10 MB

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Project constitution (`constitution.md`) has not been ratified. Governing constraints are drawn from `CLAUDE.md`.

**Effective Constraints from CLAUDE.md:**

| Gate | Status | Evidence |
|------|--------|---------|
| Separation of concerns — CLI in `internal/cmd/`, helpers in own packages | Pass | New `internal/stdin` package is transport-only; command logic stays in `internal/cmd/` |
| Machine readability — JSON input/output | Pass | `--from-stdin` is the machine-input counterpart to `--json` output |
| Graceful degradation — actionable error messages | Pass | All stdin errors use `--from-stdin: <reason>` prefix with clear remediation |
| No global state | Pass | `internal/stdin` functions are stateless; no package-level vars |
| Context as first arg on every API method | Pass | No new API methods; existing methods unchanged |
| Strong typing — no `map[string]interface{}` in typed commands | Pass | Group B uses concrete structs; Group A uses `map[string]any` (matching existing `bulk.ParseFile` pattern) |
| Test-first API clients | Pass | No new API client methods; `internal/stdin` gets its own unit tests |

All gates pass. No violations require justification.

**Post-Phase-1 re-check**: All gates remain satisfied. New `internal/stdin` package is self-contained and testable without httptest. Command changes are additive (new flag, new branch in RunE). No complexity violations.

## Project Structure

### Documentation (this feature)

```text
specs/007-stdin-input/
├── plan.md                     # This file
├── research.md                 # Phase 0 output
├── data-model.md               # Phase 1 output
├── contracts/
│   └── stdin-payloads.md       # Phase 1 output — JSON schemas per command
└── tasks.md                    # Phase 2 output (/speckit-tasks — not created here)
```

### Source Code Changes

```text
internal/
├── stdin/
│   ├── stdin.go                # NEW: ReadJSON[T] and ReadRecords helpers
│   └── stdin_test.go           # NEW: unit tests for both helpers
└── cmd/
    ├── users.go                # MODIFIED: users update, users delete, users bulk *
    ├── clients.go              # MODIFIED: merge-bulk, unmerge-bulk, alias-bulk, dnt bulk commands
    ├── consumers.go            # MODIFIED: delete-bulk, chrome-users import, create, update, password set
    ├── groups.go               # MODIFIED: groups members import
    ├── hrdc.go                 # MODIFIED: hrdc import
    ├── schedules.go            # MODIFIED: schedules create
    ├── signals.go              # MODIFIED: signals create, signals update
    └── alarms.go               # MODIFIED: alarms create, alarms update
```

### Commands Modified by Group

**Group A — file-record commands** (add `--from-stdin` as alternative to `--file`):

| File | Commands |
|---|---|
| `clients.go` | `clients merge-bulk`, `clients unmerge-bulk`, `clients alias-bulk`, `clients donottrack add-bulk`, `clients donottrack remove-bulk` |
| `consumers.go` | `consumers delete-bulk`, `consumers chrome-users import`, `consumers create`, `consumers update` |
| `groups.go` | `groups members import` |
| `hrdc.go` | `hrdc import` |
| `schedules.go` | `schedules create` |
| `signals.go` | `signals create`, `signals update` |
| `alarms.go` | `alarms create`, `alarms update` |

**Group B — typed-struct commands** (add `--from-stdin` as alternative to individual flags):

| File | Commands |
|---|---|
| `users.go` | `users update`, `users bulk start-tracking`, `users bulk stop-tracking`, `users bulk delete-entity`, `users bulk delete-data` |

**Group C — confirmation-bypass commands**:

| File | Commands |
|---|---|
| `users.go` | `users delete` |
| `consumers.go` | `consumers password set` |

## Complexity Tracking

No constitution violations. Section omitted.
