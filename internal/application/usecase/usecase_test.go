package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/handiism/infinita/internal/application/usecase"
	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/testutil/assertdomain"
)

func TestSettingsUseCase_ShowReturnsRepositoryValues(t *testing.T) {
	repo := &fakeSettingsRepository{settings: entity.Settings{
		StorageMode:    "local",
		ReportTimezone: "Asia/Jakarta",
		AnalyticsOptIn: true,
	}}
	uc := usecase.NewSettingsUseCase(repo, &fakeInitialBalanceRepository{})

	got, err := uc.Show(context.Background())
	require.NoError(t, err)
	require.Equal(t, repo.settings, got)
}

func TestSettingsUseCase_SetInitialBalanceValidatesAmount(t *testing.T) {
	repo := &fakeInitialBalanceRepository{initial: entity.InitialBalance{
		InitialBalanceMinor: 10,
		CurrencyCode:        "IDR",
		InitializedAt:       "2026-03-29T10:00:00Z",
	}}
	uc := usecase.NewSettingsUseCase(&fakeSettingsRepository{}, repo)

	initial, err := uc.SetInitialBalance(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, int64(10), repo.amount)
	require.Equal(t, "IDR", repo.currency)
	require.Equal(t, repo.setResult, initial)

	_, err = uc.SetInitialBalance(context.Background(), -1)
	assertdomain.Code(t, err, domainerror.ErrInvalidAmount.Code)
}

func TestSettingsUseCase_ResetInitialBalanceSetsZero(t *testing.T) {
	repo := &fakeInitialBalanceRepository{}
	uc := usecase.NewSettingsUseCase(&fakeSettingsRepository{}, repo)

	require.NoError(t, uc.ResetInitialBalance(context.Background()))
	require.Equal(t, int64(0), repo.amount)
	require.Equal(t, "IDR", repo.currency)
}

func TestBudgetUseCase_SetBudgetValidates(t *testing.T) {
	cases := []struct {
		name     string
		category string
		month    string
		limit    int64
		wantErr  error
	}{
		{name: "invalid limit", category: "Food", month: "2024-01", limit: 0, wantErr: domainerror.ErrInvalidAmount},
		{name: "invalid month", category: "Food", month: "2024-13", limit: 1000, wantErr: domainerror.ErrInvalidMonth},
		{name: "invalid category", category: "", month: "2024-01", limit: 1000, wantErr: domainerror.ErrInvalidCategory},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uc := usecase.NewBudgetUseCase(&fakeBudgetRepository{}, &fakeCategoryRepository{})
			err := uc.SetBudget(context.Background(), tc.category, tc.month, tc.limit)
			assertdomain.Code(t, err, tc.wantErr.(domainerror.DomainError).Code)
		})
	}
}

func TestBudgetUseCase_SetBudgetWrapsCategoryErrors(t *testing.T) {
	budgetRepo := &fakeBudgetRepository{}
	categoryRepo := &fakeCategoryRepository{err: errors.New("oops")}
	uc := usecase.NewBudgetUseCase(budgetRepo, categoryRepo)

	err := uc.SetBudget(context.Background(), "Transport", "2024-01", 1000)
	require.Error(t, err)
	require.Contains(t, err.Error(), "category lookup")
}

func TestBudgetUseCase_SetBudgetSuccess(t *testing.T) {
	budgetRepo := &fakeBudgetRepository{}
	categoryRepo := &fakeCategoryRepository{category: entity.Category{ID: 2}}
	uc := usecase.NewBudgetUseCase(budgetRepo, categoryRepo)

	require.NoError(t, uc.SetBudget(context.Background(), "Transport", "2024-01", 12345))
	require.Equal(t, int64(2), budgetRepo.categoryID)
	require.Equal(t, "2024-01", budgetRepo.month)
	require.Equal(t, int64(12345), budgetRepo.limit)
}

func TestBudgetUseCase_StatusValidatesMonth(t *testing.T) {
	uc := usecase.NewBudgetUseCase(&fakeBudgetRepository{}, &fakeCategoryRepository{})
	_, err := uc.Status(context.Background(), "2024-13")
	assertdomain.Code(t, err, domainerror.ErrInvalidMonth.Code)
}

func TestBudgetUseCase_StatusReturnsRepo(t *testing.T) {
	expected := []entity.BudgetStatus{{MonthlyLimitMinor: 5000}}
	budgetRepo := &fakeBudgetRepository{statuses: expected}
	uc := usecase.NewBudgetUseCase(budgetRepo, &fakeCategoryRepository{})

	got, err := uc.Status(context.Background(), "2024-01")
	require.NoError(t, err)
	require.Equal(t, expected, got)
}

