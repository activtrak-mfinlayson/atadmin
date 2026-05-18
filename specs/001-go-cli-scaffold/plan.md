# Implementation Plan: Go CLI Foundation & CI/CD Setup

**Branch**: `001-go-cli-scaffold` | **Date**: 2026-05-18 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/001-go-cli-scaffold/spec.md`

## Summary

Scaffold the `atadmin` Go CLI with a working root command, `--version` flag, `auth` subcommand placeholder, and Viper-based configuration loading. Establish a GitHub Actions CI pipeline that builds, lints, tests, and security-scans on every push and PR, and uploads a binary artifact on successful main-branch runs.

## Technical Context

**Language/Version**: Go 1.23  
**Primary Dependencies**: `spf13/cobra` v1 (command routing), `spf13/viper` v2 (configuration), `golang.org/x/term` (TTY detection), `golangci-lint` v1.x (linting gate, external CI tool), `govulncheck` (security scan via `go run`)  
**Storage**: N/A — no persistent storage in scaffold  
**Testing**: `go test ./...`; table-driven tests; `net/http/httptest` for future API client tests  
**Target Platform**: Linux x86_64 (CI runner, primary runtime), macOS arm64/amd64 (development)  
**Project Type**: CLI tool  
**Performance Goals**: CI build-lint-test-scan cycle completes in under 5 minutes (SC-002)  
**Constraints**: Single binary output (`atadmin`); no multi-platform cross-compilation in v1  
**Scale/Scope**: Internal operator/admin tool; single-user per invocation; no concurrency requirements at scaffold stage  

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The project constitution (`constitution.md`) has not yet been ratified. Governing constraints are drawn from `CLAUDE.md`.

**Effective Constraints from CLAUDE.md**:

| Gate | Status | Evidence |
|------|--------|---------|
| Separation of concerns — CLI in `internal/cmd/`, API in `internal/api/` | Pass | Project structure enforces this; directories are distinct packages |
| Machine readability — `--json` flag on all data-fetching commands | Pass | Not applicable to scaffold root command; required for all future list/get commands |
| Test-first API clients — use `httptest.NewServer` | Pass | Enforced by testing guidelines; scaffold includes example `httptest` pattern |
| Graceful degradation — actionable error messages | Pass | FR-001, FR-003, FR-004; acceptance scenarios 1.4 and 2.2 verify error paths |
| No global HTTP client | Pass | Client struct instantiated per-command; no `http.DefaultClient` usage |
| `context.Context` as first arg on every API method | Pass | Code convention; enforced by linter config (`contextcheck` rule) |
| Auth injection via `http.RoundTripper` | Pass | Noted in contracts; scaffold includes the `RoundTripper` stub |
| Strong typing — no `map[string]interface{}` | Pass | Enforced by linter (`gocritic`, `staticcheck`) |
| Linting is a required CI gate | Pass | FR-013; `golangci-lint-action` runs before tests; failure blocks merge |

All gates pass. No violations require justification.

**Post-Phase-1 re-check**: All gates remain satisfied. Project structure (see below) cleanly separates concerns; contracts define the `RoundTripper` interface; no complexity violations introduced.

## Project Structure

### Documentation (this feature)

```text
specs/001-go-cli-scaffold/
├── plan.md              # This file (/speckit.plan output)
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── commands.md
└── tasks.md             # Phase 2 output (/speckit.tasks — not created here)
```

### Source Code (repository root)

```text
cmd/
└── atadmin/
    └── main.go                  # Entry point — calls internal/cmd.Execute()

internal/
├── cmd/
│   ├── root.go                  # Root cobra command, --version flag, config init
│   └── auth.go                  # auth subcommand group (placeholder)
├── api/
│   ├── client.go                # Client struct + authRoundTripper
│   └── models.go                # Shared request/response types
└── config/
    └── config.go                # Viper init, Config struct, env-var bindings

.github/
└── workflows/
    └── ci.yml                   # Build → Lint → Test → Security scan → Artifact

bin/                             # Build output (gitignored)
go.mod                           # go 1.23, module github.com/activtrak-mfinlayson/atadmin
go.sum
.golangci.yml                    # Linter configuration
.gitignore
```

## Complexity Tracking

No constitution violations. Section omitted.
