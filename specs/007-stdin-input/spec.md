# Spec: Input via Stdin (--from-stdin)

## 1. Overview
Allow LLMs and scripts to pipe JSON payloads directly into mutation commands via `os.Stdin`.

## 2. Why
LLMs often struggle with complex bash escaping (like passing a nested JSON string into a `--metadata` flag). Allowing agents to write a payload to a file and pipe it (`cat payload.json | atadmin users update 123 --from-stdin`) bypasses shell escaping issues entirely.

## 3. Requirements (The "What")
1. **Stdin Flag**: Add a global (or command-specific) `--from-stdin` boolean flag to all mutating commands (e.g., `update`, `create`, `delete`, `bulk`).
    - *Decision:* The CLI will read `os.Stdin`, unmarshal it directly into the expected Go struct for that command, and execute the API call.
2. **Interactive Bypass**: If `--from-stdin` is used, the CLI must automatically skip all interactive confirmation prompts (like "Are you sure? [y/N]").

## 4. Out of Scope
- Supporting formats other than JSON via stdin.

## 5. Review & Acceptance Checklist
- [ ] Running a mutation command with `--from-stdin` correctly reads a JSON payload from stdin.
- [ ] The command successfully skips interactive prompts when `--from-stdin` is present.
- [ ] The command fails gracefully with a structured error if the piped JSON is invalid or missing required fields.
