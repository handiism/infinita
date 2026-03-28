package dto

type BudgetSetRequest struct {
	Category string
	Month    string
	Limit    int64
}

type BudgetStatusResponse struct {
	CategoryName          string
	CategoryKey           string
	Month                 string
	MonthlyLimitMinor     int64
	SpentMonthToDateMinor int64
	RemainingMinor        int64
	IsOverLimit           bool
}
