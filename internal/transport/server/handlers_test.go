package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/handiism/infinita/internal/application/port/input"
	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
)

func TestAddTransactionReturnsFailForWrappedDomainError(t *testing.T) {
	h := NewHandler(
		stubTransactionUseCase{addErr: fmt.Errorf("wrapped: %w", domainerror.ErrUnknownCategory)},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{},
	)

	req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(`{"type":"expense","amountMinor":1000,"currencyCode":"IDR","category":"food","date":"2026-03-29"}`))
	rec := httptest.NewRecorder()

	h.AddTransaction(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var resp Response
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, StatusFail, resp.Status)

	data, ok := resp.Data.([]interface{})
	require.True(t, ok)
	require.Len(t, data, 1)

	first, ok := data[0].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, domainerror.ErrUnknownCategory.Code, first["code"])
}

func TestAddTransactionRejectsUnsupportedCurrency(t *testing.T) {
	h := NewHandler(
		stubTransactionUseCase{},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{},
	)

	req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(`{"type":"expense","amountMinor":1000,"currencyCode":"USD","category":"food","date":"2026-03-29"}`))
	rec := httptest.NewRecorder()

	h.AddTransaction(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), domainerror.ErrInvalidCurrency.Code)
}

func TestListTransactionsWritesTopLevelMeta(t *testing.T) {
	createdAt := time.Date(2026, 3, 29, 10, 0, 0, 0, time.UTC)
	h := NewHandler(
		stubTransactionUseCase{listResult: input.TransactionListResult{
			Transactions: []entity.Transaction{{
				ID:                   "txn-1",
				Type:                 "expense",
				AmountMinor:          5000,
				CurrencyCode:         "IDR",
				CategoryID:           1,
				CategoryNameSnapshot: "food",
				Date:                 "2026-03-29",
				Description:          "lunch",
				CreatedAt:            createdAt,
			}},
			Total: 1,
		}},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{},
	)

	req := httptest.NewRequest(http.MethodGet, "/transactions?limit=10&offset=5", nil)
	rec := httptest.NewRecorder()

	h.ListTransactions(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp Response
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, StatusSuccess, resp.Status)

	data, ok := resp.Data.([]interface{})
	require.True(t, ok)
	require.Len(t, data, 1)

	meta, ok := resp.Meta.(map[string]interface{})
	require.True(t, ok)
	require.EqualValues(t, 1, meta["total"])
	require.EqualValues(t, 10, meta["limit"])
	require.EqualValues(t, 5, meta["offset"])

	first, ok := data[0].(map[string]interface{})
	require.True(t, ok)
	_, hasNestedData := first["data"]
	require.False(t, hasNestedData)
}

type stubTransactionUseCase struct {
	addErr     error
	listResult input.TransactionListResult
	listErr    error
	listFn     func(context.Context, *string, int, int) (input.TransactionListResult, error)
}

func (s stubTransactionUseCase) AddTransaction(context.Context, string, int64, string, string, string) (entity.Transaction, error) {
	return entity.Transaction{}, s.addErr
}

func (s stubTransactionUseCase) ListTransactions(ctx context.Context, category *string, limit, offset int) (input.TransactionListResult, error) {
	if s.listFn != nil {
		return s.listFn(ctx, category, limit, offset)
	}
	return s.listResult, s.listErr
}

type stubCategoryUseCase struct{}

func (stubCategoryUseCase) List(context.Context) ([]entity.Category, error) {
	return nil, nil
}

func (stubCategoryUseCase) Create(context.Context, string, string) (entity.Category, error) {
	return entity.Category{}, nil
}

type stubBudgetUseCase struct{}

func (stubBudgetUseCase) SetBudget(context.Context, string, string, int64) error {
	return nil
}

func (stubBudgetUseCase) Status(context.Context, string) ([]entity.BudgetStatus, error) {
	return nil, nil
}

type stubReportUseCase struct {
	dailyResult   entity.DailySummary
	monthlyResult entity.MonthlySummary
}

func (s stubReportUseCase) Daily(context.Context, string) (entity.DailySummary, error) {
	if s.dailyResult.Period != "" || s.dailyResult.CurrencyCode != "" {
		return s.dailyResult, nil
	}
	return entity.DailySummary{CurrencyCode: "IDR"}, nil
}

func (s stubReportUseCase) Monthly(context.Context, string) (entity.MonthlySummary, error) {
	if s.monthlyResult.Period != "" || s.monthlyResult.CurrencyCode != "" {
		return s.monthlyResult, nil
	}
	return entity.MonthlySummary{CurrencyCode: "IDR"}, nil
}

type stubSettingsUseCase struct {
	setInitialBalanceFn func(context.Context, int64) (entity.InitialBalance, error)
}

func (stubSettingsUseCase) Show(context.Context) (entity.Settings, error) {
	return entity.Settings{}, nil
}

func (s stubSettingsUseCase) SetInitialBalance(ctx context.Context, amount int64) (entity.InitialBalance, error) {
	if s.setInitialBalanceFn != nil {
		return s.setInitialBalanceFn(ctx, amount)
	}
	return entity.InitialBalance{}, nil
}

func (stubSettingsUseCase) ResetInitialBalance(context.Context) error {
	return nil
}

