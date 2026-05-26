# Implementation Plan: Safe Exploration (--dry-run)

**Branch**: `008-dry-run` | **Date**: 2026-05-20 | **Spec**: [spec.md](spec.md)  
**Input**: `specs/008-dry-run/spec.md`

## Summary

Add a persistent `--dry-run` flag to the root command that short-circuits all mutating HTTP requests (`POST`, `PUT`, `PATCH`, `DELETE`) inside `Client.doRequest()`, printing a structured JSON preview instead of executing the operation. GET requests are unaffected. The flag is implemented as a field on `api.Client`, keeping all interception logic in one place.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: cobra v1.10.2, viper v1.21.0, standard `net/http`  
**Storage**: N/A (CLI tool)  
**Testing**: `go test` + `net/http/httptest`  
**Target Platform**: Linux/macOS/Windows CLI  
**Project Type**: CLI  
**Performance Goals**: N/A  
**Constraints**: Zero HTTP requests must be made when `--dry-run` is set for mutating calls  
**Scale/Scope**: ~50 mutating commands across 10+ resource types

## Constitution Check

*GATE: Constitution file is an unfilled template — no gates defined. Proceeding on CLAUDE.md project principles.*

Project principles from `CLAUDE.md` that apply:
- **Separation of Concerns**: dry-run logic lives in `internal/api/`, not scattered across `internal/cmd/`
- **Machine Readability**: dry-run output is JSON (consistent with `--json` flag)
- **Test-First API Clients**: `httptest`-based tests verify no HTTP calls are made
- **Standard Streams**: dry-run JSON output goes to `stdout`; diagnostic messages go to `stderr`

No violations. No complexity justification needed.

## Project Structure

### Documentation (this feature)

```text
specs/008-dry-run/
├── plan.md              ← this file
├── research.md          ← Phase 0 output
├── data-model.md        ← Phase 1 output
├── quickstart.md        ← Phase 1 output
├── contracts/
│   └── dry-run-output.md  ← CLI output contract
└── tasks.md             ← Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
internal/
├── api/
│   ├── client.go        ← add DryRun bool + Out io.Writer fields; update NewClient()
│   └── helpers.go       ← add dry-run short-circuit in doRequest()
└── cmd/
    └── root.go          ← add --dry-run persistent flag; pass to NewClient()

tests (existing pattern):
internal/api/*_test.go   ← add dry-run httptest coverage for each resource
internal/cmd/*_test.go   ← add CLI-level dry-run output assertions
```

**Structure Decision**: Single project (Option 1). Dry-run adds two small changes to existing files — no new packages needed.

## Complexity Tracking

No constitution violations. No entries required.
