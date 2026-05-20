# Research: Identity API CLI Commands

**Feature**: 003-identity-api-cli  
**Date**: 2026-05-19

## Findings Summary

All research questions were resolved from the swagger document (`docs/identity-swagger.json`) and the existing codebase patterns. No external sources were required.

| ID | Question | Status | Decision |
|---|---|---|---|
| R-001 | Identity API base URL | Resolved | Same host as Admin API; no extra config |
| R-002 | Revision auto-fetch behavior | Resolved | Always auto-fetch (both TTY and script mode) |
| R-003 | Wire format shapes | Resolved | Three shapes: list page, single entity, bulk response |
| R-004 | PATCH/DELETE endpoint mechanics | Resolved | PATCH path-revision, DELETE query-revision, group ops via body |
| R-005 | Bulk revision handling | Resolved | Pre-fetch each entity (up to 10 concurrent GETs) |
| R-006 | 409 error messaging | Resolved | Extend `checkResponse` with explicit 409 case |

## Detail

See `plan.md` Phase 0 entries R-001 through R-006 for full rationale and alternatives considered.

## No NEEDS CLARIFICATION Markers

The spec had zero unresolved markers at the time planning began. All five clarification questions (command noun, revision mode, agents placement, page size, error handling) were resolved before planning.
