# TRD: Infinita Personal Finance App (MVP CLI)

## Metadata
- Document ID: TRD
- Title: Infinita Personal Finance App (MVP CLI)
- Owner: TBD
- Reviewers: TBD
- Status: Draft
- Version: 0.21
- Created date: 2026-03-28
- Last updated date: 2026-03-29
- Related PRD: `docs/prd.md`
- Related ADRs: `docs/adr/ADR-003-adopt-sqlc-golang-migrate-tiered-testing.md`, `docs/adr/ADR-004-adopt-ulid-for-transaction-identifiers.md`, `docs/adr/ADR-005-enable-server-transport-with-embedded-architecture.md`
- Related API Spec: `docs/api/openapi.yaml`

## Context
This TRD translates the PRD into verifiable technical requirements for a CLI-first personal finance MVP focused on fast daily logging, category-based budgeting, simple reporting, optional initial balance, and privacy-first local-only storage.

Per ADR-005, the CLI communicates with business logic through an embedded HTTP server running in the same process. The CLI acts as a thin client making HTTP requests to localhost. All persistence and domain logic reside behind the HTTP API.

Related PRD: `docs/prd.md`
Related API Spec: `docs/api/openapi.yaml`

## Technical Goals
- Provide a deterministic and testable CLI command system for core flows: add, list/filter, budget, and report.
- Ensure secure local-only persistence in MVP, with no telemetry enabled by default.
- Maintain clear English command help, outputs, and errors with consistent terminology.
- Guarantee calculation accuracy for budgets and summaries through automated tests.
- Support optional initial balance as starting point for balance summaries.
- Keep storage behavior explicit in CLI settings with local mode fixed for MVP.
- Embed an HTTP server within the CLI process for clean transport separation (ADR-005).
- Ensure CLI-server communication uses localhost HTTP with JSend-style response format.

## Scope
### In Scope
- CLI command surface and argument validation for transactions, categories, budgets, and reports.
- Local-only persistence layer for transactions, categories, budget limits, initial balance, and app settings.
- Domain services for budget tracking and daily/monthly reporting.
- English-only CLI copy standards and error contract.
- Automated test suite for functional and non-functional acceptance criteria.
- Embedded HTTP server (stdlib net/http) for CLI communication (ADR-005).
- REST API endpoints for all business operations (OpenAPI spec at docs/api/openapi.yaml).
- JSend-style response format with per-field per-rule error codes.
- Server lifecycle management (auto-start/stop with CLI process, health check endpoint).

### Out of Scope
- Bank/e-wallet integrations.
- Full cross-device synchronization workflow.
- GUI applications (web/mobile).
- Investment/tax/bookkeeping advanced features.
- AI insights.
- Transaction edit/delete flows for MVP.
- Remote server deployment or TLS/mutual authentication for MVP.
- External API exposure beyond localhost for MVP.

## Requirements Mapping
| PRD Requirement | Technical Requirement(s) |
|---|---|
| PRD-FR-001, PRD-FR-002, PRD-FR-003 | TRD-CLI-001, TRD-VAL-001, TRD-DATA-001 |
| PRD-FR-004, PRD-FR-005 | TRD-CLI-002, TRD-DATA-002 |
| PRD-FR-006, PRD-FR-007 | TRD-CLI-003, TRD-API-001 |
| PRD-FR-008, PRD-FR-009 | TRD-CLI-004, TRD-DOM-001, TRD-API-002 |
| PRD-FR-010, PRD-FR-011 | TRD-CLI-005, TRD-DOM-002, TRD-API-003 |
| PRD-FR-012 | TRD-UX-001 |
| PRD-FR-013, PRD-FR-016 | TRD-CLI-006, TRD-DATA-003, TRD-INF-003, TRD-INF-004, TRD-API-004 |
| PRD-FR-014 | TRD-DATA-005 |
| PRD-FR-015 | TRD-CLI-007, TRD-DOM-005, TRD-DATA-004, TRD-API-005 |
| PRD-FR-017 | TRD-SRV-001, TRD-SRV-002, TRD-SRV-003, TRD-SRV-004 |
| PRD-FR-018 | TRD-CLI-011, TRD-CLI-012, TRD-SRV-005 |
| PRD-NFR-001 | TRD-VAL-001 |
| PRD-NFR-002 | TRD-SEC-001, TRD-SEC-002 |
| PRD-NFR-003 | TRD-QA-001 |
| PRD-NFR-004 | TRD-NFR-001 |
| PRD-NFR-005 | TRD-OBS-001, TRD-OBS-002, TRD-OBS-003 |
| PRD-NFR-006 | TRD-SRV-006 |

