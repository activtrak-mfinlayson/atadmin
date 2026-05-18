# Data Model: Go CLI Foundation & CI/CD Setup

**Feature**: 001-go-cli-scaffold | **Date**: 2026-05-18

This scaffold has no persistent data store. The "data model" describes the in-process types that represent configuration, the command hierarchy, and the API client shape established by the scaffold.

---

## Config

The user-level configuration loaded by Viper from file, environment variables, and flags.

| Field | Type | Source | Description |
|-------|------|--------|-------------|
| `token` | string | file / `ATADMIN_TOKEN` / `--token` flag | API bearer token |
| `base_url` | string | file / `ATADMIN_BASE_URL` / `--base-url` flag | Remote API base URL |
| `format` | string | file / `ATADMIN_FORMAT` / `--format` flag | Output format: `table` (default) or `json` |
| `timeout` | duration | file / `ATADMIN_TIMEOUT` | HTTP request timeout (default: 30s) |

**Config file path** (Viper search order):
1. `$XDG_CONFIG_HOME/atadmin/config.yaml`
2. `$HOME/.config/atadmin/config.yaml`
3. `$HOME/.atadmin/config.yaml` (legacy fallback)

**File permissions**: Saved with `0600` (owner read/write only).

**Precedence** (highest to lowest): CLI flag → environment variable → config file → built-in default.

---

## Command Hierarchy

The `atadmin` command tree at scaffold completion. Commands marked `[stub]` are registered but not yet implemented.

```
atadmin
├── --version         (flag — prints version string, exits 0)
├── --format          (persistent flag — output format for all subcommands)
├── --token           (persistent flag — overrides config token)
├── --base-url        (persistent flag — overrides config base URL)
│
└── auth              (subcommand group)
    └── login [stub]  (opens browser, prompts for token, validates, saves config)
```

Each subcommand group follows noun-verb structure: `atadmin <resource> <action>`.

---

## API Client

The `Client` struct wraps `net/http` with auth injection and is instantiated per command invocation.

| Field | Type | Description |
|-------|------|-------------|
| `BaseURL` | `*url.URL` | Parsed remote API base URL |
| `HTTPClient` | `*http.Client` | Underlying HTTP client with `authRoundTripper` set as Transport |
| `UserAgent` | string | `atadmin/<version>` |

**`authRoundTripper`** (implements `http.RoundTripper`):

| Field | Type | Description |
|-------|------|-------------|
| `token` | string | Bearer token injected into every request |
| `inner` | `http.RoundTripper` | Delegate transport (defaults to `http.DefaultTransport`) |

The `authRoundTripper.RoundTrip` method:
1. Clones the incoming request (never mutates the original)
2. Sets `Authorization: Bearer <token>` on the clone
3. Delegates to `inner.RoundTrip(clonedReq)`

---

## State Transitions

The scaffold has one meaningful state: whether the user is authenticated (token present in config).

```
[unauthenticated]
        │
        │  atadmin auth login (future)
        ▼
[authenticated]
        │
        │  token expires / user runs auth logout (future)
        ▼
[unauthenticated]
```

At scaffold stage, the CLI exits with a clear error if `token` is empty when a command requiring auth is invoked.