func TestReportUseCase_Daily(t *testing.T) {
	repo := &fakeTransactionRepository{dailyIncome: 20000, dailyExpense: 7500}
	settingsRepo := &fakeSettingsRepository{settings: entity.Settings{ReportTimezone: "UTC"}}
	uc := usecase.NewReportUseCase(repo, &fakeInitialBalanceRepository{}, settingsRepo)

	summary, err := uc.Daily(context.Background(), "2024-01-02")
	require.NoError(t, err)
	require.Equal(t, int64(20000), summary.IncomeTotalMinor)
	require.Equal(t, int64(7500), summary.ExpenseTotalMinor)
	require.Equal(t, int64(12500), summary.NetBalanceMinor)

	summary, err = uc.Daily(context.Background(), " 2024-01-02 ")
	require.NoError(t, err)
	require.Equal(t, "2024-01-02", summary.Period)

	_, err = uc.Daily(context.Background(), "2024-13-02")
	assertdomain.Code(t, err, domainerror.ErrInvalidDate.Code)
}

func TestReportUseCase_Monthly(t *testing.T) {
	repo := &fakeTransactionRepository{
		monthlyIncome:     40000,
		monthlyExpense:    15000,
		cumulativeIncome:  45000,
		cumulativeExpense: 18000,
		topCategories:     []entity.TopSpendingCategory{{Category: "Food", AmountMinor: 15000}},
	}
	balanceRepo := &fakeInitialBalanceRepository{initial: entity.InitialBalance{InitialBalanceMinor: 10000}}
	settingsRepo := &fakeSettingsRepository{settings: entity.Settings{ReportTimezone: "UTC"}}
	uc := usecase.NewReportUseCase(repo, balanceRepo, settingsRepo)

	summary, err := uc.Monthly(context.Background(), "2024-01")
	require.NoError(t, err)
	require.Equal(t, int64(40000), summary.IncomeTotalMinor)
	require.Equal(t, int64(15000), summary.ExpenseTotalMinor)
	require.Equal(t, entity.TopSpendingCategory{Category: "Food", AmountMinor: 15000}, summary.TopCategories[0])
	require.Equal(t, int64(37000), summary.ClosingBalanceMinor)

	summary, err = uc.Monthly(context.Background(), " 2024-01 ")
	require.NoError(t, err)
	require.Equal(t, "2024-01", summary.Period)

	_, err = uc.Monthly(context.Background(), "2024-13")
	assertdomain.Code(t, err, domainerror.ErrInvalidMonth.Code)
}

func TestTransactionUseCase_AddTransaction(t *testing.T) {
	txRepo := &recordingTransactionRepository{createErr: nil}
	categoryRepo := &fakeCategoryRepository{category: entity.Category{ID: 3, Name: "Food"}}
	uc := usecase.NewTransactionUseCase(txRepo, categoryRepo)

	_, err := uc.AddTransaction(context.Background(), "income", 5000, "Food", "2024-01-05", "ok")
	require.NoError(t, err)
	require.Equal(t, int64(5000), txRepo.created.AmountMinor)
	require.Equal(t, "Food", txRepo.created.CategoryNameSnapshot)

	_, err = uc.AddTransaction(context.Background(), "transfer", 1000, "Food", "2024-01-05", "")
	assertdomain.Code(t, err, domainerror.ErrInvalidTransactionType.Code)
	_, err = uc.AddTransaction(context.Background(), "income", 0, "Food", "2024-01-05", "")
	assertdomain.Code(t, err, domainerror.ErrInvalidAmount.Code)
	_, err = uc.AddTransaction(context.Background(), "income", 1000, "", "2024-01-05", "")
	assertdomain.Code(t, err, domainerror.ErrInvalidCategory.Code)
	_, err = uc.AddTransaction(context.Background(), "income", 1000, "Food", "bad-date", "")
	assertdomain.Code(t, err, domainerror.ErrInvalidDate.Code)
}

func TestTransactionUseCase_ListTransactions(t *testing.T) {
	txRepo := &recordingTransactionRepository{listResponse: []entity.Transaction{{ID: "x"}}}
	uc := usecase.NewTransactionUseCase(txRepo, &fakeCategoryRepository{})

	got, err := uc.ListTransactions(context.Background(), nil, -1, -5)
	require.NoError(t, err)
	require.Equal(t, []entity.Transaction{{ID: "x"}}, got.Transactions)
	require.Equal(t, 50, txRepo.listLimit)
	require.Equal(t, 0, txRepo.listOffset)

	cat := "Food"
	_, err = uc.ListTransactions(context.Background(), &cat, 600, 0)
	require.NoError(t, err)
	require.Equal(t, 500, txRepo.listLimit)
}

func TestTransactionUseCase_ListTransactionsReturnsTotal(t *testing.T) {
	txRepo := &recordingTransactionRepository{
		listResponse:  []entity.Transaction{{ID: "x"}},
		countResponse: 42,
	}
	uc := usecase.NewTransactionUseCase(txRepo, &fakeCategoryRepository{})

	result, err := uc.ListTransactions(context.Background(), nil, 10, 0)
	require.NoError(t, err)
	require.Equal(t, 42, result.Total)
}

