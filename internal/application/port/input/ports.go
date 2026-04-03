package input

import (
	"context"

	"github.com/handiism/infinita/internal/domain/entity"
)

type TransactionUseCase interface {
	AddTransaction(ctx context.Context, entryType string, amountMinor int64, category string, date string, description string) (entity.Transaction, error)
	ListTransactions(ctx context.Context, category *string, limit, offset int) (TransactionListResult, error)
}

type TransactionListResult struct {
	Transactions []entity.Transaction
	Total        int
}

type CategoryUseCase interface {
	List(ctx context.Context) ([]entity.Category, error)
	Create(ctx context.Context, name, description string) (entity.Category, error)
}

type BudgetUseCase interface {
	SetBudget(ctx context.Context, category, month string, limit int64) error
	Status(ctx context.Context, month string) ([]entity.BudgetStatus, error)
}

type ReportUseCase interface {
	Daily(ctx context.Context, date string) (entity.DailySummary, error)
	Monthly(ctx context.Context, month string) (entity.MonthlySummary, error)
}

type SettingsUseCase interface {
	Show(ctx context.Context) (entity.Settings, error)
	SetInitialBalance(ctx context.Context, amount int64) (entity.InitialBalance, error)
	ResetInitialBalance(ctx context.Context) error
	SetReportTimezone(ctx context.Context, timezone string) error
}
