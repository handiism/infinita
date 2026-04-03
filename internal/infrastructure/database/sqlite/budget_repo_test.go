package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/handiism/infinita/internal/domain/entity"
	"github.com/handiism/infinita/internal/infrastructure/database/sqlite/sqlc"
)

func TestBudgetRepository_UpsertAndListByMonth(t *testing.T) {
	db := newTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("close test db: %v", err)
		}
	}()
	ctx := context.Background()
	queries := sqlc.New(db)
	catRepo := NewCategoryRepository(queries)
	budgetRepo := NewBudgetRepository(queries)

	cat, err := catRepo.GetByNormalizedKey(ctx, "food")
	require.NoError(t, err)

	require.NoError(t, budgetRepo.UpsertBudget(ctx, cat.ID, "2026-03", 50000))

	statuses, err := budgetRepo.ListBudgetsByMonth(ctx, "2026-03")
	require.NoError(t, err)
	require.Len(t, statuses, 1)
	require.Equal(t, "Food", statuses[0].CategoryName)
	require.Equal(t, "food", statuses[0].CategoryKey)
	require.Equal(t, "2026-03", statuses[0].Month)
	require.Equal(t, int64(50000), statuses[0].MonthlyLimitMinor)
	require.Equal(t, int64(0), statuses[0].SpentMonthToDateMinor)
	require.Equal(t, int64(50000), statuses[0].RemainingMinor)
	require.False(t, statuses[0].IsOverLimit)
	require.Equal(t, "IDR", statuses[0].CurrencyCode)
}

func TestBudgetRepository_UpsertUpdatesExisting(t *testing.T) {
	db := newTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("close test db: %v", err)
		}
	}()
	ctx := context.Background()
	queries := sqlc.New(db)
	catRepo := NewCategoryRepository(queries)
	budgetRepo := NewBudgetRepository(queries)

	cat, err := catRepo.GetByNormalizedKey(ctx, "transport")
	require.NoError(t, err)

	require.NoError(t, budgetRepo.UpsertBudget(ctx, cat.ID, "2026-03", 30000))
	require.NoError(t, budgetRepo.UpsertBudget(ctx, cat.ID, "2026-03", 60000))

	statuses, err := budgetRepo.ListBudgetsByMonth(ctx, "2026-03")
	require.NoError(t, err)
	require.Len(t, statuses, 1)
	require.Equal(t, int64(60000), statuses[0].MonthlyLimitMinor)
}

func TestBudgetRepository_ListByMonthEmpty(t *testing.T) {
	db := newTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("close test db: %v", err)
		}
	}()
	budgetRepo := NewBudgetRepository(sqlc.New(db))

	statuses, err := budgetRepo.ListBudgetsByMonth(context.Background(), "2025-01")
	require.NoError(t, err)
	require.Empty(t, statuses)
}

func TestBudgetRepository_SpentReflectsTransactions(t *testing.T) {
	db := newTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("close test db: %v", err)
		}
	}()
	ctx := context.Background()
	queries := sqlc.New(db)
	catRepo := NewCategoryRepository(queries)
	txnRepo := NewTransactionRepository(queries)
	budgetRepo := NewBudgetRepository(queries)

	cat, err := catRepo.GetByNormalizedKey(ctx, "food")
	require.NoError(t, err)

	require.NoError(t, budgetRepo.UpsertBudget(ctx, cat.ID, "2026-03", 100000))

	require.NoError(t, txnRepo.Create(ctx, entity.Transaction{
		ID: "txn-budget-1", Type: "expense", AmountMinor: 40000,
		CurrencyCode: "IDR", CategoryID: cat.ID,
		CategoryNameSnapshot: cat.Name, Date: "2026-03-10",
	}))
	require.NoError(t, txnRepo.Create(ctx, entity.Transaction{
		ID: "txn-budget-2", Type: "expense", AmountMinor: 25000,
		CurrencyCode: "IDR", CategoryID: cat.ID,
		CategoryNameSnapshot: cat.Name, Date: "2026-03-20",
	}))

	statuses, err := budgetRepo.ListBudgetsByMonth(ctx, "2026-03")
	require.NoError(t, err)
	require.Len(t, statuses, 1)
	require.Equal(t, int64(65000), statuses[0].SpentMonthToDateMinor)
	require.Equal(t, int64(35000), statuses[0].RemainingMinor)
	require.False(t, statuses[0].IsOverLimit)
}

