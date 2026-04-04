package server

type TransactionRequest struct {
	Type         string `json:"type"`
	AmountMinor  int64  `json:"amountMinor"`
	CurrencyCode string `json:"currencyCode"`
	Category     string `json:"category"`
	Date         string `json:"date"`
	Description  string `json:"description"`
}

type TransactionResponse struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"`
	AmountMinor          int64  `json:"amountMinor"`
	CurrencyCode         string `json:"currencyCode"`
	CategoryID           int64  `json:"categoryId"`
	CategoryNameSnapshot string `json:"categoryNameSnapshot"`
	Date                 string `json:"date"`
	Description          string `json:"description"`
	CreatedAt            string `json:"createdAt"`
}

type PaginationMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type TransactionListData struct {
	Data []TransactionResponse `json:"data"`
	Meta PaginationMeta        `json:"meta"`
}

type CategoryListData struct {
	Data []CategoryResponse `json:"data"`
}

type BudgetStatusListData struct {
	Data []BudgetStatusResponse `json:"data"`
}

type CategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CategoryResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type BudgetRequest struct {
	Category          string `json:"category"`
	Month             string `json:"month"`
	MonthlyLimitMinor int64  `json:"monthlyLimitMinor"`
}

type BudgetStatusResponse struct {
	CategoryName          string `json:"categoryName"`
	CurrencyCode          string `json:"currencyCode"`
	MonthlyLimitMinor     int64  `json:"monthlyLimitMinor"`
	SpentMonthToDateMinor int64  `json:"spentMonthToDateMinor"`
	RemainingMinor        int64  `json:"remainingMinor"`
	IsOverLimit           bool   `json:"isOverLimit"`
}

type TopCategoryResponse struct {
	Category    string `json:"category"`
	AmountMinor int64  `json:"amountMinor"`
}

type DailyReportResponse struct {
	Period            string `json:"period"`
	Date              string `json:"date"`
	CurrencyCode      string `json:"currencyCode"`
	IncomeTotalMinor  int64  `json:"incomeTotalMinor"`
	ExpenseTotalMinor int64  `json:"expenseTotalMinor"`
	NetBalanceMinor   int64  `json:"netBalanceMinor"`
}

type MonthlyReportResponse struct {
	Period              string                `json:"period"`
	Month               string                `json:"month"`
	CurrencyCode        string                `json:"currencyCode"`
	IncomeTotalMinor    int64                 `json:"incomeTotalMinor"`
	ExpenseTotalMinor   int64                 `json:"expenseTotalMinor"`
	NetBalanceMinor     int64                 `json:"netBalanceMinor"`
	ClosingBalanceMinor int64                 `json:"closingBalanceMinor"`
	TopCategories       []TopCategoryResponse `json:"topCategories"`
}

type InitialBalanceResponse struct {
	InitialBalanceMinor int64  `json:"initialBalanceMinor"`
	CurrencyCode        string `json:"currencyCode"`
	InitializedAt       string `json:"initializedAt"`
}

type SetInitialBalanceRequest struct {
	InitialBalanceMinor int64 `json:"initialBalanceMinor"`
}

type HealthResponse struct {
	Status string `json:"status"`
}