func (stubSettingsUseCase) SetReportTimezone(context.Context, string) error {
	return nil
}

func TestGetDailyReportUsesSummaryPeriod(t *testing.T) {
	h := NewHandler(
		stubTransactionUseCase{},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{dailyResult: entity.DailySummary{Period: "2026-03-29", CurrencyCode: "IDR"}},
		stubSettingsUseCase{},
	)

	req := httptest.NewRequest(http.MethodGet, "/reports/daily?date=+2026-03-29+", nil)
	rec := httptest.NewRecorder()

	h.GetDailyReport(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Status string              `json:"status"`
		Data   DailyReportResponse `json:"data"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, string(StatusSuccess), resp.Status)
	require.Equal(t, "daily", resp.Data.Period)
	require.Equal(t, "2026-03-29", resp.Data.Date)
}

func TestGetMonthlyReportUsesSummaryPeriod(t *testing.T) {
	h := NewHandler(
		stubTransactionUseCase{},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{monthlyResult: entity.MonthlySummary{Period: "2026-03", CurrencyCode: "IDR"}},
		stubSettingsUseCase{},
	)

	req := httptest.NewRequest(http.MethodGet, "/reports/monthly?month=+2026-03+", nil)
	rec := httptest.NewRecorder()

	h.GetMonthlyReport(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Status string                `json:"status"`
		Data   MonthlyReportResponse `json:"data"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, string(StatusSuccess), resp.Status)
	require.Equal(t, "monthly", resp.Data.Period)
	require.Equal(t, "2026-03", resp.Data.Month)
}

func TestSetInitialBalanceReturnsPersistedPayload(t *testing.T) {
	h := NewHandler(
		stubTransactionUseCase{},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{setInitialBalanceFn: func(_ context.Context, amount int64) (entity.InitialBalance, error) {
			return entity.InitialBalance{
				InitialBalanceMinor: amount,
				CurrencyCode:        "IDR",
				InitializedAt:       "2026-03-29T10:00:00Z",
			}, nil
		}},
	)

	req := httptest.NewRequest(http.MethodPut, "/settings/initial-balance", strings.NewReader(`{"initialBalanceMinor":100,"currencyCode":"IDR"}`))
	rec := httptest.NewRecorder()

	h.SetInitialBalance(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Status string                 `json:"status"`
		Data   InitialBalanceResponse `json:"data"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, string(StatusSuccess), resp.Status)
	require.Equal(t, int64(100), resp.Data.InitialBalanceMinor)
	require.Equal(t, "IDR", resp.Data.CurrencyCode)
	require.Equal(t, "2026-03-29T10:00:00Z", resp.Data.InitializedAt)
}

func TestSetInitialBalanceRejectsUnsupportedCurrency(t *testing.T) {
	h := NewHandler(
		stubTransactionUseCase{},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{},
	)

	req := httptest.NewRequest(http.MethodPut, "/settings/initial-balance", strings.NewReader(`{"initialBalanceMinor":100,"currencyCode":"USD"}`))
	rec := httptest.NewRecorder()

	h.SetInitialBalance(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), domainerror.ErrInvalidCurrency.Code)
}

func TestListTransactionsRejectsLimitAboveMaximum(t *testing.T) {
	h := NewHandler(
		stubTransactionUseCase{listFn: func(_ context.Context, _ *string, _, _ int) (input.TransactionListResult, error) {
			t.Fatal("ListTransactions use case should not be called for invalid limit")
			return input.TransactionListResult{}, nil
		}},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{},
	)

	req := httptest.NewRequest(http.MethodGet, "/transactions?limit=600", nil)
	rec := httptest.NewRecorder()

	h.ListTransactions(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var resp Response
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, StatusFail, resp.Status)

	data, ok := resp.Data.([]interface{})
	require.True(t, ok)
	require.Len(t, data, 1)

	first, ok := data[0].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, domainerror.ErrInvalidFlag.Code, first["code"])
	require.Equal(t, "limit", first["field"])
}

func TestListTransactionsRejectsNegativeOffset(t *testing.T) {
	h := NewHandler(
		stubTransactionUseCase{listFn: func(_ context.Context, _ *string, _, _ int) (input.TransactionListResult, error) {
			t.Fatal("ListTransactions use case should not be called for invalid offset")
			return input.TransactionListResult{}, nil
		}},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{},
	)

	req := httptest.NewRequest(http.MethodGet, "/transactions?offset=-5", nil)
	rec := httptest.NewRecorder()

	h.ListTransactions(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var resp Response
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, StatusFail, resp.Status)

	data, ok := resp.Data.([]interface{})
	require.True(t, ok)
	require.Len(t, data, 1)

	first, ok := data[0].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, domainerror.ErrInvalidFlag.Code, first["code"])
	require.Equal(t, "offset", first["field"])
}

func TestResetInitialBalanceIncludesNullDataField(t *testing.T) {
	h := NewHandler(
		stubTransactionUseCase{},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{},
	)

	req := httptest.NewRequest(http.MethodDelete, "/settings/initial-balance", nil)
	rec := httptest.NewRecorder()

	h.ResetInitialBalance(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var payload map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
	require.Equal(t, string(StatusSuccess), payload["status"])

	data, exists := payload["data"]
	require.True(t, exists)
	require.Nil(t, data)
}