func TestBudgetRepository_OverLimitDetected(t *testing.T) {
	db := newTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("close test db: %v", err)
		}
	}()
	ctx := context.Background()
	queries := sqlc.New(db)
	catRepo := NewCategoryRepository(queries)
	txnRepo := NewTransactionRepository(queries)
	budgetRepo := NewBudgetRepository(queries)

	cat, err := catRepo.GetByNormalizedKey(ctx, "food")
	require.NoError(t, err)

	require.NoError(t, budgetRepo.UpsertBudget(ctx, cat.ID, "2026-04", 10000))

	require.NoError(t, txnRepo.Create(ctx, entity.Transaction{
		ID: "txn-over-1", Type: "expense", AmountMinor: 15000,
		CurrencyCode: "IDR", CategoryID: cat.ID,
		CategoryNameSnapshot: cat.Name, Date: "2026-04-01",
	}))

	statuses, err := budgetRepo.ListBudgetsByMonth(ctx, "2026-04")
	require.NoError(t, err)
	require.Len(t, statuses, 1)
	require.Equal(t, int64(15000), statuses[0].SpentMonthToDateMinor)
	require.Equal(t, int64(-5000), statuses[0].RemainingMinor)
	require.True(t, statuses[0].IsOverLimit)
}

func TestBudgetRepository_MultipleCategoriesSortedByName(t *testing.T) {
	db := newTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("close test db: %v", err)
		}
	}()
	ctx := context.Background()
	queries := sqlc.New(db)
	catRepo := NewCategoryRepository(queries)
	budgetRepo := NewBudgetRepository(queries)

	food, err := catRepo.GetByNormalizedKey(ctx, "food")
	require.NoError(t, err)
	transport, err := catRepo.GetByNormalizedKey(ctx, "transport")
	require.NoError(t, err)

	require.NoError(t, budgetRepo.UpsertBudget(ctx, transport.ID, "2026-03", 20000))
	require.NoError(t, budgetRepo.UpsertBudget(ctx, food.ID, "2026-03", 50000))

	statuses, err := budgetRepo.ListBudgetsByMonth(ctx, "2026-03")
	require.NoError(t, err)
	require.Len(t, statuses, 2)
	require.Equal(t, "Food", statuses[0].CategoryName)
	require.Equal(t, "Transport", statuses[1].CategoryName)
}

func TestBudgetRepository_IncomeNotCountedAsSpent(t *testing.T) {
	db := newTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("close test db: %v", err)
		}
	}()
	ctx := context.Background()
	queries := sqlc.New(db)
	catRepo := NewCategoryRepository(queries)
	txnRepo := NewTransactionRepository(queries)
	budgetRepo := NewBudgetRepository(queries)

	cat, err := catRepo.GetByNormalizedKey(ctx, "savings")
	require.NoError(t, err)

	require.NoError(t, budgetRepo.UpsertBudget(ctx, cat.ID, "2026-03", 50000))

	require.NoError(t, txnRepo.Create(ctx, entity.Transaction{
		ID: "txn-income-1", Type: "income", AmountMinor: 100000,
		CurrencyCode: "IDR", CategoryID: cat.ID,
		CategoryNameSnapshot: cat.Name, Date: "2026-03-15",
	}))

	statuses, err := budgetRepo.ListBudgetsByMonth(ctx, "2026-03")
	require.NoError(t, err)
	require.Len(t, statuses, 1)
	require.Equal(t, int64(0), statuses[0].SpentMonthToDateMinor)
	require.Equal(t, int64(50000), statuses[0].RemainingMinor)
}
