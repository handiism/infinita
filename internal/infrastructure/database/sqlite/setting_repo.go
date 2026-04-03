package sqlite

import (
	"context"
	"fmt"

	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/infrastructure/database/sqlite/sqlc"
)

type SettingRepository struct {
	queries *sqlc.Queries
}

func NewSettingRepository(queries *sqlc.Queries) *SettingRepository {
	return &SettingRepository{queries: queries}
}

func (r *SettingRepository) GetSettings(ctx context.Context) (entity.Settings, error) {
	storage, err := r.queries.GetSetting(ctx, "storage_mode")
	if err != nil {
		return entity.Settings{}, fmt.Errorf("storage mode: %w", err)
	}
	timezone, err := r.queries.GetSetting(ctx, "report_timezone")
	if err != nil {
		return entity.Settings{}, fmt.Errorf("report timezone: %w", err)
	}
	return entity.Settings{
		StorageMode:    storage.Value,
		ReportTimezone: timezone.Value,
	}, nil
}

func (r *SettingRepository) SetStorageMode(ctx context.Context, mode string) error {
	if mode != "local" {
		return domainerror.ErrInvalidStorageMode.WithField("storage_mode").WithHint("storage mode must remain 'local' in MVP")
	}
	return r.queries.UpsertSetting(ctx, sqlc.UpsertSettingParams{Key: "storage_mode", Value: mode})
}

func (r *SettingRepository) SetReportTimezone(ctx context.Context, timezone string) error {
	return r.queries.UpsertSetting(ctx, sqlc.UpsertSettingParams{Key: "report_timezone", Value: timezone})
}
