# Research: Go CLI Foundation & CI/CD Setup

**Feature**: 001-go-cli-scaffold | **Date**: 2026-05-18

## Go Version

**Decision**: Go 1.23  
**Rationale**: Current stable release; introduced range-over-function iterators and improved toolchain management. Aligns with CLAUDE.md's "Go 1.21+" minimum. Using a specific patch-less minor version (`go 1.23`) in `go.mod` allows toolchain to auto-select the latest patch.  
**Alternatives considered**: Go 1.21 (minimum from CLAUDE.md — too old, misses 2 years of improvements), Go 1.22 (stable, slightly older)

---

## Project Module Path

**Decision**: `github.com/activtrak-mfinlayson/atadmin`  
**Rationale**: Matches the organization (ActivTrak (mfinlayson)) and binary name (`atadmin`). Consistent with the `com.birchgrovesoftware.browsetrak` native messaging host naming already in use in the org.  
**Alternatives considered**: `github.com/activtrak/atadmin` (uses product name instead of org name)

---

## Configuration File Location

**Decision**: XDG Base Directory — `~/.config/atadmin/config.yaml`  
**Rationale**: XDG is the modern standard adopted by `kubectl`, `helm`, `gh`, `stripe`. Viper adds `$HOME/.config/atadmin` and `$HOME/.atadmin` as search paths (latter for backward compat). Config file is saved with `0600` permissions per CLAUDE.md auth guidelines.  
**Alternatives considered**: `~/.atadmin/config.yaml` (legacy dot-directory, no XDG compliance)

---

## Linter Configuration

**Decision**: Enable the following linters beyond golangci-lint defaults:  
- `errcheck` — enforce error return handling  
- `gocritic` — style and correctness  
- `gosimple` — simplification hints  
- `staticcheck` — SA-class static analysis  
- `stylecheck` — naming and commentary rules  
- `revive` — replaces deprecated `golint`  
- `nolintlint` — prevents misuse of `//nolint` directives  
- `exhaustive` — exhaustive switch on enums  

**Rationale**: These linters catch real bugs and enforce the code conventions in CLAUDE.md (strong typing, no ignored errors, consistent style) without being overly noisy on a new project.  
**Alternatives considered**: Enabling `wrapcheck` (too strict for a scaffold), `godot` (comment-period enforcement — low value)

---

## GitHub Actions Versions

**Decision**:  
- `actions/checkout@v4`  
- `actions/setup-go@v5`  
- `golangci/golangci-lint-action@v6`  
- `actions/upload-artifact@v4`  

**Rationale**: These are the current stable major versions. `golangci-lint-action` v6 supports the latest lint engine.  
**Alternatives considered**: Pinning to specific minor versions — deferred; major version pins are standard practice for GitHub Actions.

---

## CI Pipeline Design

**Decision**: Single job with sequential steps (build → lint → test → security scan → upload artifact).  
**Rationale**: A single job keeps the pipeline simple for a scaffold. Build failure short-circuits lint, test, and scan. Lint and test failures both block merge (FR-013, FR-009). Artifact upload only runs on `main` branch pushes. The pipeline runs on every push to any branch and on every PR to main (FR-007).  
**Alternatives considered**: Matrix strategy for multiple OS/Go versions — deferred to a later feature per the Assumptions; parallel jobs — unnecessary complexity at scaffold stage.

---

## Auth Subcommand Strategy

**Decision**: `atadmin auth login` as the primary auth entry point — browser-open → paste token → validate → save config.  
**Rationale**: Matches CLAUDE.md's explicit auth UX spec: `github.com/pkg/browser` to open token generation page, masked token input, lightweight validation call, `0600` config save. Scaffold creates the command stub; full implementation is a subsequent feature.  
**Alternatives considered**: API-key-in-flag only (no persistent auth) — rejected; CLAUDE.md requires persistent config.

---

## RoundTripper Auth Pattern

**Decision**: Custom `authRoundTripper` struct implementing `http.RoundTripper`, cloning the request and injecting `Authorization: Bearer <token>` before calling the inner transport.  
**Rationale**: Directly specified in CLAUDE.md. Keeps API client methods free of auth logic. Composable: a logging RoundTripper can wrap the auth RoundTripper.  
**Alternatives considered**: Injecting headers in each API method call — explicitly rejected by CLAUDE.md.
