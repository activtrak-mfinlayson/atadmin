# Quickstart: atadmin

**Feature**: 001-go-cli-scaffold | **Date**: 2026-05-18

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.23+ | https://go.dev/dl/ |
| golangci-lint | latest | `brew install golangci-lint` or https://golangci-lint.run/usage/install/ |
| git | any | pre-installed on most systems |

## Clone & Build

```bash
git clone https://github.com/activtrak-mfinlayson/atadmin
cd atadmin
go build -o bin/atadmin ./cmd/atadmin
```

Verify the build:

```bash
./bin/atadmin --help
./bin/atadmin --version
```

Expected output from `--version`:
```
atadmin version 0.1.0
```

## Run Tests

```bash
go test ./...
```

With coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out   # opens browser
```

## Run Linter

```bash
golangci-lint run
```

Zero warnings expected on a clean scaffold.

## Run Security Scan

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## Configure atadmin

Create the config directory and file:

```bash
mkdir -p ~/.config/atadmin
cat > ~/.config/atadmin/config.yaml << 'EOF'
token: ""
base_url: "https://api.activtrak.com"
format: "table"
timeout: "30s"
EOF
chmod 0600 ~/.config/atadmin/config.yaml
```

Override any value with environment variables:

```bash
export ATADMIN_TOKEN=your-api-token
export ATADMIN_BASE_URL=https://api.activtrak.com
export ATADMIN_FORMAT=json
```

Or pass as flags:

```bash
atadmin --token your-token --format json <resource> <action>
```

## CI (GitHub Actions)

The pipeline runs automatically on every push and pull request. No manual steps required.

To see the pipeline configuration: `.github/workflows/ci.yml`

Pipeline stages:
1. **Build** — `go build -o bin/atadmin ./cmd/atadmin`
2. **Lint** — `golangci-lint run` (blocks merge on failure)
3. **Test** — `go test ./...`
4. **Security scan** — `govulncheck ./...`
5. **Upload artifact** — `bin/atadmin` uploaded on main-branch pushes only

## Add a New Command

Follow the noun-verb pattern (`atadmin <resource> <action>`):

1. Add API model and client method in `internal/api/` with `httptest` coverage.
2. Add Cobra command in `internal/cmd/` and register it on the appropriate parent.
3. Ensure the command supports `--format json` if it returns data.
4. Run `go test ./...` and `golangci-lint run` before opening a PR.
