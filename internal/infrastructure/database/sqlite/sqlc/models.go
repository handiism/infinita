package sqlc

import "database/sql"

type Category struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	NormalizedKey string `json:"normalized_key"`
	Description   string `json:"description"`
	CreatedAt     string `json:"created_at"`
}

type Transaction struct {
	ID                   string         `json:"id"`
	Type                 string         `json:"type"`
	AmountMinor          int64          `json:"amount_minor"`
	CurrencyCode         string         `json:"currency_code"`
	CategoryID           int64          `json:"category_id"`
	CategoryNameSnapshot string         `json:"category_name_snapshot"`
	Date                 string         `json:"date"`
	Description          sql.NullString `json:"description"`
	CreatedAt            string         `json:"created_at"`
}

type TransactionTotals struct {
	IncomeTotalMinor  int64 `json:"income_total_minor"`
	ExpenseTotalMinor int64 `json:"expense_total_minor"`
}

type TopCategory struct {
	CategoryName string `json:"category_name"`
	CategoryKey  string `json:"category_key"`
	AmountMinor  int64  `json:"amount_minor"`
}

type BudgetStatus struct {
	CategoryName          string `json:"category_name"`
	CategoryKey           string `json:"category_key"`
	Month                 string `json:"month"`
	MonthlyLimitMinor     int64  `json:"monthly_limit_minor"`
	SpentMonthToDateMinor int64  `json:"spent_month_to_date_minor"`
}

type Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AnalyticsConsent struct {
	ID             int64 `json:"id"`
	AnalyticsOptIn int   `json:"analytics_opt_in"`
}

type InitialBalance struct {
	ID                  int64  `json:"id"`
	InitialBalanceMinor int64  `json:"initial_balance_minor"`
	CurrencyCode        string `json:"currency_code"`
	InitializedAt       string `json:"initialized_at"`
}

type CreateTransactionParams struct {
	ID                   string
	Type                 string
	AmountMinor          int64
	CurrencyCode         string
	CategoryID           int64
	CategoryNameSnapshot string
	Date                 string
	Description          *string
}

type UpsertBudgetParams struct {
	CategoryID        int64
	Month             string
	MonthlyLimitMinor int64
}

type UpsertSettingParams struct {
	Key   string
	Value string
}

type UpsertInitialBalanceParams struct {
	InitialBalanceMinor int64
	CurrencyCode        string
}
