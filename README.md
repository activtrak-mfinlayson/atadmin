# atadmin

ActivTrak administration command-line interface.

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [golangci-lint](https://golangci-lint.run/usage/install/) (for linting)

## Build

```bash
go build -o bin/atadmin ./cmd/atadmin
./bin/atadmin --help
```

## Test

```bash
go test ./...
```

## Lint

```bash
golangci-lint run
```

## Security scan

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## Configuration

`atadmin` stores credentials in `~/.config/atadmin/config.yaml` (permissions `0600`).

```bash
# Authenticate (opens browser, prompts for token):
atadmin auth login

# Use a named profile:
atadmin auth login --profile staging
atadmin clients list --profile staging

# Override via environment variables:
export ATADMIN_TOKEN=your-token
export ATADMIN_BASE_URL=https://api.activtrak.com
```

## Usage Examples

```bash
# Clients
atadmin clients list
atadmin clients get 42
atadmin clients update 42 --alias "Jane Smith"
atadmin clients delete --ids 10,11,12
atadmin clients merge --source 10 --target 20
atadmin clients donottrack list

# Groups
atadmin groups list
atadmin groups create "Engineering"
atadmin groups members add --group 5 --member 42 --type client

# Consumers (admin/report users)
atadmin consumers list
atadmin consumers create --file users.csv
atadmin consumers role set 7 --role viewer

# Devices
atadmin devices list
atadmin devices delete --ids 1,2,3

# Schedules
atadmin schedules list
atadmin schedules reporting default get

# Signals & Alarms
atadmin signals list
atadmin alarms list --json

# API Keys
atadmin apikeys list
atadmin apikeys create --name "CI Token"

# Audit Log
atadmin auditlog list --from 2024-01-01 --to 2024-01-31

# Account Settings
atadmin settings privacy get
atadmin settings sso enabled
atadmin settings ping

# HRDC Integration
atadmin hrdc ping
atadmin hrdc import --file employees.csv
```

## Global Flags

| Flag | Description |
|---|---|
| `--profile <name>` | Use a named config profile (default: `default`) |
| `--token <str>` | Override bearer token (env: `ATADMIN_TOKEN`) |
| `--base-url <url>` | Override API base URL (env: `ATADMIN_BASE_URL`) |
| `--verbose` | Print HTTP request/response details to stderr |
| `--format table\|json` | Output format (env: `ATADMIN_FORMAT`) |

## Output

- **Table format** (default): aligned columns for lists, key-value pairs for single items.
- **JSON format** (`--json` or `--format=json`): raw JSON for scripting.
- **Diagnostics** (errors, warnings, progress) go to stderr; data goes to stdout.
- **Non-TTY/script mode**: mutation commands print only the resource ID to stdout.

## Full quickstart

See [specs/001-go-cli-scaffold/quickstart.md](specs/001-go-cli-scaffold/quickstart.md) for configuration, adding new commands, and CI details.
