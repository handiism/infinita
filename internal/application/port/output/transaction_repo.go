package output

import (
	"context"

	"github.com/handiism/infinita/internal/domain/entity"
)

type TransactionRepository interface {
	Create(ctx context.Context, txn entity.Transaction) error
	List(ctx context.Context, categoryKey *string, limit, offset int) ([]entity.Transaction, error)
	Count(ctx context.Context, categoryKey *string) (int, error)
	SumTotalsForDay(ctx context.Context, date string) (int64, int64, error)
	SumTotalsForMonth(ctx context.Context, month string) (int64, int64, error)
	SumCumulativeTotalsUpToDate(ctx context.Context, date string) (int64, int64, error)
	TopCategoriesForMonth(ctx context.Context, month string) ([]entity.TopSpendingCategory, error)
}
