package input

import (
	"context"

	"github.com/handiism/infinita/internal/domain/entity"
)

// TransactionUseCase defines the operations for managing transactions.
type TransactionUseCase interface {
	AddTransaction(ctx context.Context, entryType string, amountMinor int64, category string, date string, description string) (entity.Transaction, error)
	ListTransactions(ctx context.Context, category *string, limit, offset int) (TransactionListResult, error)
}

// TransactionListResult holds paginated transaction listing.
type TransactionListResult struct {
	Transactions []entity.Transaction
	Total        int
}

// CategoryUseCase defines the operations for managing categories.
type CategoryUseCase interface {
	List(ctx context.Context) ([]entity.Category, error)
	Create(ctx context.Context, name, description string) (entity.Category, error)
}

// BudgetUseCase defines the operations for managing budgets.
type BudgetUseCase interface {
	SetBudget(ctx context.Context, category, month string, limit int64) error
	Status(ctx context.Context, month string) ([]entity.BudgetStatus, error)
}

// ReportUseCase defines the operations for generating financial reports.
type ReportUseCase interface {
	Daily(ctx context.Context, date string) (entity.DailySummary, error)
	Monthly(ctx context.Context, month string) (entity.MonthlySummary, error)
}

type SettingsUseCase interface {
	SetInitialBalance(ctx context.Context, amount int64) (entity.InitialBalance, error)
	ResetInitialBalance(ctx context.Context) error
}
