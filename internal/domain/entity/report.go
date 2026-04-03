package entity

// DailySummary holds aggregated transaction data for a single day.
type DailySummary struct {
	Period            string
	CurrencyCode      string
	IncomeTotalMinor  int64
	ExpenseTotalMinor int64
	NetBalanceMinor   int64
}

// TopSpendingCategory represents a category with the highest spending in a period.
type TopSpendingCategory struct {
	Category    string
	AmountMinor int64
}

type MonthlySummary struct {
	Period              string
	CurrencyCode        string
	IncomeTotalMinor    int64
	ExpenseTotalMinor   int64
	NetBalanceMinor     int64
	ClosingBalanceMinor int64
	TopCategories       []TopSpendingCategory
}
