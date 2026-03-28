package sqlc

import (
	"context"
	"database/sql"
)

type Querier interface {
	CreateCategory(context.Context, string, string, string) error
	GetCategoryByNormalizedKey(context.Context, string) (Category, error)
	ListCategories(context.Context) ([]Category, error)
	CreateTransaction(context.Context, CreateTransactionParams) error
	ListTransactions(context.Context, sql.NullString, int32, int32) ([]Transaction, error)
	SumTransactionTotalsForDay(context.Context, string) (TransactionTotals, error)
	SumTransactionTotalsForMonth(context.Context, string) (TransactionTotals, error)
	SumCumulativeTotalsUpToDate(context.Context, string) (TransactionTotals, error)
	TopCategoriesForMonth(context.Context, string) ([]TopCategory, error)
	UpsertBudget(context.Context, UpsertBudgetParams) error
	GetBudgetsForMonth(context.Context, string) ([]BudgetStatus, error)
	GetSetting(context.Context, string) (Setting, error)
	UpsertSetting(context.Context, UpsertSettingParams) error
	GetAnalyticsConsent(context.Context) (AnalyticsConsent, error)
	SetAnalyticsConsent(context.Context, bool) error
	GetInitialBalance(context.Context) (InitialBalance, error)
	UpsertInitialBalance(context.Context, UpsertInitialBalanceParams) error
}

type Queries struct {
	db *sql.DB
}

func New(db *sql.DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) CreateCategory(ctx context.Context, name, normalizedKey, description string) error {
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO categories (name, normalized_key, description) VALUES (?, ?, ?);`,
		name, normalizedKey, description)
	return err
}

func (q *Queries) GetCategoryByNormalizedKey(ctx context.Context, key string) (Category, error) {
	var category Category
	err := q.db.QueryRowContext(ctx,
		`SELECT id, name, normalized_key, description, created_at FROM categories WHERE normalized_key = ?;`,
		key).Scan(&category.ID, &category.Name, &category.NormalizedKey, &category.Description, &category.CreatedAt)
	if err != nil {
		return Category{}, err
	}
	return category, nil
}

func (q *Queries) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, name, normalized_key, description, created_at FROM categories ORDER BY name COLLATE NOCASE ASC;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	categories := make([]Category, 0)
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.NormalizedKey, &category.Description, &category.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (q *Queries) CreateTransaction(ctx context.Context, arg CreateTransactionParams) error {
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO transactions (id, type, amount_minor, currency_code, category_id, category_name_snapshot, date, description)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?);`,
		arg.ID, arg.Type, arg.AmountMinor, arg.CurrencyCode, arg.CategoryID, arg.CategoryNameSnapshot, arg.Date, arg.Description)
	return err
}

func (q *Queries) ListTransactions(ctx context.Context, categoryKey sql.NullString, limit int32, offset int32) ([]Transaction, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT t.id, t.type, t.amount_minor, t.currency_code, t.category_id, t.category_name_snapshot, t.date, t.description, t.created_at
        FROM transactions t
        JOIN categories c ON c.id = t.category_id
        WHERE (?1 IS NULL OR c.normalized_key = ?1)
        ORDER BY t.date DESC, t.created_at DESC
        LIMIT ?2 OFFSET ?3;`,
		nullableStringArg(categoryKey), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Transaction
	for rows.Next() {
		var txn Transaction
		if err := rows.Scan(&txn.ID, &txn.Type, &txn.AmountMinor, &txn.CurrencyCode, &txn.CategoryID, &txn.CategoryNameSnapshot, &txn.Date, &txn.Description, &txn.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, txn)
	}
	return items, rows.Err()
}

func (q *Queries) SumTransactionTotalsForDay(ctx context.Context, date string) (TransactionTotals, error) {
	var totals TransactionTotals
	err := q.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(CASE WHEN type = 'income' THEN amount_minor ELSE 0 END), 0) AS income_total_minor,
                COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_minor ELSE 0 END), 0) AS expense_total_minor
         FROM transactions
         WHERE date = ?;`,
		date).Scan(&totals.IncomeTotalMinor, &totals.ExpenseTotalMinor)
	if err != nil {
		return TransactionTotals{}, err
	}
	return totals, nil
}

func (q *Queries) SumTransactionTotalsForMonth(ctx context.Context, month string) (TransactionTotals, error) {
	var totals TransactionTotals
	err := q.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(CASE WHEN type = 'income' THEN amount_minor ELSE 0 END), 0) AS income_total_minor,
                COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_minor ELSE 0 END), 0) AS expense_total_minor
         FROM transactions
         WHERE substr(date, 1, 7) = ?;`,
		month).Scan(&totals.IncomeTotalMinor, &totals.ExpenseTotalMinor)
	if err != nil {
		return TransactionTotals{}, err
	}
	return totals, nil
}

func (q *Queries) SumCumulativeTotalsUpToDate(ctx context.Context, date string) (TransactionTotals, error) {
	var totals TransactionTotals
	err := q.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(CASE WHEN type = 'income' THEN amount_minor ELSE 0 END), 0) AS income_total_minor,
                COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_minor ELSE 0 END), 0) AS expense_total_minor
         FROM transactions
         WHERE date <= ?;`,
		date).Scan(&totals.IncomeTotalMinor, &totals.ExpenseTotalMinor)
	if err != nil {
		return TransactionTotals{}, err
	}
	return totals, nil
}