Traceability note:
- `TRD-ARCH-001` to `TRD-ARCH-007` and `TRD-TECH-001` to `TRD-TECH-007` are cross-cutting implementation constraints derived from architecture decisions (ADR) and stack selection. They are intentionally not modeled as one-to-one mappings to a single PRD requirement.
- `TRD-SRV-001` to `TRD-SRV-010` are derived from ADR-005 (enable server transport with embedded architecture) and apply to the HTTP server implementation.

## Architecture Overview
The MVP is a CLI application with an embedded HTTP server and local-only persistence. The CLI acts as a thin client communicating with the server via localhost HTTP (ADR-005).

This project adopts **Hexagonal (Ports & Adapters)** architecture adapted from `go-clean-arch-poc`:

1. **Domain Layer (Core)**
   - Domain entities/value objects and domain rules.
   - No dependency on transport, DB, or framework packages.

2. **Application Layer (Use Cases + Ports)**
   - Use-case orchestration for transaction, category, budget, report, and settings flows.
   - Input/Output ports (interfaces) as stable contracts.

3. **Transport Layer (Adapters)**
   - **CLI adapter** for command parsing and terminal rendering.
   - **Server adapter** (HTTP handlers) for REST API exposure.
   - CLI communicates with Server adapter via localhost HTTP.

4. **Infrastructure Layer (Output Port Implementations)**
    - SQLite repositories and other runtime adapters (config).
    - Implements application output ports.

5. **Bootstrap / Composition Root**
   - Shared runtime wiring is centralized in an internal bootstrap package.
   - It resolves the data directory, opens the local database, constructs repositories and use cases, and builds the embedded HTTP server used by both `cmd/cli` and `cmd/server`.

6. **Dependency Rule**
    - Dependencies must point inward: `transport -> application -> domain` and `infrastructure -> application/domain`.
    - Core layers (`domain`, `application`) must not import transport/infrastructure packages.

## Directory Structure (Go Clean Architecture Reference)

Required directory layout for this project:

```text
.
├── cmd/
│   ├── cli/                                    # CLI entrypoint (embeds HTTP server)
│   └── server/                                 # Standalone server entrypoint (dev use)
├── internal/
│   ├── bootstrap/                               # Shared composition root/runtime wiring for cmd entrypoints
│   ├── domain/
│   │   ├── entity/                             # Transaction, Category, Budget, Settings, Report
│   │   ├── valueobject/                        # EntryType, CategoryKey, ULID
│   │   └── error/                              # Structured domain errors
│   ├── application/
│   │   ├── port/
│   │   │   ├── input/                          # Use case interfaces
│   │   │   └── output/                         # Repository interfaces
│   │   ├── usecase/                            # Use case implementations
│   │   └── validation/                         # Shared amount/date/month/category/timezone parsing
│   ├── transport/
│   │   ├── cli/                                # CLI commands and presenter (urfave/cli)
│   │   ├── client/                             # HTTP client for CLI-to-server communication
│   │   └── server/                             # HTTP handlers/router/DTOs for REST API
│   ├── infrastructure/
│   │   └── database/
│   │       └── sqlite/
│   │           ├── sqlc/                       # sqlc generated code (committed)
│   │           ├── migrations/                 # golang-migrate files (also sqlc schema input)
│   │           ├── queries/                    # sqlc query definitions
│   │           ├── db.go                       # Connection setup, migration runner
│   │           ├── transaction_repo.go         # Implements output port
│   │           ├── category_repo.go
│   │           ├── budget_repo.go
│   │           ├── setting_repo.go
│   │           ├── initial_balance_repo.go
│   │           ├── values.go                   # Nullable / type conversion helpers
│   │           └── testutil_test.go            # Test helpers (t.TempDir pattern)
│   └── testutil/
│       └── assertdomain/                       # Domain error assertion helpers
├── sqlc.yaml                                   # sqlc configuration
└── docs/
```

Notes:
- Primary entry point is CLI (`cmd/cli`) which embeds an HTTP server goroutine.
- The standalone server entry point (`cmd/server`) is optional and intended for development/testing.
- Shared runtime/bootstrap wiring is factored into `internal/bootstrap` so both entrypoints use the same repository, use-case, and server construction path.
- Storage for MVP remains local SQLite only.
- CLI communicates with business logic via HTTP requests to embedded server (see ADR-005).
- sqlc generated code (`sqlc/`) is committed to version control for build reproducibility.
- Migration files serve dual purpose: golang-migrate for schema evolution and sqlc schema input for query validation.
- sqlc query files (`queries/`) contain raw SQL with `-- name:` annotations; running `sqlc generate` produces output in `sqlc/`.
- Repository implementations wrap sqlc-generated code and map sqlc types to domain types at the infrastructure boundary.

