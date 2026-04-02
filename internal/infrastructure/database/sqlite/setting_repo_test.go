package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/handiism/infinita/internal/infrastructure/database/sqlite/sqlc"
)

func TestSettingRepository_GetAndUpdate(t *testing.T) {
	db := newTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	queries := sqlc.New(db)
	repo := NewSettingRepository(queries)

	settings, err := repo.GetSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, "local", settings.StorageMode)
	require.Equal(t, "Asia/Jakarta", settings.ReportTimezone)
	require.False(t, settings.AnalyticsOptIn)

	require.NoError(t, repo.SetStorageMode(ctx, "local"))
	require.NoError(t, repo.SetReportTimezone(ctx, "UTC"))
	require.NoError(t, repo.SetAnalyticsOptIn(ctx, true))

	settings, err = repo.GetSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, "local", settings.StorageMode)
	require.Equal(t, "UTC", settings.ReportTimezone)
	require.True(t, settings.AnalyticsOptIn)

	optIn, err := repo.GetAnalyticsOptIn(ctx)
	require.NoError(t, err)
	require.True(t, optIn)

	err = repo.SetStorageMode(ctx, "remote")
	require.Error(t, err)
	require.Contains(t, err.Error(), "STORAGE_MODE_UNAVAILABLE")
}
