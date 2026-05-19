# Implementation Plan: ActivTrak Admin API CLI Wrapper

**Branch**: `002-wrap-admin-api` | **Date**: 2026-05-18 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/002-wrap-admin-api/spec.md`

## Summary

Wrap all 110 endpoints of the ActivTrak Admin API into the `atadmin` CLI using 11 resource-grouped subcommand namespaces (noun-verb structure). Builds on the feature-001 scaffold (cobra, viper, authRoundTripper) by adding named profile support, retry-on-429, `--verbose` tracing, TTY-aware output formatting, and bulk file parsing (JSON+CSV). All API client methods are covered by `httptest`-backed tests; all CLI commands are covered by cobra buffer tests.

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**: cobra v1.10, viper v1.21, `golang.org/x/term` (TTY detection; add via `go get`) — standard library only for retry, tabwriter, csv, json
**Storage**: `~/.config/atadmin/config.yaml` (profile-namespaced YAML, `0600` permissions)
**Testing**: `go test ./...`; API layer uses `net/http/httptest`; CLI layer uses `cmd.SetOut()` / `cmd.SetErr()` buffers
**Target Platform**: macOS, Linux, Windows (cross-platform)
**Project Type**: CLI
**Performance Goals**: < 5s per command round-trip (3 retries × ~1s backoff)
**Constraints**: No bloated HTTP frameworks; `text/tabwriter` for tables; retry max 3 attempts
**Scale/Scope**: 110 API endpoints, 11 resource groups, ~30 new Go source files in `internal/`

## Constitution Check

Constitution template is unfilled — no constitutional gates defined. No violations to assess.

## Project Structure

### Documentation (this feature)

```text
specs/002-wrap-admin-api/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── contracts/
│   └── cli-commands.md  # Phase 1 output — full command contract
└── tasks.md             # Phase 2 output (created by /speckit-tasks)
```

### Source Code (repository root)

```text
cmd/atadmin/
└── main.go                          # existing — no changes needed

internal/
  api/
    client.go                        # EXTEND: add retryRoundTripper, verboseRoundTripper; update NewClient signature to accept verbose bool and max retries
    models.go                        # EXTEND: add all entity structs (Client, Consumer, Device, Group, Alarm, Signal, Schedule, ApiKey, AuditLog, all AccountSettings sub-types)
    accounts.go                      # NEW: account settings API methods (GET/PUT for all 21 sub-resources)
    accounts_test.go
    alarms.go                        # NEW: alarm CRUD + conditions/fields
    alarms_test.go
    apikeys.go                       # NEW: API key CRUD + util backfill
    apikeys_test.go
    auditlog.go                      # NEW: audit log list + attachment download
    auditlog_test.go
    clients.go                       # NEW: client CRUD, merge, restore, DNT, alias
    clients_test.go
    consumers.go                     # NEW: consumer CRUD, role, password, SSO, groups
    consumers_test.go
    devices.go                       # NEW: device list, get, delete, restore, uninstall
    devices_test.go
    groups.go                        # NEW: group CRUD, member management, bulk import/export
    groups_test.go
    hrdc.go                          # NEW: HRDC ping + bulk import
    hrdc_test.go
    schedules.go                     # NEW: schedule CRUD, user assignments, reporting/shift variants
    schedules_test.go
    signals.go                       # NEW: signal CRUD
    signals_test.go

  cmd/
    root.go                          # EXTEND: add --profile, --verbose persistent flags; wire all 11 subcommand groups
    root_test.go                     # existing + new flag tests
    auth.go                          # IMPLEMENT: auth login with --profile flag, browser open, masked input, token validation, 0600 config write
    accounts.go                      # NEW: atadmin settings <sub-resource> get|set
    accounts_test.go
    alarms.go                        # NEW: atadmin alarms list|get|details|create|update|delete|conditions|fields
    alarms_test.go
    apikeys.go                       # NEW: atadmin apikeys list|create|update|delete|util backfill
    apikeys_test.go
    auditlog.go                      # NEW: atadmin auditlog list|attachment get
    auditlog_test.go
    clients.go                       # NEW: atadmin clients list|get|update|delete|restore|merge|alias|donottrack
    clients_test.go
    consumers.go                     # NEW: atadmin consumers list|get|create|update|delete|role|password|sso|groups|chrome-users
    consumers_test.go
    devices.go                       # NEW: atadmin devices list|get|delete|restore|uninstall
    devices_test.go
    groups.go                        # NEW: atadmin groups list|summary|get|search|create|rename|delete|members|clients|devices
    groups_test.go
    hrdc.go                          # NEW: atadmin hrdc ping|import
    hrdc_test.go
    notifications.go                 # NEW: atadmin notifications list [DEPRECATED]
    schedules.go                     # NEW: atadmin schedules list|get|create|delete|reporting|shift|users|user
    schedules_test.go
    signals.go                       # NEW: atadmin signals list|create|update|delete
    signals_test.go

  config/
    config.go                        # EXTEND: profile support — LoadProfile(name string), ListProfiles(), SaveProfile(name string, cfg *Config); nested YAML under profiles.<name>.*
    config_test.go

  output/                            # NEW package
    output.go                        # Table(out, headers, rows), KeyValue(out, fields), JSON(out, v), PrintError(err, stderr)
    output_test.go

  tty/                               # NEW package
    tty.go                           # IsTerminal() bool — wraps golang.org/x/term

  bulk/                              # NEW package
    bulk.go                          # ParseFile(path string) ([]map[string]any, error) — JSON+CSV auto-detect by extension
    bulk_test.go