func (q *Queries) TopCategoriesForMonth(ctx context.Context, month string) ([]TopCategory, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT c.name AS category_name, c.normalized_key AS category_key, COALESCE(SUM(t.amount_minor), 0) AS amount_minor
         FROM categories c
         LEFT JOIN transactions t ON t.category_id = c.id AND t.type = 'expense' AND substr(t.date, 1, 7) = ?
         GROUP BY c.id
         HAVING SUM(t.amount_minor) > 0
         ORDER BY amount_minor DESC, category_key ASC;`,
		month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TopCategory
	for rows.Next() {
		var item TopCategory
		if err := rows.Scan(&item.CategoryName, &item.CategoryKey, &item.AmountMinor); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (q *Queries) UpsertBudget(ctx context.Context, arg UpsertBudgetParams) error {
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO budgets (category_id, month, monthly_limit_minor, updated_at)
        VALUES (?, ?, ?, datetime('now'))
        ON CONFLICT(category_id, month) DO UPDATE SET
            monthly_limit_minor = excluded.monthly_limit_minor,
            updated_at = datetime('now');`,
		arg.CategoryID, arg.Month, arg.MonthlyLimitMinor)
	return err
}

func (q *Queries) GetBudgetsForMonth(ctx context.Context, month string) ([]BudgetStatus, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT c.name AS category_name, c.normalized_key AS category_key, b.month AS month, b.monthly_limit_minor AS monthly_limit_minor,
                COALESCE((
                    SELECT SUM(amount_minor)
                    FROM transactions t
                    WHERE t.category_id = b.category_id
                      AND t.type = 'expense'
                      AND substr(t.date, 1, 7) = b.month
                ), 0) AS spent_month_to_date_minor
         FROM budgets b
         JOIN categories c ON c.id = b.category_id
         WHERE b.month = ?
         ORDER BY c.name COLLATE NOCASE ASC;`,
		month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BudgetStatus
	for rows.Next() {
		var item BudgetStatus
		if err := rows.Scan(&item.CategoryName, &item.CategoryKey, &item.Month, &item.MonthlyLimitMinor, &item.SpentMonthToDateMinor); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (q *Queries) GetSetting(ctx context.Context, key string) (Setting, error) {
	var setting Setting
	err := q.db.QueryRowContext(ctx,
		`SELECT key, value FROM settings WHERE key = ?;`,
		key).Scan(&setting.Key, &setting.Value)
	if err != nil {
		return Setting{}, err
	}
	return setting, nil
}

func (q *Queries) UpsertSetting(ctx context.Context, arg UpsertSettingParams) error {
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO settings (key, value) VALUES (?, ?)
        ON CONFLICT(key) DO UPDATE SET value = excluded.value;`,
		arg.Key, arg.Value)
	return err
}

func (q *Queries) GetAnalyticsConsent(ctx context.Context) (AnalyticsConsent, error) {
	var consent AnalyticsConsent
	err := q.db.QueryRowContext(ctx,
		`SELECT id, analytics_opt_in FROM analytics_consent WHERE id = 1;`).Scan(&consent.ID, &consent.AnalyticsOptIn)
	return consent, err
}

func (q *Queries) SetAnalyticsConsent(ctx context.Context, optIn bool) error {
	value := 0
	if optIn {
		value = 1
	}
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO analytics_consent (id, analytics_opt_in) VALUES (1, ?)
        ON CONFLICT(id) DO UPDATE SET analytics_opt_in = excluded.analytics_opt_in;`,
		value)
	return err
}

func (q *Queries) GetInitialBalance(ctx context.Context) (InitialBalance, error) {
	var initial InitialBalance
	err := q.db.QueryRowContext(ctx,
		`SELECT id, initial_balance_minor, currency_code, initialized_at FROM initial_balance WHERE id = 1;`).Scan(&initial.ID, &initial.InitialBalanceMinor, &initial.CurrencyCode, &initial.InitializedAt)
	return initial, err
}

func (q *Queries) UpsertInitialBalance(ctx context.Context, arg UpsertInitialBalanceParams) error {
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO initial_balance (id, initial_balance_minor, currency_code, initialized_at)
        VALUES (1, ?, ?, datetime('now'))
        ON CONFLICT(id) DO UPDATE SET
            initial_balance_minor = excluded.initial_balance_minor,
            currency_code = excluded.currency_code,
            initialized_at = datetime('now');`,
		arg.InitialBalanceMinor, arg.CurrencyCode)
	return err
}

func nullableStringArg(value sql.NullString) interface{} {
	if value.Valid {
		return value.String
	}
	return nil
}
