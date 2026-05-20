# Spec: Context Window Protections

## 1. Overview
Add context window protections to the CLI to prevent massive JSON payloads from crashing LLM agents.

## 2. Why
AI agents and LLMs using this CLI tool have limited context windows. Commands that return large datasets (like `atadmin users list --json`) dump megabytes of data, causing token limit errors, stalling agents, and incurring high API costs. We need safe defaults and filtering to make the CLI "AI-friendly".

## 3. Requirements (The "What")
1. **Field Filtering (`--fields`)**: Let agents extract only the specific keys they need from the JSON response (e.g., `--fields id,email`). 
    - *Decision:* This will be implemented **client-side** (fetch full payload from API, strip keys before writing to stdout) to avoid modifying remote APIs.
    - *Decision:* Supports **top-level keys only** for V1 to keep the implementation fast and simple.
2. **Safe JSON Pagination**: Enforce smaller pagination limits by default when the CLI is in JSON output mode.
    - *Decision:* If `--json` is passed without an explicit `--limit`, the CLI will default to a safe limit of **50 items**.
3. **Summary Mode (`--summary`)**: Add a flag to return aggregate statistics rather than the full arrays of objects.
    - *Decision:* Output will take the shape `{"total_items": N, "returned_items": N, "has_more": true/false}`.

## 4. Out of Scope
- Modifying the underlying remote API.
- Deeply nested JSON field filtering (e.g., `user.profile.email`).
- Non-JSON output formats (these protections are specifically targeting JSON dumps meant for machine/LLM consumption).

## 5. Review & Acceptance Checklist
- [ ] Running a list command with `--fields id,email` outputs JSON containing *only* those top-level keys for each object.
- [ ] Running a list command with `--json` without an explicit limit defaults to returning a maximum of 50 items.
- [ ] Running a list command with `--summary` outputs aggregate statistics instead of the data array.