```

## Implementation Sequence

The following order minimizes blocked work:

1. **Shared infrastructure first** — `output/`, `tty/`, `bulk/`, config profile support, and client.go RoundTripper extensions. These are dependencies for everything else.
2. **API layer per resource group** — each `internal/api/<resource>.go` with `httptest` tests. Start with clients (most complex), then groups, consumers, devices, then settings, schedules, alarms, signals, apikeys, auditlog, hrdc.
3. **CLI layer per resource group** — wire each `internal/cmd/<resource>.go` after its API counterpart passes tests.
4. **Auth implementation** — implement `auth login` with `--profile` support and token validation ping.
5. **Root.go wiring** — add all subcommands to root after each group is complete.

## Key Design Decisions

### RoundTripper Chain

```
verboseRoundTripper → retryRoundTripper → authRoundTripper → http.DefaultTransport
```

- `verboseRoundTripper`: logs `> METHOD URL` and `< STATUS` to stderr when `--verbose` is set
- `retryRoundTripper`: retries on HTTP 429 up to 3 times with exponential backoff (`2^n` seconds)
- `authRoundTripper`: injects `Authorization: Bearer <token>` (already implemented)

### Profile Loading

`config.Load(profileName string)` reads `v.Sub("profiles." + profileName)`. If the sub-key is absent, returns an error listing available profile names. The `--profile` flag (default: `"default"`) is registered as a persistent pre-run flag on root, resolved before any subcommand runs.

### Output Routing

| Response shape | Default format | `--json` |
|---|---|---|
| List (array) | tabwriter table | raw JSON array |
| Single object / settings | key: value pairs | raw JSON object |
| Mutation success | resource ID on stdout | same |
| Error | actionable message on stderr | same (errors never go to stdout) |

### Bulk File Handling

`bulk.ParseFile(path)` returns `[]map[string]any`. Each resource's bulk command maps this to its typed request struct. File extension `.json` or `.csv` is required; other extensions return an actionable error.

### TTY-Aware Behavior

`tty.IsTerminal()` is called at command startup:
- **true** (interactive): Use table/key-value formatting, interactive password prompts
- **false** (piped/CI): Bare output only (just IDs on mutations), fail fast on missing required flags

## Complexity Tracking

No constitution violations; no complexity justification required.