## Technical Diagrams
- Mermaid source: `docs/diagrams/trd/001-trd-system-architecture.mmd`
- Mermaid source: `docs/diagrams/trd/001-trd-sequence-add-transaction.mmd`
- Mermaid source: `docs/diagrams/trd/001-trd-sequence-monthly-report.mmd`
- Mermaid source: `docs/diagrams/trd/001-trd-data-model-erd.mmd`

## Technical Requirements

### Architecture / Layering
- `TRD-ARCH-001`: Golang implementation for current MVP must separate modules/packages for `cli` (command adapter) and `application` (usecase + ports), with `domain` as core business rules.
- `TRD-ARCH-002`: `cli` layer must depend on `application` input ports/interfaces and must not access persistence adapters directly.
- `TRD-ARCH-003`: Business rules must reside in `domain` + `application` layers, and these layers must not import CLI transport packages.
- `TRD-ARCH-004`: Server transport adapter (`internal/transport/server`) must be implemented for HTTP API exposure to CLI.
- `TRD-ARCH-005`: Current MVP package layout must follow Hexagonal layering for implemented parts: `domain`, `application` (usecase + ports), `transport` (cli, client, server), and `infrastructure` (sqlite/config/observability adapters).
- `TRD-ARCH-006`: Dependency direction must be enforced so that `domain` and `application` are framework-agnostic and import no transport/infrastructure implementation packages.
- `TRD-ARCH-007`: CLI must not import infrastructure packages directly. All CLI-to-business-logic communication must go through the HTTP client to the embedded server.

### Server Transport / HTTP API
- `TRD-SRV-001`: The embedded HTTP server must use Go stdlib `net/http` ServeMux with method-based routing (Go 1.22+). No external HTTP framework dependency.
- `TRD-SRV-002`: Server must listen on a random available port on localhost (127.0.0.1) only. Port must not be exposed to external networks.
- `TRD-SRV-003`: Server must auto-start as a goroutine when CLI launches and auto-stop on CLI exit. Graceful shutdown must complete within 5 seconds.
- `TRD-SRV-004`: Server must expose a `/health` endpoint that returns HTTP 200 when ready. CLI must poll this endpoint with timeout before executing commands.
- `TRD-SRV-005`: CLI must communicate with business logic exclusively through HTTP requests to the embedded server. Direct repository access from CLI is forbidden.
- `TRD-SRV-006`: Server startup must complete within 5 seconds to avoid perceptible CLI latency (PRD-NFR-006).
- `TRD-SRV-007`: All HTTP responses must follow JSend-style format: `{status: "success"|"fail"|"error", data?: any, message?: string, code?: string, meta?: object}`.
- `TRD-SRV-008`: Validation errors must return `{status: "fail", data: [ErrorObject]}` where `data` is an array of `{code, message, field?, hint?}` objects.
- `TRD-SRV-009`: Server errors must return `{status: "error", code: string, message: string, meta: null}`.
- `TRD-SRV-010`: API paths must be flat style without version prefix (e.g., `/transactions`, `/categories`, `/budgets`). See OpenAPI spec for complete endpoint list.

### Frontend (CLI)
- `TRD-CLI-001`: The system must provide a command to create transactions with required inputs: `type`, `amount`, `category`, `date`; `description` is optional. CLI must normalize monetary input to `amountMinor` (integer minor units) before sending HTTP request to server. Server-side validation enforces strict rules per TRD-VAL-001.
- `TRD-CLI-002`: The system must provide commands to list default categories and create custom categories.
- `TRD-CLI-003`: The system must provide commands to list transactions and filter by category.
- `TRD-CLI-004`: The system must provide commands to set monthly budget per category and show remaining budget and over-limit status.
- `TRD-CLI-005`: The system must provide commands to render daily summaries with income, expense, and net balance, and monthly summaries with income, expense, net balance, closing balance, and top spending categories.
- `TRD-CLI-006`: The system must provide settings commands to show active storage mode as `local` for MVP.
- `TRD-CLI-007`: The system must provide setup/settings commands to set and reset optional initial balance; if omitted, initial balance defaults to `0`.
- `TRD-CLI-008`: Transaction listing command must support pagination via `limit` and `offset`, with defaults `limit=50`, `offset=0`, and maximum `limit=500`.
- `TRD-CLI-009`: CLI process exit codes must be stable: `0` for success, `2` for validation/domain input errors, and `3` for runtime/storage errors.
- `TRD-CLI-010`: MVP command surface must expose canonical initial-balance management commands under settings: `settings set-initial-balance` to set the value explicitly and `settings reset-initial-balance` to reset the persisted value to `0`.
- `TRD-CLI-011`: CLI must use an HTTP client (`internal/transport/client`) to communicate with the embedded server. All commands must serialize input to JSON, make HTTP requests to localhost, and deserialize responses before rendering output.
- `TRD-CLI-012`: CLI must map HTTP response status codes to exit codes: HTTP 2xx → exit 0, HTTP 4xx → exit 2, HTTP 5xx → exit 3. Validation errors in JSend `fail` response must be formatted for terminal output.
- `TRD-UX-001`: All help text, success output, and validation/runtime errors must be in English and use a single approved terminology glossary.

