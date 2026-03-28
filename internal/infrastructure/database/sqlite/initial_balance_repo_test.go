package sqlite

import (
	"context"
	"testing"

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

	require.NoError(t, repo.Set(ctx, 25000, "IDR"))
	balance, err = repo.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(25000), balance.InitialBalanceMinor)
}
