package sqlite

import (
	"context"
	"fmt"

	"github.com/handiism/infinita/internal/domain/entity"
	"github.com/handiism/infinita/internal/domain/valueobject"
	"github.com/handiism/infinita/internal/infrastructure/database/sqlite/sqlc"
)

type BudgetRepository struct {
	queries *sqlc.Queries
}

func NewBudgetRepository(queries *sqlc.Queries) *BudgetRepository {
	return &BudgetRepository{queries: queries}
}

func (r *BudgetRepository) UpsertBudget(ctx context.Context, categoryID int64, month string, limit int64) error {
	if err := r.queries.UpsertBudget(ctx, sqlc.UpsertBudgetParams{CategoryID: categoryID, Month: month, MonthlyLimitMinor: limit}); err != nil {
		return fmt.Errorf("upsert budget: %w", err)
	}
	return nil
}

func (r *BudgetRepository) ListBudgetsByMonth(ctx context.Context, month string) ([]entity.BudgetStatus, error) {
	rows, err := r.queries.GetBudgetsForMonth(ctx, month)
	if err != nil {
		return nil, fmt.Errorf("query budgets: %w", err)
	}
	result := make([]entity.BudgetStatus, 0, len(rows))
	for _, row := range rows {
		spentMinor, err := sqliteInt64(row.SpentMonthToDateMinor)
		if err != nil {
			return nil, fmt.Errorf("convert spent month to date minor: %w", err)
		}

		remaining := row.MonthlyLimitMinor - spentMinor
		result = append(result, entity.BudgetStatus{
			CategoryName:          row.CategoryName,
			CategoryKey:           row.CategoryKey,
			CurrencyCode:          valueobject.DefaultCurrencyCode,
			Month:                 row.Month,
			MonthlyLimitMinor:     row.MonthlyLimitMinor,
			SpentMonthToDateMinor: spentMinor,
			RemainingMinor:        remaining,
			IsOverLimit:           remaining < 0,
		})
	}
	return result, nil
}
