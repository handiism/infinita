# ADR-004: Adopt ULID for Transaction Identifiers

## Status
- Accepted

## Context
- Transaction IDs needed to stay short, sortable, and safe for future distributed scenarios.
- UUID v4 was unique enough, but not sortable and more verbose for CLI and storage use.

## Decision
- Use ULID for transaction identifiers.
- Keep ID generation centralized in one function.
- Continue storing transaction IDs as text without changing the schema shape.

## Consequences
- Transaction IDs are shorter and naturally sortable.
- The project adds a ULID dependency and shifts ID generation behavior away from UUID.

## Related Documents
- `docs/trd.md`
- `docs/adr/ADR-002-defer-server-separation-to-roadmap.md`
- `docs/diagrams/trd/001-trd-data-model-erd.mmd`
- https://github.com/ulid/spec
