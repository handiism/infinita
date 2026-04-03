package output

import (
	"context"

	"github.com/handiism/infinita/internal/domain/entity"
)

// SettingsRepository defines the interface for settings persistence.
type SettingsRepository interface {
	GetSettings(ctx context.Context) (entity.Settings, error)
	SetStorageMode(ctx context.Context, mode string) error
	SetReportTimezone(ctx context.Context, timezone string) error
}
