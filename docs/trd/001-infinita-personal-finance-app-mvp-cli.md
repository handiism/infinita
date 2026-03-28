# TRD-001: Infinita Personal Finance App (MVP CLI)

## Metadata
- Document ID: TRD-001
- Title: Infinita Personal Finance App (MVP CLI)
- Owner:
- Reviewers:
- Status: Draft
- Version: 0.5
- Created date: 2026-03-28
- Last updated date: 2026-03-28
- Related PRD: `docs/prd/001-infinita-personal-finance-app.md`
- Related ADRs:

## Context
This TRD translates PRD-001 into verifiable technical requirements for a CLI-first personal finance MVP focused on fast daily logging, category-based budgeting, simple reporting, optional initial balance, and privacy-first local-only storage.

Related PRD: `docs/prd/001-infinita-personal-finance-app.md`

## Technical Goals
- Provide a deterministic and testable CLI command system for core flows: add, list/filter, budget, and report.
- Ensure secure local-only persistence in MVP, with no telemetry enabled by default.
- Maintain clear English command help, outputs, and errors with consistent terminology.
- Guarantee calculation accuracy for budgets and summaries through automated tests.
- Support optional initial balance as starting point for balance summaries.
- Keep storage behavior explicit in CLI settings with local mode fixed for MVP.

## Scope
### In Scope
- CLI command surface and argument validation for transactions, categories, budgets, and reports.
- Local-only persistence layer for transactions, categories, budget limits, initial balance, and app settings.
- Domain services for budget tracking and daily/monthly reporting.
- English-only CLI copy standards and error contract.
- Automated test suite for functional and non-functional acceptance criteria.

### Out of Scope
- Bank/e-wallet integrations.
- Full cross-device synchronization workflow.
- GUI applications (web/mobile).
- Investment/tax/bookkeeping advanced features.
- AI insights.
- Transaction edit/delete flows for MVP.

## Requirements Mapping
| PRD Requirement | Technical Requirement(s) |
|---|---|
| PRD-FR-001, PRD-FR-002, PRD-FR-003 | TRD-CLI-001, TRD-VAL-001, TRD-DATA-001 |
| PRD-FR-004, PRD-FR-005 | TRD-CLI-002, TRD-DATA-002 |
| PRD-FR-006, PRD-FR-007 | TRD-CLI-003, TRD-API-001 |
| PRD-FR-008, PRD-FR-009 | TRD-CLI-004, TRD-DOM-001, TRD-API-002 |
| PRD-FR-010, PRD-FR-011 | TRD-CLI-005, TRD-DOM-002, TRD-API-003 |
| PRD-FR-012 | TRD-UX-001 |
| PRD-FR-013, PRD-FR-016 | TRD-CLI-006, TRD-DATA-003, TRD-INF-003, TRD-API-004 |
| PRD-FR-014 | TRD-DATA-005 |
| PRD-FR-015 | TRD-CLI-007, TRD-DOM-003, TRD-DOM-005, TRD-DATA-004, TRD-API-005 |
| PRD-NFR-001 | TRD-VAL-001 |
| PRD-NFR-002 | TRD-SEC-001, TRD-SEC-002 |
| PRD-NFR-003 | TRD-QA-001 |
| PRD-NFR-004 | TRD-NFR-001 |
| PRD-NFR-005 | TRD-OBS-001, TRD-OBS-002, TRD-OBS-003 |

## Architecture Overview
The MVP is a CLI application with local-only persistence and clear boundaries:

1. **CLI Layer**
   - Parses commands and flags.
   - Renders English help/output/error messages.
   - Converts input into application requests.

2. **Application/Domain Layer**
   - Transaction service (create/list/filter).
   - Category service (default/custom categories).
   - Budget service (set limits, compute remaining/over-limit status).
   - Report service (daily/monthly summaries, with top spending categories and closing balance in monthly mode).
   - Settings service (storage mode, analytics opt-in, and initialization values).

3. **Persistence Layer**
   - Repository interfaces for transactions, categories, budgets, initial balance, and settings.
   - Local storage adapter (MVP runtime).

4. **Cross-Cutting**
   - Validation module.
   - Time/calendar utility for daily/monthly grouping.
   - Optional analytics module (disabled unless explicit opt-in).

## Technical Requirements

