package client

import (
	"fmt"
	"time"

	domainerror "github.com/handiism/infinita/internal/domain/error"
)

type Response struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Code    string      `json:"code,omitempty"`
	Meta    interface{} `json:"meta"`
}

type ClientError struct {
	Code           string
	Message        string
	DomainError    domainerror.DomainError
	MultipleErrors []domainerror.DomainError
	StatusCode     int
}

func (e *ClientError) Error() string {
	if e.DomainError.Code != "" {
		return e.DomainError.Error()
	}
	if len(e.MultipleErrors) > 0 {
		return fmt.Sprintf("%s: %s (and %d more errors)",
			e.MultipleErrors[0].Code,
			e.MultipleErrors[0].Message,
			len(e.MultipleErrors)-1)
	}
	if e.Code != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return "unknown error"
}

func (e *ClientError) ExitCode() int {
	if e.StatusCode >= 500 {
		return 3
	}
	if e.StatusCode >= 400 {
		return 2
	}
	return 3
}

func (e *ClientError) ToDomainError() domainerror.DomainError {
	if e.DomainError.Code != "" {
		return e.DomainError
	}
	return domainerror.DomainError{
		Code:    e.Code,
		Message: e.Message,
	}
}

func (e *ClientError) ToDomainErrors() []domainerror.DomainError {
	if e.DomainError.Code != "" {
		return []domainerror.DomainError{e.DomainError}
	}
	if len(e.MultipleErrors) > 0 {
		return e.MultipleErrors
	}
	return []domainerror.DomainError{{Code: e.Code, Message: e.Message}}
}

func IsClientError(err error) bool {
	_, ok := err.(*ClientError)
	return ok
}

type TransactionRequest struct {
	Type         string `json:"type"`
	AmountMinor  int64  `json:"amountMinor"`
	CurrencyCode string `json:"currencyCode"`
	Category     string `json:"category"`
	Date         string `json:"date"`
	Description  string `json:"description"`
}

type TransactionResponse struct {
	ID                   string    `json:"id"`
	Type                 string    `json:"type"`
	AmountMinor          int64     `json:"amountMinor"`
	CurrencyCode         string    `json:"currencyCode"`
	CategoryID           int64     `json:"categoryId"`
	CategoryNameSnapshot string    `json:"categoryNameSnapshot"`
	Date                 string    `json:"date"`
	Description          string    `json:"description"`
	CreatedAt            time.Time `json:"createdAt"`
}

type TransactionListResponse []TransactionResponse

type PaginationMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
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

type CategoryListResponse []CategoryResponse

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

type BudgetStatusListResponse []BudgetStatusResponse

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

type SettingsResponse struct {
	StorageMode    string `json:"storageMode"`
	AnalyticsOptIn bool   `json:"analyticsOptIn"`
	ReportTimezone string `json:"reportTimezone"`
}

type InitialBalanceResponse struct {
	InitialBalanceMinor int64  `json:"initialBalanceMinor"`
	CurrencyCode        string `json:"currencyCode"`
	InitializedAt       string `json:"initializedAt"`
}

type SetInitialBalanceRequest struct {
	InitialBalanceMinor int64  `json:"initialBalanceMinor"`
	CurrencyCode        string `json:"currencyCode"`
}

type SetAnalyticsOptInRequest struct {
	AnalyticsOptIn bool `json:"analyticsOptIn"`
}

type SetReportTimezoneRequest struct {
	ReportTimezone string `json:"reportTimezone"`
}
