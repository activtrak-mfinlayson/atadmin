# Repository Guidelines (Go API CLI)

This repository is for a Go-based CLI tool that acts as a wrapper for a remote API (similar to the `gh` or `kubectl` CLIs). The goal is to build a fast, testable, and user-friendly command-line interface.

## Intended Stack
- **Language:** Go (1.21+).
- **CLI Framework:** `spf13/cobra` for command routing and flag parsing.
- **Configuration:** `spf13/viper` for handling config files (`~/.config/app/config.yaml`) and environment variables.
- **HTTP Client:** Standard `net/http` combined with custom `http.RoundTripper` implementations for auth injection and logging. No bloated HTTP frameworks.
- **Formatting:** Standard `text/tabwriter` for human-readable tables, standard `encoding/json` for machine-readable output.

## Development Principles
- **Separation of Concerns:** Keep CLI presentation logic (Cobra commands) strictly separate from API communication logic (HTTP client).
- **Machine Readability:** Every command that fetches data MUST support a `--json` (or `--format=json`) flag to output raw JSON, enabling it to be easily scripted by other tools (or AI agents).
- **Test-First API Clients:** When writing API wrappers, use `net/http/httptest` to mock the remote API. Test the actual HTTP serialization and deserialization, not just an interface mock.
- **Graceful Degradation:** The CLI should provide clear, actionable error messages when the API is unreachable, times out, or returns a 4xx/5xx status code.

## Expected Project Shape
- `cmd/appname/`: The `main.go` entrypoint. Should be virtually empty, just calling `cmd.Execute()`.
- `internal/cmd/`: Cobra command definitions (`root.go`, `get.go`, `create.go`). This layer handles flags, reads config, calls the API, and formats output.
- `internal/api/`: The API client layer. Contains the `Client` struct, request/response models, and methods to interact with the remote API.
- `internal/config/`: Configuration structs and Viper initialization.

## Go API Client Guidelines
- **Avoid Global State:** The API client must be instantiated and passed down to commands. Do not use a global HTTP client.
- **Context is King:** Every API method must take a `context.Context` as its first argument to support timeouts and cancellations.
- **Auth Injection:** Use a custom `http.RoundTripper` to automatically inject Bearer tokens or API keys into requests. Do not manually add auth headers in every API method.
- **Strong Typing:** Define explicit Go structs for API requests and responses using JSON struct tags. Avoid `map[string]interface{}`.

## CLI & Output Guidelines
- **Standard Streams:** Write all normal output (data, tables, JSON) to `os.Stdout`. Write all diagnostic output (errors, warnings, "Loading..." messages, progress bars) to `os.Stderr`. This allows users to do `app get items > items.json` without corrupting the file with log messages.
- **Configuration Fallbacks:** Flags override Environment Variables, which override Config Files. Viper handles this automatically if set up correctly.

## Testing Guidelines
- Use table-driven tests for both CLI commands and API methods.
- **API Tests:** Use `httptest.NewServer` to start a local test server, configure the API client to point to its URL, and assert that the client sends the correct HTTP method, path, and payload.
- **CLI Tests:** Cobra commands can be tested by passing a buffer to `cmd.SetOut()` and `cmd.SetErr()`, executing the command, and asserting against the buffered output.

## Top-Tier CLI UX Patterns (Inspired by `gh` and `stripe`)
- **TTY Detection:** Use `golang.org/x/term` to detect if the CLI is running interactively. If `term.IsTerminal(int(os.Stdout.Fd()))` is true, it is safe to use colors, spinners, and interactive prompts (e.g., using `AlecAivazis/survey`). If false (running in CI or piped), disable colors, skip prompts, and fail fast if required arguments are missing. Never hang waiting for interactive stdin in a script.
- **Noun-Verb Command Structure:** Structure commands as `<app> <resource> <action>` (e.g., `app user create`, `app project list`) rather than verb-noun (`app create user`). This scales much better and groups help menus logically by resource.
- **Actionable Errors:** Never return raw HTTP errors to the user. Wrap them in actionable advice. Instead of `HTTP 401`, output: `Error: Unauthorized. Your token may have expired. Try running 'app auth login'.`
- **Environment Overrides:** Every config file setting must have an environment variable equivalent with a consistent prefix (e.g., `APP_TOKEN`, `APP_FORMAT`).
- **Graceful Degradation of ANSI:** Automatically honor the `NO_COLOR=1` environment variable to strip ANSI escape codes for environments that don't support them.
- **Silent Success, Loud Failure:** For mutation commands (create/update/delete) in non-interactive/script mode, output *only* the ID/URL of the created resource on success to stdout. This allows easy piping (`app project create | xargs app project view`).
- **Auth Subcommand (Bring-Your-Own-Token):** Since the API requires a webpage-generated API key, make `app auth login` as graceful as possible. It should: 1) Automatically open the user's browser to the exact token generation page (e.g., using `github.com/pkg/browser`), 2) Interactively prompt the user to paste the token (masking the input like a password), 3) Instantly validate the token by hitting a lightweight `/me` or `/health` endpoint, and 4) Save the config file with restricted permissions (`0600`) so the user never has to manually edit dotfiles.

## Exact Commands for Agents
When asked to test, build, or lint, use these exact commands:
- **Go tests:** `go test ./...`
- **Go tests with coverage:** `go test -coverprofile=coverage.out ./...`
- **Go security scan:** `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`
- **Go build:** `go build -o bin/atadmin ./cmd/atadmin`
- **Linting:** `golangci-lint run`

## Agent Instructions
- Before editing, inspect existing conventions in `internal/cmd/` and `internal/api/`.
- When adding a new API endpoint, add the models and client method in `internal/api/` FIRST, along with its `httptest` coverage.
- Next, add the Cobra command in `internal/cmd/` and wire it up to the API client.
- Ensure the new command supports standard output (tables/text) and `--json` output if it returns data.
- Never write API credentials or tokens to logs or stdout.
- If a command or test fails, read the error output, fix the code, and retry autonomously before asking for help.

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan at:
specs/001-go-cli-scaffold/plan.md
<!-- SPECKIT END -->