### Frontend (CLI)
- `TRD-CLI-001`: The system must provide a command to create transactions with required inputs: `type`, `amount`, `category`, `date`; `description` is optional. Parsed amount input must be normalized into `amountMinor` before persistence using the strict normalization rules defined in this document (MVP fixed-currency behavior).
- `TRD-CLI-002`: The system must provide commands to list default categories and create custom categories.
- `TRD-CLI-003`: The system must provide commands to list transactions and filter by category.
- `TRD-CLI-004`: The system must provide commands to set monthly budget per category and show remaining budget and over-limit status.
- `TRD-CLI-005`: The system must provide commands to render daily summaries with income, expense, and net balance, and monthly summaries with income, expense, net balance, closing balance, and top spending categories.
- `TRD-CLI-006`: The system must provide settings commands to show active storage mode as `local` for MVP.
- `TRD-CLI-007`: The system must provide setup/settings command to set optional initial balance; if omitted, initial balance defaults to `0`.
- `TRD-UX-001`: All help text, success output, and validation/runtime errors must be in English and use a single approved terminology glossary.

### Backend / Domain
- `TRD-DOM-001`: Budget computation must calculate remaining amount per category as `monthlyLimit - monthToDateExpense` and mark over-limit when remaining `< 0`.
- `TRD-DOM-002`: Summary computation must aggregate totals by date bucket (day/month). Category aggregation and deterministic ordering rules for top spending categories apply to monthly summaries, with ordering defined as `amountMinor DESC`, then normalized category key (case-insensitive category name) `ASC` as tie-breaker.
- `TRD-DOM-003`: Report totals must be bucket-scoped for the requested `daily|monthly` period: `incomeTotalMinor` and `expenseTotalMinor` include only transactions inside that single bucket; `netBalanceMinor` must equal `incomeTotalMinor - expenseTotalMinor` for the same bucket.
- `TRD-DOM-005`: Closing balance at period end must be cumulative and use `initialBalanceMinor + cumulativeIncomeToPeriodEndMinor - cumulativeExpenseToPeriodEndMinor`, where cumulative values are from app start through the end of the requested monthly period bucket.
- `TRD-DOM-004`: Daily and monthly bucket boundaries must use `reportTimezone` from settings; default timezone is captured on first run and can be changed only via explicit settings command.

### Validation
- `TRD-VAL-001`: Transaction validation must reject requests where `amountMinor <= 0`, invalid `type`, invalid date format, or unknown category, and must return a structured English error message.
- `TRD-VAL-002`: Validation and domain errors must follow `{code, message, field?, hint?}` with stable error codes including `INVALID_AMOUNT`, `INVALID_DATE`, `UNKNOWN_CATEGORY`, `INVALID_TYPE`, and `STORAGE_MODE_UNAVAILABLE`.

### Data
- `TRD-DATA-001`: Transactions must be persisted with fields: `id`, `type`, `amountMinor`, `currencyCode`, `categoryId`, `categoryNameSnapshot?`, `date`, `description?`, `createdAt`.
- `TRD-DATA-002`: Categories must support seeded defaults and user-created custom values, with uniqueness enforced case-insensitively using a normalized category key.
- `TRD-DATA-003`: Financial data in MVP must be persisted locally on-device only.
- `TRD-DATA-004`: Initial balance must be stored as `initialBalanceMinor` with `currencyCode` and audit timestamp.
- `TRD-DATA-005`: The MVP must not implement bank integration data models or bank synchronization connectors.

### API / Service Contracts (Internal)
- `TRD-API-001`: Transaction query contract must support category filter and stable sort by date (desc), then creation timestamp (desc).
- `TRD-API-002`: Budget query contract must return `category`, `currencyCode`, `monthlyLimitMinor`, `spentMonthToDateMinor`, `remainingMinor`, and `isOverLimit`.
- `TRD-API-003`: Report query contract must return `{period, currencyCode, incomeTotalMinor, expenseTotalMinor, netBalanceMinor}` for daily mode, and must additionally return `{closingBalanceMinor, topCategories[]}` for monthly mode.
- `TRD-API-004`: Settings contract is local CLI-side (non-cloud HTTP) and must return `{storageMode, analyticsOptIn, reportTimezone}` with idempotent update operations.
- `TRD-API-005`: Initialization contract must return `{initialBalanceMinor, currencyCode, initializedAt}` and support explicit set/reset by user command.

### Infrastructure
- `TRD-INF-001`: Application configuration must define local data directory path and default to a per-user application data location.
- `TRD-INF-002`: The system must initialize default categories during first-run bootstrap idempotently.
- `TRD-INF-003`: MVP runtime configuration must enforce local-only persistence and reject non-local storage mode activation.
- `TRD-QA-001`: The build/test pipeline must include automated execution of core CLI scenarios (`add`, `list`, `budget`, `report`) and publish pass/fail summary for each run.

### Security
- `TRD-SEC-001`: Data files or database artifacts must be created with user-only read/write permissions where supported by OS.
- `TRD-SEC-002`: The project must document local data handling rules, including data path, permission model, backup guidance, and manual deletion steps.