### Backend / Domain
- `TRD-DOM-001`: Budget computation must calculate remaining amount per category as `monthlyLimitMinor - spentMonthToDateMinor` and mark over-limit when `remainingMinor < 0`.
- `TRD-DOM-002`: Summary computation must aggregate totals by date bucket (day/month). Category aggregation and deterministic ordering rules for top spending categories apply to monthly summaries, with ordering defined as `amountMinor DESC`, then normalized category key (case-insensitive category name) `ASC` as tie-breaker.
- `TRD-DOM-003`: Report totals must be bucket-scoped for the requested `daily|monthly` period: `incomeTotalMinor` and `expenseTotalMinor` include only transactions inside that single bucket; `netBalanceMinor` must equal `incomeTotalMinor - expenseTotalMinor` for the same bucket.
- `TRD-DOM-004`: Daily and monthly bucket boundaries must use `reportTimezone` from settings; MVP bootstrap seeds the default timezone to `Asia/Jakarta`, and it can be changed only via explicit settings command.
- `TRD-DOM-005`: Closing balance at period end must be cumulative and use `initialBalanceMinor + cumulativeIncomeToPeriodEndMinor - cumulativeExpenseToPeriodEndMinor`, where cumulative values are from app start through the end of the requested monthly period bucket.

### Validation
- `TRD-VAL-001`: Transaction validation must reject requests where `amount <= 0`, invalid decimal token format, more than 2 fractional digits, invalid `type`, invalid date format, or unknown category, and must return a structured English error message. Valid inputs must be converted to integer minor units (`amountMinor`) with exact conversion and no implicit rounding.
- `TRD-VAL-002`: Validation and domain errors must follow `{code, message, field?, hint?}` with stable error codes including `INVALID_AMOUNT`, `INVALID_DATE`, `UNKNOWN_CATEGORY`, `INVALID_TYPE`, and `STORAGE_MODE_UNAVAILABLE`.
- `TRD-VAL-003`: All monetary CLI inputs (`add --amount`, `budget set --amount`, `settings set-initial-balance --amount`) must use the same decimal-token validation and exact normalization rules to minor units; implicit rounding is forbidden.

### Data
- `TRD-DATA-001`: Transactions must be persisted with fields: `id` (ULID, 26-char lexicographically sortable string — see ADR-004), `type`, `amountMinor`, `currencyCode`, `categoryId`, `categoryNameSnapshot?`, `date`, `description?`, `createdAt`.
- `TRD-DATA-002`: Categories must support seeded defaults and user-created custom values, with uniqueness enforced case-insensitively using a normalized category key.
- `TRD-DATA-003`: Financial data in MVP must be persisted locally on-device only.
- `TRD-DATA-004`: Initial balance must be stored as `initialBalanceMinor` (integer minor units) with `currencyCode` and audit timestamp.
- `TRD-DATA-005`: The MVP must not implement bank integration data models or bank synchronization connectors.

### API / Service Contracts (Internal)
- `TRD-API-001`: Transaction query contract must support category filter, pagination (`limit`, `offset`), and stable sort by date (desc), then creation timestamp (desc).
- `TRD-API-002`: Budget query contract must return `category`, `currencyCode`, `monthlyLimitMinor`, `spentMonthToDateMinor`, `remainingMinor`, and `isOverLimit`.
- `TRD-API-003`: Report query contract must return `{period, currencyCode, incomeTotalMinor, expenseTotalMinor, netBalanceMinor}` for daily mode, and must additionally return `{closingBalanceMinor, topCategories[]}` for monthly mode.
- `TRD-API-004`: Settings contract is local CLI-side (non-cloud HTTP) and must return `{storageMode, analyticsOptIn, reportTimezone}` with idempotent update operations.
- `TRD-API-005`: Initialization contract must return `{initialBalanceMinor, currencyCode, initializedAt}` and support explicit set/reset by user command. In MVP, reset semantics mean persisting `initialBalanceMinor = 0` rather than deleting the initialization record.

