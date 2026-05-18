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

## Full quickstart

See [specs/001-go-cli-scaffold/quickstart.md](specs/001-go-cli-scaffold/quickstart.md) for configuration, adding new commands, and CI details.
