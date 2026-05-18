# Feature Specification: Go CLI Foundation & CI/CD Setup

**Feature Branch**: `001-go-cli-scaffold`
**Created**: 2026-05-18
**Status**: Draft
**Input**: User description: "please create a git repo, and scaffold a go CLI. we want github actions to build. please pay attention to @CLAUDE.md"

## Clarifications

### Session 2026-05-18

- Q: What should the CLI binary/tool name be? → A: `atadmin`
- Q: Should linting be a required CI gate (blocks PR merge on failure)? → A: Yes — linting failure blocks merge, same weight as test failure

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Developer Runs the CLI Locally (Priority: P1)

A developer clones the repository and wants to build and invoke the CLI to confirm the scaffold is functional before adding any features.

**Why this priority**: This is the foundational capability — everything else depends on a working local build. Without it no development can proceed.

**Independent Test**: A developer can run the build command and invoke the resulting `atadmin` binary with `--help`, receiving a structured help message with the tool name and available commands.

**Acceptance Scenarios**:

1. **Given** a freshly cloned repository, **When** a developer runs the build command, **Then** a binary is produced in the output directory with no errors.
2. **Given** the built binary, **When** the developer runs `atadmin --help`, **Then** a structured help message displays "atadmin", the tool description, and available commands.
3. **Given** the built binary, **When** the developer runs it with `--version`, **Then** the current version string is printed.
4. **Given** an unrecognized command, **When** the developer runs it, **Then** a clear error message is printed to stderr with a hint to run `--help`.

---

### User Story 2 - Developer Runs the Test Suite (Priority: P2)

A developer wants to confirm the project's test infrastructure is wired up and that an example test passes before contributing.

**Why this priority**: A working test suite is a prerequisite for confident contribution. It validates that the project's quality gates exist from day one.

**Independent Test**: A developer can run the test command and see all tests pass with a coverage summary, validating both the test runner and the initial test coverage.

**Acceptance Scenarios**:

1. **Given** the repository, **When** a developer runs the test command, **Then** all tests pass and a coverage report is printed.
2. **Given** a deliberate test failure (e.g., wrong assertion), **When** the tests are run, **Then** the failure is clearly reported with the file and line number.

---

### User Story 3 - CI Pipeline Validates Pull Requests (Priority: P3)

A contributor opens a pull request. The CI pipeline automatically builds the project, runs all tests, and reports the result back to GitHub so the team knows whether it is safe to merge.

**Why this priority**: Automated validation prevents broken code from entering the main branch and creates a clear quality signal without manual review overhead.

**Independent Test**: A pull request can be opened against the main branch and, within a reasonable time, a green/red status check appears on GitHub reflecting the build and test results.

**Acceptance Scenarios**:

1. **Given** a pull request is opened or updated, **When** CI triggers, **Then** the pipeline builds the project and runs all tests automatically.
2. **Given** all tests pass, **When** CI completes, **Then** a green status check is posted to the pull request.
3. **Given** a test fails, **When** CI completes, **Then** a red status check is posted and logs indicate which test failed.
4. **Given** a pull request with a linting violation, **When** CI runs, **Then** a red status check is posted and the linting output identifies the offending file and rule.
5. **Given** a push to the main branch, **When** CI triggers, **Then** the same build, lint, and test steps run and a binary artifact is produced and stored.

---

### Edge Cases

- What happens when the build environment is missing a required tool or dependency? The build fails with a clear, actionable message identifying the missing component.
- What happens if a CI job is cancelled mid-run? The pipeline reports the run as cancelled (not succeeded) and does not upload partial artifacts.
- What happens if two CI jobs run simultaneously for the same branch? Each runs independently; results do not interfere with each other.
- What happens if the linter is not installed in the CI environment? The pipeline fails the lint stage with a clear message indicating the linter is unavailable, not a code style error.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI binary is named `atadmin`. It MUST expose a root command that displays a help message including the tool name, purpose, and list of available subcommands when invoked with `--help` or with no arguments.
- **FR-002**: The CLI MUST report its version string when invoked with `--version`.
- **FR-003**: The CLI MUST print all error messages and diagnostics to stderr and all data output to stdout.
- **FR-004**: The CLI MUST follow a noun-verb command structure (e.g., `<tool> <resource> <action>`) to enable logical grouping as the command set grows.
- **FR-005**: The CLI MUST read configuration from a user-level config file and support overrides via environment variables, with command-line flags taking the highest precedence.
- **FR-006**: The CLI MUST provide an `auth` subcommand group as a placeholder for future authentication flows.
- **FR-007**: The automated build pipeline MUST run on every push to any branch and on every pull request targeting the main branch.
- **FR-008**: The automated build pipeline MUST compile the project and report failure if compilation fails.
- **FR-009**: The automated build pipeline MUST run all tests and report failure if any test fails.
- **FR-010**: The automated build pipeline MUST upload a compiled binary as a downloadable artifact on successful runs against the main branch.
- **FR-011**: The project structure MUST separate CLI presentation logic from API communication logic, with each layer in its own directory.
- **FR-012**: The project MUST include a security vulnerability scan as part of the CI pipeline.
- **FR-013**: The automated build pipeline MUST run a linter on every push and pull request; a linting failure MUST block merge in the same way as a test failure.

### Key Entities

- **CLI Binary**: The compiled executable named `atadmin`, delivered to end users. Has a version and a set of registered commands.
- **Command**: A named operation exposed to users. Has a noun-verb path, flags, and produces output to stdout or stderr.
- **Configuration**: User-level settings (API endpoint, token, output format). Loaded from a config file with environment variable and flag overrides.
- **CI Pipeline**: The automated workflow triggered by code changes. Composed of stages: build, lint, test, security scan, and artifact upload. Lint and test failures both block merge.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer unfamiliar with the project can clone the repository, build `atadmin`, and run `atadmin --help` successfully within 5 minutes.
- **SC-002**: The CI pipeline completes a full build-test-scan cycle in under 5 minutes for a standard changeset.
- **SC-003**: 100% of tests must pass in CI before a pull request can be considered for merge.
- **SC-004**: A compiled binary artifact is available for download within 10 minutes of a successful push to the main branch.
- **SC-005**: Any failed test in CI produces a log entry sufficient to identify the failing assertion without reading source code.
- **SC-006**: The CLI returns a non-zero exit code for every error condition and zero for every success, enabling reliable scripting.

## Assumptions

- The CLI is intended for internal operator/admin use, not end-user-facing consumer distribution.
- The initial scaffold ships with zero production API integration; all API client stubs are placeholders to be filled in subsequent features.
- The CI environment provides a standard hosted runner (Linux) with the required language runtime pre-installed.
- Authentication for the target API is token-based (user-generated API key), not interactive OAuth; the `auth` subcommand structure reflects this.
- A single binary is sufficient for the initial scaffold; multi-platform cross-compilation is deferred to a later feature.
- The config file location follows the OS-appropriate user config directory convention.
- The project uses semantic versioning; the initial scaffold version is `0.1.0`.