### Infrastructure
- `TRD-INF-001`: Application configuration must define local data directory path and default to a per-user application data location.
- `TRD-INF-002`: The system must initialize default categories during first-run bootstrap idempotently.
- `TRD-INF-003`: MVP runtime configuration must enforce local-only persistence and reject non-local storage mode activation.
- `TRD-INF-004`: Local persistence engine for MVP must use SQLite with foreign-key enforcement enabled and database file stored under per-user application data directory. Schema migration must use golang-migrate with migrations embedded in the CLI binary. See TRD-TECH-006.
- `TRD-INF-005`: sqlc-generated code must be committed to version control. The CI pipeline must verify that committed generated code matches `sqlc generate` output to detect stale generated code.
- `TRD-INF-006`: sqlc configuration (`sqlc.yaml`) must enable `emit_interface` (generates `Querier` interface for testability), `emit_pointers_for_null_types` (cleaner domain type mapping), `emit_empty_slices` (consistent slice returns for empty results), and `omit_unused_structs` (reduce generated code bloat). sqlc schema input must reference the same migration files used by golang-migrate.
- `TRD-QA-001`: The build/test pipeline must include automated execution of core CLI scenarios (`add`, `list`, `budget`, `report`) and publish pass/fail summary for each run.
- `TRD-QA-002`: Repository integration tests must use real SQLite databases created in `t.TempDir()` directories with migrations applied per test. No in-memory databases in repository tests.
- `TRD-QA-003`: Service unit tests must mock output port interfaces only. No database access in service tests. All domain rules and validation logic must have dedicated unit test coverage.
- `TRD-QA-004`: Command integration tests must execute the compiled CLI binary as a subprocess and assert exit codes (`0`, `2`, `3`) and output content. Tests must not call use-case methods directly.

### Technology Stack
- `TRD-TECH-001`: The MVP CLI must be implemented in **Golang**.
- `TRD-TECH-002`: The project must target a stable Go release line (minimum `go 1.22`) and lock module dependencies via `go.mod`/`go.sum`. Go 1.22+ is required for ServeMux method-based routing.
- `TRD-TECH-003`: Package layout must follow hexagonal layering: `domain`, `application` (usecase + ports), `transport` (cli, client, server), and `infrastructure` (sqlite/config/observability adapters).
- `TRD-TECH-004`: HTTP server must use Go stdlib `net/http` ServeMux with method-based routing (Go 1.22+). No external HTTP framework dependency per ADR-005.
- `TRD-TECH-005`: SQL-to-Go code generation must use **sqlc** (sqlc.dev). Raw SQL queries are written in `.sql` files with `-- name:` annotations; sqlc generates type-safe Go structs, methods, and an optional `Querier` interface. Generated code must reside in the infrastructure layer and must never be imported by domain or application layers. Repository implementations wrap sqlc-generated code and perform type mapping at the infrastructure boundary. See ADR-003.
- `TRD-TECH-006`: Database schema migration must use **golang-migrate** (`github.com/golang-migrate/migrate/v4`). Migration files use sequential versioning with up/down pairs. Migrations must be embedded in the CLI binary via `embed.FS` + `iofs` source driver for single-binary distribution. Migrations must run automatically on application startup. See ADR-003.
- `TRD-TECH-007`: SQLite driver must be `mattn/go-sqlite3` for consistency with golang-migrate's `database/sqlite3` driver. Build requires CGO (C compiler). sqlc generates code against the `database/sql` interface; driver choice is a runtime concern.

### Security
- `TRD-SEC-001`: Data files or database artifacts must be created with user-only read/write permissions where supported by OS.
- `TRD-SEC-002`: The project must document local data handling rules, including data path, permission model, backup guidance, and manual deletion steps. See [`docs/security/local-data.md`](../security/local-data.md).

## Non-Functional Requirements
- `TRD-NFR-001` (Performance): For familiar users, transaction entry flow (command issue to success response) must complete within 15 seconds in benchmark profile `MVP-CLI-BENCH-01`.
- `TRD-NFR-002` (Reliability): Core CLI scenarios (`add`, `list`, `budget`, `report`) must pass with at least 99% success across a minimum 1,000 automated executions.
- `TRD-NFR-003` (Scalability): Listing and reporting must remain functionally correct for datasets up to 100,000 local transactions.
- `TRD-OBS-001` (Observability): No telemetry events are emitted by the application.
- `TRD-NFR-004` (Compliance/Privacy): All user data remains local on the device with no remote transmission.
- `TRD-NFR-005` (Mode Transparency): CLI settings output must always show active storage mode as `local` in MVP.

## API / Data Contracts

### CLI Command Surface (MVP Canonical)
- `infinita add --type <income|expense> --amount <decimal> --category <name> --date <YYYY-MM-DD> [--description <text>]`
- `infinita list [--category <name>] [--limit <n>] [--offset <n>]`
- `infinita category list`
- `infinita category create --name <name> [--description <text>]`
- `infinita budget set --category <name> --amount <decimal> --month <YYYY-MM>`
- `infinita budget status --month <YYYY-MM>`
- `infinita report daily --date <YYYY-MM-DD>`
- `infinita report monthly --month <YYYY-MM>`
- `infinita settings show`
- `infinita settings report-timezone --timezone <IANA name>`
- `infinita settings set-initial-balance --amount <decimal>`
- `infinita settings reset-initial-balance`

