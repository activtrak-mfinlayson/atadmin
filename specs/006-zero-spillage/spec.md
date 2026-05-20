# Spec: Zero-Spillage Rule

## 1. Overview
Guarantee that the CLI's stdout emits *only* valid JSON when `--json` is active, and ensure errors are formatted as structured JSON to aid autonomous agents in self-correction.

## 2. Why
When agents parse JSON output, any stray logs (like "Connecting..." or ANSI color codes for spinners) printed to stdout break the JSON parser and halt the agent. Agents also struggle with raw text errors; structured errors with a `suggestion` field allow them to auto-correct.

## 3. Requirements (The "What")
1. **Stdout Purity**: All logs, warnings, progress bars, and "Connecting..." messages must be printed to `os.Stderr`. Only final data output goes to `os.Stdout`.
    - *Decision:* We will audit the codebase for `fmt.Print*` and `log.Print*` and redirect all non-data output to `os.Stderr`.
2. **Structured Errors**: When `--json` (or `--format=json`) is enabled, errors must be printed to stdout as a JSON object: `{"error": "...", "suggestion": "..."}`.
    - *Decision:* Create a global error interceptor in `cmd.Execute()` or a standard error printing function that detects the format flag and wraps the error before exiting.

## 4. Out of Scope
- Rewriting the entire logging system (we will just redirect existing output).

## 5. Review & Acceptance Checklist
- [ ] Running a command with `--json` outputs exactly 1 valid JSON object/array to stdout.
- [ ] Any "Loading" or progress text is sent to stderr.
- [ ] Failing a command with `--json` outputs a valid JSON error object containing `error` and `suggestion` fields.
