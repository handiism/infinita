# ADR-003: Adopt sqlc, golang-migrate, and Tiered Testing Strategy

## Status
Accepted

## Date
2026-03-28

## Owners / Decision Makers
- Engineering

## Context
The MVP needs three technical foundations:

1. **Type-safe data access**: Hand-written SQL in Go repositories is error-prone — column mismatches, wrong types, and missing query coverage are caught only at runtime. A compile-time SQL compiler eliminates this class of bugs.

2. **Schema migration management**: SQLite schema will evolve beyond MVP. Manual DDL execution is fragile and unreproducible. A migration tool ensures deterministic schema evolution with version tracking and rollback support.

3. **Reliable testing strategy**: The codebase needs clear testing boundaries — repository layer (database integration), service layer (business logic with mocks), and CLI layer (end-to-end command execution). Each tier has distinct isolation requirements.

Additionally, the project uses SQLite as its sole persistence engine in MVP. SQLite is an embedded, file-based database that runs in-process. It does not run as a network service and therefore does not require containerized instances for testing.

## Decision

### 1. Use sqlc for SQL-to-Go Code Generation

- Write raw SQL queries in `.sql` files with `-- name:` annotations.
- sqlc parses schema and queries, then generates fully type-safe Go code (structs, methods, and an optional `Querier` interface).
- Generated code lives in the **infrastructure layer** (`internal/infrastructure/database/sqlite/sqlc/`).
- Repository implementations wrap sqlc-generated code and map sqlc types to domain types at the infrastructure boundary.
- Domain and application layers never import sqlc-generated types.

### 2. Use golang-migrate for Database Schema Migration

- Migration files use sequential versioning with up/down pairs.
- Migrations are embedded in the binary via `embed.FS` + `iofs` source driver for single-binary CLI deployment.
- Migrations run automatically on application startup.
- Schema files also serve as input to sqlc for query validation.

### 3. Three-Tier Testing Strategy

| Tier | Scope | Isolation | Approach |
|---|---|---|---|
| **Repository** | SQL queries, data mapping, constraints | Real SQLite file per test | `t.TempDir()` — each test gets its own temp directory with a fresh SQLite database; migrations run per test |
| **Service** | Business logic, validation, budget math, report computation | Mocked repositories | Mock application output port interfaces; verify orchestration and domain rules |
| **Command** | Full CLI command flow including argument parsing and output rendering | Real binary execution | Execute CLI binary with arguments; assert exit codes, stdout, stderr |

### 4. SQLite Driver

- Use `mattn/go-sqlite3` as the SQLite driver for consistency with golang-migrate's `database/sqlite3` driver.
- sqlc generates code against `database/sql` interface; driver choice is a runtime concern.
- Build requires CGO (C compiler). Pre-built binaries distributed to end users.

## Options Considered

### SQL Access Layer

1. **Hand-written `database/sql` queries** — no compile-time safety, high boilerplate.
2. **ORM (GORM, Ent)** — abstraction leak, complex API, hides SQL behavior.
3. **Query builder (squirrel, goqu)** — runtime query construction, no compile-time SQL validation.
4. **sqlc** (chosen) — write SQL, get type-safe Go. Zero runtime overhead. Compile-time query validation.

### Migration Tool

1. **Manual DDL execution** — no version tracking, no rollback.
2. **Goose** — alternative migration tool; similar features but less ecosystem adoption.
3. **golang-migrate** (chosen) — mature, embedded migration support via `iofs`, CLI and programmatic API, wide community adoption.

### Testing Strategy for SQLite

1. **testcontainers-go** — designed for containerized services (Postgres, MySQL, Redis). No SQLite module exists. SQLite is embedded and in-process; containerizing it adds Docker dependency with zero benefit.
2. **In-memory SQLite (`:memory:`)** — fast but does not test file-based behavior (WAL, locking). Private per-connection, making connection sharing tricky.
3. **`t.TempDir()` with file-based SQLite** (chosen) — each test gets a real file-based SQLite database in an auto-cleaned temp directory. Tests real code paths. Supports `t.Parallel()`. Zero external dependencies.

## Rationale

- **sqlc** eliminates an entire class of runtime SQL errors by catching them at code generation time. The `Querier` interface (via `emit_interface: true`) enables clean mocking in service tests. Raw SQL is fully visible and auditable.
- **golang-migrate** is the most widely adopted Go migration library. `iofs` + `embed.FS` support makes it ideal for single-binary CLI distribution. Sequential versioning is appropriate for a single-developer MVP.
- **`t.TempDir()`** is the idiomatic Go pattern for SQLite testing. It provides real file-based isolation, parallel-safe, zero dependencies, and automatic cleanup. Testcontainers is documented for future use if/when server databases are added.

## Consequences

### Positive
- Compile-time SQL type safety reduces bugs.
- Migration version tracking ensures reproducible schema state.
- Clear testing boundaries make each tier independently maintainable.
- No Docker dependency for testing in MVP.
- sqlc's `Querier` interface enables clean service-layer mocking.
- Output port interfaces are now explicitly modeled in the architecture diagram as a `PORTS` subgraph, making the hexagonal boundary between application and infrastructure layers visually clear.

### Negative
- sqlc adds a code generation step to the build workflow (`sqlc generate`).
- `mattn/go-sqlite3` requires CGO — build machine needs a C compiler.
- Migration files serve dual purpose (schema evolution + sqlc input); care needed to keep them compatible.

### Operational / Maintenance
- CI pipeline must install sqlc CLI before build.
- CI pipeline must have C compiler for CGO (gcc or equivalent).
- Generated sqlc code should be committed to version control (not generated in CI) for build reproducibility.
- When adding a migration, run `sqlc generate` to update generated code.

## Implementation Notes

### sqlc Configuration Location

`sqlc.yaml` at project root:

```yaml
version: "2"
sql:
  - engine: "sqlite"
    queries: "internal/infrastructure/database/sqlite/queries/"
    schema: "internal/infrastructure/database/sqlite/migrations/"
    gen:
      go:
        package: "sqlc"
        out: "internal/infrastructure/database/sqlite/sqlc"
        sql_package: "database/sql"
        emit_json_tags: true
        emit_interface: true
        emit_pointers_for_null_types: true
        emit_empty_slices: true
        omit_unused_structs: true
```

### Migration File Naming

```
internal/infrastructure/database/sqlite/migrations/
├── 000001_init_schema.up.sql
├── 000001_init_schema.down.sql
```

### Test Helper Pattern

```go
// internal/infrastructure/database/sqlite/testutil_test.go
func newTestDB(t *testing.T) *sql.DB {
    t.Helper()
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "test.db")
    db, err := sql.Open("sqlite3", dbPath)
    // ... run migrations, return db
}
```

### Mocking Pattern for Service Tests

Service tests mock application output port interfaces (e.g., `TransactionRepository`, `CategoryRepository`). sqlc's generated `Querier` interface is internal to the infrastructure layer and not exposed to application/domain.

## Related Documents
- `docs/trd/001-infinita-personal-finance-app-mvp-cli.md`
- `docs/adr/ADR-001-adopt-hexagonal-clean-architecture-go.md`
- `docs/adr/ADR-002-defer-server-separation-to-roadmap.md`
- `docs/prd/001-infinita-personal-finance-app.md`