CLI execution behavior:
- `list` command default pagination: `limit=50`, `offset=0`; max `limit=500`.
- Exit codes are fixed: `0` success, `2` validation/domain input error, `3` runtime/storage error.
- All CLI commands that accept monetary input must normalize using one shared parser/normalizer function before calling domain services.

### HTTP API Endpoints
The embedded HTTP server exposes REST endpoints for all business operations. Complete API specification is in `docs/api/openapi.yaml`.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check (returns HTTP 200 when server ready) |
| `/transactions` | POST | Create a transaction |
| `/transactions` | GET | List transactions with filters and pagination |
| `/categories` | GET | List all categories |
| `/categories` | POST | Create a custom category |
| `/budgets` | PUT | Set monthly budget for a category |
| `/budgets/status` | GET | Get budget status for a month |
| `/reports/daily` | GET | Get daily summary |
| `/reports/monthly` | GET | Get monthly summary |
| `/settings` | GET | Get current settings |
| `/settings/report-timezone` | PUT | Update report timezone |
| `/settings/initial-balance` | PUT | Set initial balance |
| `/settings/initial-balance` | DELETE | Reset initial balance to 0 |

Request/Response format:
- All requests use JSON body (POST/PUT).
- All responses follow JSend style: `{status, data?, message?, code?, meta?}`.
- See `docs/api/openapi.yaml` for complete schema definitions.

### Monetary Input Normalization (Shared)
Shared monetary parsing/normalization rules apply to all CLI amount inputs:
- `add --amount`
- `budget set --amount`
- `settings set-initial-balance --amount`

Rules:
- Input token must be a numeric decimal token (digits, optional single `.`, surrounding spaces trimmed by parser).
- Max fractional scale is `2`; values with scale `> 2` must be rejected with `INVALID_AMOUNT`.
- Value must be `> 0` for transaction and budget inputs; initial balance may be `>= 0`.
- Thousand separators and locale-formatted tokens are rejected unless explicitly normalized into valid decimal token first.
- Exact conversion only: `amountMinor = amount * 100`; no floating-point conversion and no implicit rounding.

### Transaction Create Contract (CLI Input)
```json
{
  "type": "income|expense",
  "amount": "10000",
  "currencyCode": "IDR",
  "category": "food",
  "date": "YYYY-MM-DD",
  "description": "optional"
}
```

Boundary rule:
- CLI must accept `amount` as user-entered monetary input and normalize it exactly once before calling internal/domain transaction-create service.
- Monetary values must use integer minor units end-to-end for storage and arithmetic (`INTEGER` in SQLite), not floating-point.
- Normalization rule: `amountMinor = amount * 100` with exact decimal parsing; conversion must fail with `INVALID_AMOUNT` if exact conversion is not possible under max scale `2`.
- MVP currency output remains fixed to `IDR`, and fractional scale up to 2 digits is intentionally allowed for internal consistency and forward-compatible monetary handling.
- Accepted `amount` format is numeric decimal token (digits, optional one decimal separator `.`, optional leading/trailing spaces trimmed by parser).
- Values with more than 2 fractional digits must be rejected with `INVALID_AMOUNT`; they must not be rounded implicitly.
- Thousand separators and locale-formatted number inputs must be rejected unless they are explicitly normalized by parser into a valid decimal token first.

Validation rules (CLI input):
- `amount > 0`
- `amount` must be a valid decimal token with max scale `2`
- `type` must be one of `income`, `expense`
- `date` must be valid ISO date (`YYYY-MM-DD`)
- `category` must exist in default/custom category set

### Transaction Create Contract (Internal Normalized)
```json
{
  "type": "income|expense",
  "amountMinor": 1000000,
  "currencyCode": "IDR",
  "category": "food",
  "date": "YYYY-MM-DD",
  "description": "optional"
}
```

Validation rules (internal normalized):
- `amountMinor > 0`
- `type` must be one of `income`, `expense`
- `date` must be valid ISO date
- `category` must exist in default/custom category set

### Budget Status Contract
```json
{
  "category": "food",
  "currencyCode": "IDR",
  "monthlyLimitMinor": 200000000,
  "spentMonthToDateMinor": 150000000,
  "remainingMinor": 50000000,
  "isOverLimit": false
}
```

### Budget Set Contract (CLI Input)
```json
{
  "category": "food",
  "amount": "2000000.00",
  "month": "YYYY-MM"
}
```

