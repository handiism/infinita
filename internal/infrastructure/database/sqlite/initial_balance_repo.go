package sqlite

import (
	"context"
	"fmt"

	"github.com/handiism/infinita/internal/domain/entity"
	"github.com/handiism/infinita/internal/infrastructure/database/sqlite/sqlc"
)

type InitialBalanceRepository struct {
	queries *sqlc.Queries
}

func NewInitialBalanceRepository(queries *sqlc.Queries) *InitialBalanceRepository {
	return &InitialBalanceRepository{queries: queries}
}

func (r *InitialBalanceRepository) Get(ctx context.Context) (entity.InitialBalance, error) {
	row, err := r.queries.GetInitialBalance(ctx)
	if err != nil {
		return entity.InitialBalance{}, fmt.Errorf("fetch initial balance: %w", err)
	}
	return entity.InitialBalance{
		InitialBalanceMinor: row.InitialBalanceMinor,
		CurrencyCode:        row.CurrencyCode,
		InitializedAt:       row.InitializedAt,
	}, nil
}

func (r *InitialBalanceRepository) Set(ctx context.Context, amount int64, currency string) (entity.InitialBalance, error) {
	row, err := r.queries.UpsertInitialBalance(ctx, sqlc.UpsertInitialBalanceParams{InitialBalanceMinor: amount, CurrencyCode: currency})
	if err != nil {
		return entity.InitialBalance{}, fmt.Errorf("upsert initial balance: %w", err)
	}
	return entity.InitialBalance{
		InitialBalanceMinor: row.InitialBalanceMinor,
		CurrencyCode:        row.CurrencyCode,
		InitializedAt:       row.InitializedAt,
	}, nil
}
