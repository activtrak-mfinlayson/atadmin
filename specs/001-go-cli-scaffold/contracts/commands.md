# CLI Command Contracts: atadmin

**Feature**: 001-go-cli-scaffold | **Date**: 2026-05-18  
**Binary**: `atadmin`  
**Version**: 0.1.0

This document defines the command interface contracts for `atadmin`. Contracts are binding: changes to flags, output format, or exit codes are breaking changes requiring a new version.

---

## Global Flags (Persistent — available on all subcommands)

| Flag | Short | Type | Default | Env Override | Description |
|------|-------|------|---------|-------------|-------------|
| `--format` | `-f` | string | `table` | `ATADMIN_FORMAT` | Output format: `table` or `json` |
| `--token` | | string | (config) | `ATADMIN_TOKEN` | API bearer token |
| `--base-url` | | string | (config) | `ATADMIN_BASE_URL` | Remote API base URL |
| `--help` | `-h` | bool | false | | Show help for the current command |

---

## Root Command: `atadmin`

**Invocation**: `atadmin [--version] [--help]`

**Behavior**:
- With no arguments or `--help`: prints help text listing available subcommands.
- With `--version`: prints `atadmin version 0.1.0` to stdout, exits 0.
- With an unrecognized subcommand: prints error to stderr, exits 1 with hint to run `atadmin --help`.

**Exit codes**:
- `0` — success
- `1` — user error (bad flag, unknown command, invalid argument)
- `2` — runtime error (API unreachable, auth failure)

**Stdout (version)**:
```
atadmin version 0.1.0
```

**Stderr (unknown command)**:
```
Error: unknown command "foo" for "atadmin"
Run 'atadmin --help' for usage.
```

---

## Subcommand Group: `auth`

**Invocation**: `atadmin auth <action>`

**Purpose**: Authentication lifecycle management.

### `atadmin auth login` *(stub in scaffold — full implementation is a subsequent feature)*

**Behavior (planned)**:
1. Opens browser to token generation page.
2. Prompts user to paste token (masked input).
3. Validates token by calling a lightweight API endpoint.
4. Saves token to config file with `0600` permissions.
5. Prints `Logged in successfully.` to stdout, exits 0.

**Exit codes**:
- `0` — login succeeded, token saved
- `1` — user cancelled
- `2` — token validation failed (API error or invalid token)

**Stub behavior (scaffold only)**:
- Prints `Not yet implemented. See: atadmin auth login --help` to stderr.
- Exits `1`.

---

## Output Contract

### Table format (default, human-readable)

- Written to stdout.
- Tab-aligned columns via `text/tabwriter`.
- Header row in ALL CAPS.
- No trailing whitespace.

### JSON format (`--format json`)

- Written to stdout.
- Compact single-object or array of objects.
- Keys in `camelCase`.
- No extra whitespace or pretty-printing (enables reliable piping and `jq` usage).

### Diagnostic output

- All warnings, progress indicators, and error messages → stderr.
- Never mixed with data output.

---

## Scripting Contract

Commands that succeed silently in non-interactive mode (mutation operations):
- Print only the resource ID or URL to stdout.
- Print nothing else.
- Exit `0`.

Enables: `atadmin resource create | xargs atadmin resource view`

TTY detection: if stdout is not a terminal (`term.IsTerminal` returns false), colors and spinners are disabled automatically. `NO_COLOR=1` environment variable also disables ANSI output.
