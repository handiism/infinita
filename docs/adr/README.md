# ADR Guide

Use an ADR when a change affects architecture, boundaries, storage, transport, validation strategy, or a lasting engineering rule.

Authoritative checklist:

1. Read existing ADRs in `docs/adr/` first.
2. Create one ADR for one decision.
3. Use the next sequential ID with a short slug.
4. If replacing an old decision, create a new ADR and mark the old one as superseded.
5. Link the relevant `docs/prd.md`, `docs/trd.md`, diagrams, and related ADRs.

Minimal template:

```md
# ADR-XXX: [Decision Title]

## Status
- Proposed | Accepted | Superseded

## Context
- What changed?
- Why is a decision needed now?

## Decision
- Chosen approach

## Consequences
- Trade-offs
- Follow-up changes needed

## Related Documents
- `docs/prd.md`
- `docs/trd.md`
```
