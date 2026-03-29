package output

import (
	"context"

	"github.com/handiism/infinita/internal/domain/entity"
)

type InitialBalanceRepository interface {
	Get(ctx context.Context) (entity.InitialBalance, error)
	Set(ctx context.Context, amount int64, currency string) (entity.InitialBalance, error)
}
