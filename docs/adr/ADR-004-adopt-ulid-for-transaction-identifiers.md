# ADR-004: Adopt ULID for Transaction Identifiers

## Status

Accepted

## Date

2026-03-29

## Owners / Decision Makers

- Engineering

## Context

Transaction IDs were originally generated using UUID v4 (`github.com/google/uuid`). While UUID v4 provides strong uniqueness guarantees, it has several drawbacks for this project's current and future needs:

- **Not sortable**: UUID v4 is entirely random and cannot be used for chronological ordering.
- **Verbose**: 36 characters with hyphens (or 32 hex without), making CLI output and log lines longer than necessary.
- **Not distributed-safe by design**: While collision probability is negligible, UUID v4 lacks explicit coordination-free guarantees for multi-instance scenarios planned in the roadmap.

The project needs transaction identifiers that are:

1. **Short**: Minimal character footprint for CLI display and storage.
2. **Sortable**: Lexicographically orderable by generation time.
3. **Anti-collision**: Cryptographically strong uniqueness guarantees.
4. **Distributed-safe**: No coordination required between generators.

## Decision

Replace UUID v4 with **ULID** (Universally Unique Lexicographically Sortable Identifier) for transaction IDs, using the `github.com/oklog/ulid/v2` library.

Key implementation details:

- A `NewID()` function is provided in `internal/domain/valueobject/ulid.go` as the single source of ID generation.
- Monotonic entropy with mutex locking ensures IDs are sortable even when generated within the same millisecond.
- The `transactions.id` column remains `TEXT PRIMARY KEY` in SQLite — no schema migration required (ULID is a 26-character Crockford Base32 string).
- Other entities (`categories`, `budgets`, `settings`, `initial_balance`, `analytics_consent`) continue using `INTEGER AUTOINCREMENT` or semantic keys as appropriate for their scope.

## Options Considered

1. **UUID v4 (status quo)** — 36 chars, random, not sortable. Well-known but doesn't meet future needs.
2. **ULID** (chosen) — 26 chars, time-prefixed, lexicographically sortable, 128-bit (48-bit timestamp + 80-bit randomness), no coordination needed.
3. **Snowflake ID** — 64-bit integer, sortable, requires machine/worker ID configuration. Overkill for single-machine MVP; introduces operational complexity.
4. **NanoID** — Configurable length (typically 21 chars), URL-safe, but not time-sortable. Better for short URLs than for transaction identifiers.
5. **UUID v7** — Time-ordered UUID variant, 36 chars. Sortable but still verbose. Less ecosystem support in Go compared to ULID.

## Rationale

- **ULID** meets all four requirements: short (26 chars vs UUID's 36), lexicographically sortable, 128-bit anti-collision, and coordination-free.
- Crockford's Base32 encoding is case-insensitive and avoids ambiguous characters (`I`, `L`, `O`, `U`), making IDs safe for CLI display and manual entry.
- `github.com/oklog/ulid/v2` is the de facto standard Go implementation with built-in monotonic entropy support.
- Drop-in replacement for UUID in the codebase — `Transaction.ID` is already `string`, and the database column is already `TEXT`.
- Future-proofs for multi-device sync scenarios on the roadmap without requiring a schema migration later.

## Consequences

### Positive

- Transaction IDs are now chronologically sortable, enabling natural ordering in CLI output and queries.
- 28% shorter than UUID (26 vs 36 chars), reducing storage and display footprint.
- No coordination needed — safe for future distributed/multi-device scenarios.
- No database schema change required (compatible with existing `TEXT PRIMARY KEY`).
- Monotonic entropy guarantees sortability even within the same millisecond.

### Negative

- Introduces a new external dependency (`github.com/oklog/ulid/v2`).
- ULID timestamp precision is milliseconds (not microseconds), which is acceptable for transaction logging.
- Monotonic entropy with mutex locking introduces minimal contention under high-throughput scenarios (not a concern for MVP CLI usage).

### Operational / Maintenance

- The `NewID()` function in `internal/domain/valueobject` is the single point of ID generation.
- If future entities need ULID IDs, the same function can be reused.
- Existing UUID-format transaction IDs in databases created before this change remain valid — no data migration needed.

## Implementation Notes

- ID generator: `internal/domain/valueobject/ulid.go`
- Consumer: `internal/application/usecase/transaction_usecase.go` (line 52: `valueobject.NewID()`)
- Removed dependency: `github.com/google/uuid`
- Added dependency: `github.com/oklog/ulid/v2`
- Tests verify: length (26), uniqueness, Crockford Base32 charset, and monotonic sortability.

## Related Documents

- `docs/trd/001-infinita-personal-finance-app-mvp-cli.md` (TRD-DATA-001)
- `docs/adr/ADR-002-defer-server-separation-to-roadmap.md`
- `docs/diagrams/trd/001-trd-data-model-erd.mmd`
- ULID specification: https://github.com/ulid/spec