## Non-Functional Requirements
- `TRD-NFR-001` (Performance): For familiar users, transaction entry flow (command issue to success response) must complete within 15 seconds in benchmark profile `MVP-CLI-BENCH-01`.
- `TRD-NFR-002` (Reliability): Core CLI scenarios (`add`, `list`, `budget`, `report`) must pass with at least 99% success across a minimum 1,000 automated executions.
- `TRD-NFR-003` (Scalability): Listing and reporting must remain functionally correct for datasets up to 100,000 local transactions.
- `TRD-OBS-001` (Observability): No telemetry events are emitted by default; analytics emission must be gated behind explicit opt-in setting.
- `TRD-NFR-004` (Compliance/Privacy): Analytics payloads, if enabled, must exclude raw financial content (amount, description, categories, transaction text).
- `TRD-NFR-005` (Mode Transparency): CLI settings output must always show active storage mode as `local` in MVP.
- `TRD-OBS-002` (Consent Control): Users must be able to review and revoke analytics consent at any time via settings command; revocation must stop new analytics emission immediately.
- `TRD-OBS-003` (Analytics Payload Contract): Minimum analytics payload is limited to `{anonymousInstallId, eventType, timestampBucket}` and must not include any financial payload fields.

## API / Data Contracts

### Transaction Create Contract (CLI Input)
```json
{
  "type": "income|expense",
  "amount": 10000,
  "currencyCode": "IDR",
  "category": "food",
  "date": "YYYY-MM-DD",
  "description": "optional"
}
```

Boundary rule:
- CLI must accept `amount` as user-entered monetary input and normalize it into `amountMinor` exactly once before calling internal/domain transaction-create service.
- MVP currency scope is fixed to `IDR` only, with fixed minor-unit scale `0` (no fractional currency unit).
- Accepted `amount` format is whole-number only (digits, optional leading/trailing spaces trimmed by parser).
- Decimal input (for example `100.5`) must be rejected with `INVALID_AMOUNT`; it must not be rounded.
- Thousand separators and locale-formatted number inputs must be rejected unless they are explicitly normalized by parser into a valid whole-number token first.

Validation rules (CLI input):
- `amount > 0`
- `amount` must be a valid whole-number token for MVP IDR fixed-scale rules (no decimal fraction)
- `type` must be one of `income`, `expense`
- `date` must be valid ISO date
- `category` must exist in default/custom category set

### Transaction Create Contract (Internal Normalized)
```json
{
  "type": "income|expense",
  "amountMinor": 10000,
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
  "monthlyLimitMinor": 2000000,
  "spentMonthToDateMinor": 1500000,
  "remainingMinor": 500000,
  "isOverLimit": false
}
```

### Report Contract

Daily report:
```json
{
  "period": "daily",
  "currencyCode": "IDR",
  "incomeTotalMinor": 500000,
  "expenseTotalMinor": 320000,
  "netBalanceMinor": 180000
}
```

Monthly report:
```json
{
  "period": "monthly",
  "currencyCode": "IDR",
  "incomeTotalMinor": 500000,
  "expenseTotalMinor": 320000,
  "netBalanceMinor": 180000,
  "closingBalanceMinor": 2180000,
  "topCategories": [
    { "category": "food", "amountMinor": 120000 },
    { "category": "transport", "amountMinor": 60000 }
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

## Dependencies
- Selected CLI framework/runtime.
- Local storage mechanism (structured file or embedded DB).
- Date/time utility with deterministic timezone handling.
- Test framework supporting CLI integration tests and scenario replay.

## Technical Risks
- Storage format change risk if MVP starts with files and later migrates to DB.
- Timezone/date-boundary risk for daily and monthly aggregation.
- Category normalization risk (case/whitespace mismatches) affecting filter and budget accuracy.
- Permission portability risk across operating systems.
- Monetary precision risk if minor-unit storage is violated in implementation.

## Testing Strategy
- Unit tests for validation, budget math, and summary aggregation.
- Integration tests for CLI command flows and error handling.
- Contract tests for internal service responses (`TRD-API-001` to `TRD-API-003`).
- Contract tests for settings and initialization contracts (`TRD-API-004`, `TRD-API-005`).
- Contract tests for error schema (`TRD-VAL-002`) and analytics payload constraints (`TRD-OBS-003`).
- Performance test scenarios for transaction entry time and report generation.
- Reliability scenario runner with >= 1,000 executions for core flows.
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
- Which local storage approach is final for MVP: structured files, encrypted files, or local embedded DB?
- Should custom categories be creatable implicitly during `add` command or only via explicit category command?
- Should category and date filtering be expanded beyond current PRD MVP scope in initial implementation?

## References
- `docs/prd/001-infinita-personal-finance-app.md`
- `docs/trd/README.md`
