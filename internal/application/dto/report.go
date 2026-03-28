package dto

type DailyReportResponse struct {
	Period            string
	CurrencyCode      string
	IncomeTotalMinor  int64
	ExpenseTotalMinor int64
	NetBalanceMinor   int64
}

type MonthlyReportResponse struct {
	Period              string
	CurrencyCode        string
	IncomeTotalMinor    int64
	ExpenseTotalMinor   int64
	NetBalanceMinor     int64
	ClosingBalanceMinor int64
	TopCategories       []ReportTopCategory
}

type ReportTopCategory struct {
	Category    string
	AmountMinor int64
}
