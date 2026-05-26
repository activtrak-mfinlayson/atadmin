# Spec: Safe Exploration (--dry-run)

## 1. Overview
Add a `--dry-run` flag to all mutating commands to allow agents and humans to preview changes without executing them.

## 2. Why
For agents operating in human-in-the-loop modes ("Ask before executing"), they need to show the user what they will do before doing it. A dry run provides a safe, verifiable JSON diff of the proposed changes.

## 3. Requirements (The "What")
1. **Dry Run Flag**: Add a `--dry-run` boolean flag to mutating commands (`create`, `update`, `delete`, `bulk`).
    - *Decision:* When true, the API client will *skip* making the actual HTTP `POST/PUT/PATCH/DELETE` request.
2. **Output Format**: Instead of executing, the command will print a JSON representation of the intended action.
    - *Decision:* The output will match the structure: `{"action": "update", "target": "users/123", "payload": {...}}` to make it easily parseable by agents.

## 4. Out of Scope
- Actually hitting the remote API and relying on a server-side "dry run" feature (since the remote API might not support it; we will just short-circuit in the CLI client).

## 5. Review & Acceptance Checklist
- [ ] Running a mutation command with `--dry-run` does not alter the remote state (no HTTP request sent).
- [ ] The command outputs a JSON object describing the action, target, and payload.
