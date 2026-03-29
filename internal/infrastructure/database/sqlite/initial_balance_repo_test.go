package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/handiism/infinita/internal/infrastructure/database/sqlite/sqlc"
)

func TestInitialBalanceRepository_GetAndSet(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	ctx := context.Background()
	repo := NewInitialBalanceRepository(sqlc.New(db))

	balance, err := repo.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(0), balance.InitialBalanceMinor)

	balance, err = repo.Set(ctx, 25000, "IDR")
	require.NoError(t, err)
	require.Equal(t, int64(25000), balance.InitialBalanceMinor)
	require.Equal(t, "IDR", balance.CurrencyCode)
	_, err = time.Parse(time.RFC3339, balance.InitializedAt)
	require.NoError(t, err)
}