Normalization and validation:
- `amount` must follow shared monetary normalization rules and be converted to `monthlyLimitMinor`.
- `category` must exist in default/custom category set.
- `month` must be valid ISO month (`YYYY-MM`).

### Budget Set Contract (Internal Normalized)
```json
{
  "category": "food",
  "monthlyLimitMinor": 200000000,
  "month": "YYYY-MM"
}
```

### Report Contract

Daily report:
```json
{
  "period": "daily",
  "currencyCode": "IDR",
  "incomeTotalMinor": 50000000,
  "expenseTotalMinor": 32000000,
  "netBalanceMinor": 18000000
}
```

Monthly report:
```json
{
  "period": "monthly",
  "currencyCode": "IDR",
  "incomeTotalMinor": 50000000,
  "expenseTotalMinor": 32000000,
  "netBalanceMinor": 18000000,
  "closingBalanceMinor": 218000000,
  "topCategories": [
    { "category": "food", "amountMinor": 12000000 },
    { "category": "transport", "amountMinor": 6000000 }
  ]
}
```

Report field semantics:
- `incomeTotalMinor`: total income within the requested period bucket only (daily/monthly).
- `expenseTotalMinor`: total expense within the requested period bucket only (daily/monthly).
- `netBalanceMinor`: period net within the requested period bucket and must equal `incomeTotalMinor - expenseTotalMinor`.
- `closingBalanceMinor`: monthly-only field; cumulative balance at month end and must equal `initialBalanceMinor + cumulativeIncomeToPeriodEndMinor - cumulativeExpenseToPeriodEndMinor`.
- `topCategories`: monthly-only field; spending categories ranked within the requested month bucket, ordered by `amountMinor DESC`, then normalized category key (case-insensitive category name) `ASC` (as defined in `TRD-DOM-002`).

### Storage Settings Contract
```json
{
  "storageMode": "local",
  "analyticsOptIn": false,
  "reportTimezone": "Asia/Jakarta"
}
```

### Initial Balance Contract
```json
{
  "initialBalanceMinor": 0,
  "currencyCode": "IDR",
  "initializedAt": "2026-03-28T09:00:00Z"
}
```

### Initial Balance Set Contract (CLI Input)
```json
{
  "amount": "500000.00"
}
```

Normalization and validation:
- `amount` must follow shared monetary normalization rules and be converted to `initialBalanceMinor`.
- Initial balance allows zero (`>= 0`) and rejects negative values.

## Dependencies
- Golang toolchain (`go 1.22` or newer for ServeMux method-based routing).
- Selected Go CLI library (or standard library `flag`) for command parsing and help output.
- Go stdlib `net/http` for embedded HTTP server (no external framework per ADR-005).
- SQLite embedded database engine (`mattn/go-sqlite3`, CGO).
- sqlc CLI tool (`github.com/sqlc-dev/sqlc`) for SQL-to-Go code generation.
- golang-migrate (`github.com/golang-migrate/migrate/v4`) for schema migration with `database/sqlite3` and `source/iofs` drivers.
- Date/time utility with deterministic timezone handling.
- ULID library (`github.com/oklog/ulid/v2`) for sortable, collision-resistant transaction identifiers (see ADR-004).
- Test framework supporting CLI integration tests and scenario replay.
- Mock generation tool for service unit tests (e.g., `gomock` or `testify/mock`).

## Technical Risks
- SQLite schema migration/versioning risk as requirements evolve beyond MVP.
- Timezone/date-boundary risk for daily and monthly aggregation.
- Category normalization risk (case/whitespace mismatches) affecting filter and budget accuracy.
- Permission portability risk across operating systems.
- Monetary precision risk if any implementation path bypasses integer minor-unit arithmetic and uses floating-point.
- Embedded server lifecycle risk: startup latency, health check timeout, graceful shutdown handling.
- HTTP overhead risk for local-only communication (mitigated by localhost-only binding).

## Testing Strategy

The MVP testing strategy follows a four-tier approach aligned with hexagonal architecture layers. Each tier has distinct scope, isolation boundaries, and tooling. See ADR-003 for rationale.

### Tier 1: Repository Integration Tests

- **Scope**: SQL queries, data mapping (sqlc types ↔ domain types), database constraints, migration correctness.
- **Isolation**: Real SQLite file per test via `t.TempDir()`. Each test gets its own temp directory with a fresh database; migrations run per test.
- **Location**: `internal/infrastructure/database/sqlite/*_test.go`
- **Pattern**: Shared test helper (`testutil_test.go`) creates a migrated database and returns `*sql.DB`. Tests call repository methods and assert persisted state.
- **Parallel-safe**: Each test uses an independent database file; supports `t.Parallel()`.
- **Note**: SQLite is an embedded, file-based database. It does not require containerized instances for testing. The `t.TempDir()` pattern provides real file-based isolation with zero external dependencies and automatic cleanup.

