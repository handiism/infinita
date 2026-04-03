# PRD: Infinita Personal Finance App (MVP CLI)

## Metadata
- Document ID: PRD
- Title: Infinita Personal Finance App (MVP CLI)
- Owner:
- Reviewers:
- Status: Draft
- Version: 0.9
- Created date: 2026-03-28
- Last updated date: 2026-03-29
- Related Documents:
  - `docs/trd.md`
  - `docs/adr/ADR-002-defer-server-separation-to-roadmap.md`
  - `docs/adr/ADR-005-enable-server-transport-with-embedded-architecture.md`
  - `docs/api/openapi.yaml`

## Problem Statement
Many individuals struggle to consistently track income, expenses, budgets, and financial summaries because existing tools are either too complex or not private enough.

The users most affected are general individual users, budget-conscious users, and privacy-focused users who want lightweight personal finance tracking without accounting-heavy workflows or third-party data sharing.

This matters now because the MVP needs a simple, private, English-first workflow that lowers the barrier to building a consistent financial tracking habit.

## Goals
- Enable fast daily transaction logging through a CLI-first MVP.
- Help users understand spending through categories, budgets, and simple summaries.
- Keep MVP storage local-only on user device with no telemetry by default.
- Provide a clear English-only CLI experience with consistent terminology.

## Non-Goals
- Automated bank or e-wallet integrations.
- Investment, tax, or advanced bookkeeping features.
- AI-based spending insights in MVP.
- GUI web or mobile dashboard in MVP.
- Full cross-device synchronization in MVP.

## Functional Requirements
- `PRD-FR-001`: Users can add income transactions through the CLI.
- `PRD-FR-002`: Users can add expense transactions through the CLI.
- `PRD-FR-003`: The system requires transaction type, amount, category, and date for each transaction; description is optional. In MVP, currency is fixed to `IDR` and amount input supports up to 2 fractional digits.
- `PRD-FR-004`: The system provides default transaction categories.
- `PRD-FR-005`: Users can create and use custom transaction categories.
- `PRD-FR-006`: Users can list stored transactions.
- `PRD-FR-007`: Users can filter listed transactions by category.
- `PRD-FR-008`: Users can set monthly budget limits by category.
- `PRD-FR-009`: Users can view remaining budget and over-limit warnings by category.
- `PRD-FR-010`: Users can view daily summaries with period totals: income, expense, and net balance for the requested day only.
- `PRD-FR-011`: Users can view monthly summaries with period totals: income, expense, net balance, closing balance, and top spending categories for the requested month only.
- `PRD-FR-012`: CLI help text, command output, and error messages are presented in English using consistent terminology.
- `PRD-FR-013`: User financial data is stored locally on the user device in MVP.
- `PRD-FR-014`: The MVP does not require bank integration.
- `PRD-FR-015`: Users can set an optional initial balance, reset it back to `0` when needed, and the active value is used as the starting point for cumulative closing-balance calculations shown in summaries.
- `PRD-FR-016`: CLI settings must clearly show that active storage mode in MVP is local.
- `PRD-FR-017`: The system must provide an embedded HTTP server that auto-starts and auto-stops with the CLI process.
- `PRD-FR-018`: The CLI must communicate with business logic exclusively through the embedded server REST API.

## Non-Functional Requirements
- `PRD-NFR-001`: The system validates that transaction amount is greater than 0, supports up to 2 fractional digits, and rejects invalid input.
- `PRD-NFR-002`: Sensitive config and user data use secure local file permissions and documented local data handling rules.
- `PRD-NFR-003`: Core CLI flows represented by example commands such as `add`, `list`, `budget`, and `report` achieve at least a 99% success rate in defined MVP test scenarios.
- `PRD-NFR-004`: New transaction entry completes within 15 seconds for familiar users.
- `PRD-NFR-005`: No telemetry is enabled by default in MVP.
- `PRD-NFR-006`: Embedded server startup must complete within 5 seconds to avoid perceptible CLI latency.

## Measurement & Privacy Strategy
- All product metrics are evaluated using local-only data; no remote analytics are collected.
- No telemetry or analytics events are emitted by the application.
- All user data remains local on the device with secure file permissions.

## Success Metrics
- At least 60% of first-week active users log at least 5 transactions.
- At least 40% of users who create budgets use the budgeting feature at least once per week.
- At least 99% of MVP command execution scenarios covering core CLI flows complete without runtime errors across at least 1,000 executions.
- New transaction entry time is 15 seconds or less for familiar users.
- 100% of users use local-only storage in MVP.

## Acceptance Criteria
- [ ] Given valid input, when the user runs an example transaction-add command, then the transaction is saved and confirmation is shown.
- [ ] Given existing categories, when the user runs an example transaction-listing or filtering command, then category filtering works correctly.
- [ ] Given monthly budgets are set, when the user runs an example budgeting command after expenses increase, then remaining budget and over-limit warnings are shown.
- [ ] Given a reporting flow, when the user runs an example daily summary command, then period totals (income, expense, net balance) are accurate for the requested day bucket.
- [ ] Given a reporting flow, when the user runs an example monthly summary command, then period totals and top spending categories are accurate for the requested month bucket and closing balance is accurate as of month end.
- [ ] Given CLI core flows, when help, output, and errors are rendered, then wording is in English and terminology is consistent.
- [ ] Given default app behavior, when data is persisted, then data stays local with secure file permissions and no telemetry by default.
- [ ] Given invalid transaction input, when the user submits an amount less than or equal to 0 or with more than 2 fractional digits, then the system rejects the input and shows a validation error.
- [ ] Given first-time setup or manual initialization, when the user provides an initial balance value, then the system stores it and uses it as the starting point for closing-balance calculations in summaries; when omitted, the system defaults initial balance to 0.
- [ ] Given an existing initial balance, when the user explicitly resets it, then the stored value becomes 0 and subsequent closing-balance calculations use 0 until another value is set.
- [ ] Given storage settings in MVP, when the user checks storage configuration, then storage mode is shown as local and data is stored only on-device.

## Risks & Dependencies
### Risks
- **CLI adoption risk**: general users may prefer visual or mobile UI.
- **UX copy consistency risk**: inconsistent English terminology may confuse users.
- **Privacy implementation risk**: secure local storage depends on the chosen stack and file handling approach.
- **Reporting consistency risk**: budget and summary calculations require strong test coverage.
- **Scope risk**: MVP complexity can grow if non-MVP features are added prematurely.

### Dependencies
- A CLI-capable application framework or runtime must be selected for implementation.
- Local storage and file-permission handling must support private on-device persistence.
- Budgeting and reporting flows depend on consistent category and transaction data models.

## Open Questions
- Should the MVP support filtering transactions by date in addition to category?
- Should custom categories be created implicitly during transaction entry or through a separate flow?

## References
- `docs/diagrams/prd/001-prd-main-flow.mmd`

## Roadmap
**MVP**
- Add transactions
- List and filter transactions
- Default and custom categories
- Monthly budgeting per category
- Basic daily and monthly cashflow summaries
- English-only command and output experience
- Local-only storage in MVP
- Optional initial balance setup
- Embedded HTTP server (auto-managed by CLI)
- REST API for all business operations (OpenAPI spec)

**v1.1**
- CSV and TXT report export
- Expanded transaction filtering and search
- Logging reminders
- Local backup and restore

**v2.0**
- Cross-device sync
- Web or mobile GUI dashboard
- Automated insights and forecasting
- Shared household budgeting