type fakeSettingsRepository struct {
	settings entity.Settings
}

func (f *fakeSettingsRepository) GetSettings(context.Context) (entity.Settings, error) {
	return f.settings, nil
}

func (*fakeSettingsRepository) SetStorageMode(context.Context, string) error    { return nil }
func (*fakeSettingsRepository) SetReportTimezone(context.Context, string) error { return nil }
func (*fakeSettingsRepository) SetAnalyticsOptIn(context.Context, bool) error   { return nil }
func (*fakeSettingsRepository) GetAnalyticsOptIn(context.Context) (bool, error) { return false, nil }

type fakeInitialBalanceRepository struct {
	amount    int64
	currency  string
	initial   entity.InitialBalance
	setResult entity.InitialBalance
}

func (f *fakeInitialBalanceRepository) Get(context.Context) (entity.InitialBalance, error) {
	return f.initial, nil
}

func (f *fakeInitialBalanceRepository) Set(_ context.Context, amount int64, currency string) (entity.InitialBalance, error) {
	f.amount = amount
	f.currency = currency
	if f.setResult.InitializedAt == "" {
		f.setResult = entity.InitialBalance{
			InitialBalanceMinor: amount,
			CurrencyCode:        currency,
			InitializedAt:       "2026-03-29 10:00:00",
		}
	}
	return f.setResult, nil
}

type fakeBudgetRepository struct {
	categoryID int64
	month      string
	limit      int64
	statuses   []entity.BudgetStatus
}

func (f *fakeBudgetRepository) UpsertBudget(_ context.Context, categoryID int64, month string, limit int64) error {
	f.categoryID = categoryID
	f.month = month
	f.limit = limit
	return nil
}

func (f *fakeBudgetRepository) ListBudgetsByMonth(_ context.Context, _ string) ([]entity.BudgetStatus, error) {
	return f.statuses, nil
}

type fakeCategoryRepository struct {
	category entity.Category
	err      error
}

func (f *fakeCategoryRepository) GetByNormalizedKey(_ context.Context, _ string) (entity.Category, error) {
	return f.category, f.err
}

func (*fakeCategoryRepository) List(context.Context) ([]entity.Category, error) { return nil, nil }
func (*fakeCategoryRepository) Create(context.Context, string, string, string) (entity.Category, error) {
	return entity.Category{}, nil
}

type fakeTransactionRepository struct {
	dailyIncome, dailyExpense           int64
	monthlyIncome, monthlyExpense       int64
	cumulativeIncome, cumulativeExpense int64
	topCategories                       []entity.TopSpendingCategory
}

func (*fakeTransactionRepository) Create(context.Context, entity.Transaction) error {
	return nil
}

func (*fakeTransactionRepository) List(context.Context, *string, int, int) ([]entity.Transaction, error) {
	return nil, nil
}

func (*fakeTransactionRepository) Count(context.Context, *string) (int, error) {
	return 0, nil
}

func (f *fakeTransactionRepository) SumTotalsForDay(context.Context, string) (int64, int64, error) {
	return f.dailyIncome, f.dailyExpense, nil
}

func (f *fakeTransactionRepository) SumTotalsForMonth(context.Context, string) (int64, int64, error) {
	return f.monthlyIncome, f.monthlyExpense, nil
}

func (f *fakeTransactionRepository) SumCumulativeTotalsUpToDate(context.Context, string) (int64, int64, error) {
	return f.cumulativeIncome, f.cumulativeExpense, nil
}

func (f *fakeTransactionRepository) TopCategoriesForMonth(context.Context, string) ([]entity.TopSpendingCategory, error) {
	return f.topCategories, nil
}

type recordingTransactionRepository struct {
	created       entity.Transaction
	createErr     error
	listLimit     int
	listOffset    int
	listCategory  *string
	listResponse  []entity.Transaction
	countResponse int
}

func (r *recordingTransactionRepository) Create(_ context.Context, txn entity.Transaction) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = txn
	return nil
}

func (r *recordingTransactionRepository) List(_ context.Context, category *string, limit, offset int) ([]entity.Transaction, error) {
	r.listCategory = category
	r.listLimit = limit
	r.listOffset = offset
	return r.listResponse, nil
}

func (*recordingTransactionRepository) SumTotalsForDay(context.Context, string) (int64, int64, error) {
	return 0, 0, nil
}

func (*recordingTransactionRepository) SumTotalsForMonth(context.Context, string) (int64, int64, error) {
	return 0, 0, nil
}

func (*recordingTransactionRepository) SumCumulativeTotalsUpToDate(context.Context, string) (int64, int64, error) {
	return 0, 0, nil
}

func (*recordingTransactionRepository) TopCategoriesForMonth(context.Context, string) ([]entity.TopSpendingCategory, error) {
	return nil, nil
}

func (r *recordingTransactionRepository) Count(context.Context, *string) (int, error) {
	return r.countResponse, nil
}