### Tier 2: Service Unit Tests (Mocked Repositories)

- **Scope**: Business logic, validation rules, budget computation, report aggregation, domain rules.
- **Isolation**: Mock application output port interfaces (e.g., `TransactionRepository`, `CategoryRepository`, `BudgetRepository`). No database access.
- **Location**: `internal/application/usecase/*_test.go`
- **Pattern**: Use generated mocks of output port interfaces. Verify correct orchestration, domain rule enforcement, and error handling. Assert interactions (calls, arguments) on mocks.
- **Key focus**: Validation logic (TRD-VAL-001 to TRD-VAL-003), budget math (TRD-DOM-001), summary computation (TRD-DOM-002, TRD-DOM-003), closing balance (TRD-DOM-005).

### Tier 2.5: HTTP Handler Tests

- **Scope**: HTTP request/response handling, routing, JSend response format, error mapping, DTO serialization.
- **Isolation**: Use `httptest` recorder with real handler functions. Mock application input port interfaces (use cases). No database access.
- **Location**: `internal/transport/server/*_test.go`
- **Pattern**: Use `httptest.NewRequest` and `httptest.NewRecorder` to test handlers directly. Mock use case interfaces to verify correct method calls with expected DTOs. Assert HTTP status codes, response body structure, and JSend format compliance.
- **Key focus**: HTTP method routing, endpoint paths, request parsing, response shaping (TRD-SRV-007 to TRD-SRV-010), validation error formatting (TRD-VAL-002).

### Tier 3: Command Integration Tests (Real CLI Execution)

- **Scope**: Full CLI command flow — argument parsing, validation, service orchestration, output rendering, exit codes.
- **Isolation**: Execute CLI binary as a subprocess with real arguments against a real (temporary) database.
- **Location**: `internal/transport/cli/*_test.go` or `test/integration/`
- **Pattern**: Build CLI binary once per test session. Each test creates a temp directory with database, executes binary with arguments, asserts exit code (0/2/3 per TRD-CLI-009), stdout content, and stderr content.
- **Key focus**: CLI surface contracts (TRD-CLI-001 to TRD-CLI-010), exit codes, English output, error formatting.

### Additional Test Categories

Current implementation note:
- The automated suite currently covers repository integration tests, service unit tests, HTTP handler tests, and command integration tests.
- Reliability runs (≥ 1,000 executions), performance benchmarks, and very-large-dataset validation are planned hardening work and may be completed after the baseline MVP feature flows are stabilized.

- Contract tests for internal service responses (`TRD-API-001` to `TRD-API-003`).
- Contract tests for settings and initialization contracts (`TRD-API-004`, `TRD-API-005`).
- Contract tests for HTTP API endpoints (`TRD-SRV-007` to `TRD-SRV-010`) matching OpenAPI spec.
- Contract tests for error schema (`TRD-VAL-002`) and analytics payload constraints (`TRD-OBS-003`).
- Server lifecycle tests: startup, health check poll, graceful shutdown (`TRD-SRV-003`, `TRD-SRV-004`).
- Performance test scenarios for transaction entry time and report generation.
- Reliability scenario runner with ≥ 1,000 executions for core flows (`add`, `list`, `budget`, `report`) per `TRD-NFR-002`.
- Security checks for local permission mode and telemetry-off default behavior.

## Rollout Plan
1. Implement command flows behind stable command names for MVP.
2. Ship with default categories bootstrap, initial balance setup flow, and local storage initialization.
3. Execute acceptance test matrix and reliability/performance/security checks.
4. Release MVP as a local CLI package/binary.

## Rollback Plan
- Keep data schema versioning for local storage artifacts and local settings contracts.
- If release fails, revert to previous CLI build while preserving local data and local settings.
- Provide migration rollback script/process if schema migration is introduced.

## Open Questions
- Should custom categories be creatable implicitly during `add` command or only via explicit category command?
- Should category and date filtering be expanded beyond current PRD MVP scope in initial implementation?

## References
- `docs/prd.md`
- `docs/diagrams/trd/001-trd-system-architecture.mmd`
- `docs/diagrams/trd/001-trd-sequence-add-transaction.mmd`
- `docs/diagrams/trd/001-trd-sequence-monthly-report.mmd`
- `docs/diagrams/trd/001-trd-data-model-erd.mmd`
- `docs/adr/ADR-002-defer-server-separation-to-roadmap.md`
- `docs/adr/ADR-003-adopt-sqlc-golang-migrate-tiered-testing.md`
- `https://raw.githubusercontent.com/handiism/go-clean-arch-poc/8a94e96666ed9715cabaa46ede768163a86ebe6a/README.md`
