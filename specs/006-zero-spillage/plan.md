# Implementation Plan: Zero-Spillage Rule

**Branch**: `006-zero-spillage` | **Date**: 2026-05-20 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/006-zero-spillage/spec.md`

## Summary

Guarantee that `atadmin`'s stdout emits *only* valid JSON when `--json` is active, and that errors in JSON mode are emitted as a structured `{"error": "...", "suggestion": "..."}` object to stdout. The fix has two parts: (1) move mutation confirmation messages ("Updated user 123") from `cmd.OutOrStdout()` to `cmd.ErrOrStderr()`, and (2) add a JSON-mode error interceptor in `Execute()` that writes structured JSON to stdout when the `--json` / `--format=json` flag is detected.

## Technical Context

**Language/Version**: Go 1.25.5  
**Primary Dependencies**: `spf13/cobra` v1.10.2, `spf13/viper` v1.21.0  
**Storage**: N/A — no new persistence; auth config already exists  
**Testing**: `go test ./...` with `net/http/httptest` for API, `cmd.SetOut()` / `cmd.SetErr()` buffers for CLI  
**Target Platform**: macOS/Linux CLI  
**Project Type**: CLI (internal refactoring — no new commands)  
**Performance Goals**: No measurable overhead — the change is output-stream routing only  
**Constraints**: No breaking changes to table-mode output; no change to public command surface  
**Scale/Scope**: ~15 command files; ~25 call-sites need stdout→stderr reclassification; 1 new helper in `internal/output/`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The project constitution is a template placeholder (not yet ratified). No gates are defined; proceeding without block. Re-check after design is complete.

**Post-design re-check**: The design adds helpers to an existing package and reclassifies ~25 call-sites. No new packages, no new dependencies, no new persistence. No violations identified.

## Project Structure

### Documentation (this feature)

```text
specs/006-zero-spillage/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── json-error.schema.json
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

This feature is a pure refactoring — no new directories. Changes touch existing files only:

```text
internal/
├── output/
│   └── output.go        # + WriteError(), DetectJSONMode(), SuggestionFor() helpers
├── cmd/
│   ├── root.go          # Modify Execute() to use WriteError
│   ├── users.go         # Move mutation confirmations → ErrOrStderr (~6 sites)
│   ├── groups.go        # Move mutation confirmations → ErrOrStderr (~10 sites)
│   ├── consumers.go     # Move mutation confirmations → ErrOrStderr (~5 sites)
│   ├── accounts.go      # Audit (most already correct)
│   ├── signals.go       # Audit (already mostly correct)
│   └── [other cmd files]# Audit for any stray stdout writes
└── (no new packages)
```

**Structure Decision**: Single project, existing layout. Feature touches only `internal/output/` and `internal/cmd/`. No new packages, no new dependencies.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations. This is a targeted internal refactoring with minimal blast radius.
