package entity

type DailySummary struct {
	Period            string
	CurrencyCode      string
	IncomeTotalMinor  int64
	ExpenseTotalMinor int64
	NetBalanceMinor   int64
}

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
