package output

import (
	"context"

	"github.com/handiism/infinita/internal/domain/entity"
)

// BudgetRepository defines the interface for budget persistence.
type BudgetRepository interface {
	UpsertBudget(ctx context.Context, categoryID int64, month string, limit int64) error
	ListBudgetsByMonth(ctx context.Context, month string) ([]entity.BudgetStatus, error)
}
