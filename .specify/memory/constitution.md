# Spec Kit Constitution

## Core Principles

### I. Test Integrity
Implementation agents may not modify, delete, skip, or weaken existing tests to make a build pass. If a test appears genuinely wrong, stop and report it as a finding for human review rather than changing it.

### II. Scope Discipline
Tasks only touch files listed in the plan's Context Files. Going outside that list requires stopping and surfacing why.

### III. Commit Cadence
Every task ends in a commit. Commit messages follow `<type>(NNN): <task summary>` where NNN is the feature number from the branch name.

### IV. Stop-and-Report on Repeated Failures
If a task fails twice or a single failure persists through three fix attempts, stop and surface it. Do not keep cycling.

### V. Sandbox Requirement
Any work using `--dangerously-skip-permissions` runs only inside a dedicated git worktree or container, never in the primary working tree.

### VI. Spec Primacy
If implementation reveals the spec is wrong, stop and return to specification rather than patching with code.

## Governance
This Constitution supersedes all other practices. All PRs and multi-agent workflows must verify compliance with these six hard rules.
