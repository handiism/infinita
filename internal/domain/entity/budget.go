package entity

type BudgetStatus struct {
	CategoryName          string
	CategoryKey           string
	Month                 string
	MonthlyLimitMinor     int64
	SpentMonthToDateMinor int64
	RemainingMinor        int64
	IsOverLimit           bool
}
