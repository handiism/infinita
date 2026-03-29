package usecase

import (
	"context"

	"github.com/handiism/infinita/internal/application/port/output"
	"github.com/handiism/infinita/internal/application/validation"
	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
)

type SettingsUseCase struct {
	settingsRepo       output.SettingsRepository
	initialBalanceRepo output.InitialBalanceRepository
}

func NewSettingsUseCase(settingsRepo output.SettingsRepository, initialBalanceRepo output.InitialBalanceRepository) *SettingsUseCase {
	return &SettingsUseCase{settingsRepo: settingsRepo, initialBalanceRepo: initialBalanceRepo}
}

func (u *SettingsUseCase) Show(ctx context.Context) (entity.Settings, error) {
	return u.settingsRepo.GetSettings(ctx)
}

func (u *SettingsUseCase) SetInitialBalance(ctx context.Context, amount int64) (entity.InitialBalance, error) {
	if amount < 0 {
		return entity.InitialBalance{}, domainerror.ErrInvalidAmount
	}
	return u.initialBalanceRepo.Set(ctx, amount, "IDR")
}

func (u *SettingsUseCase) ResetInitialBalance(ctx context.Context) error {
	_, err := u.initialBalanceRepo.Set(ctx, 0, "IDR")
	return err
}

func (u *SettingsUseCase) SetAnalyticsOptIn(ctx context.Context, optIn bool) error {
	return u.settingsRepo.SetAnalyticsOptIn(ctx, optIn)
}

func (u *SettingsUseCase) SetReportTimezone(ctx context.Context, timezone string) error {
	zone, err := validation.ParseTimezone(timezone)
	if err != nil {
		return err
	}
	return u.settingsRepo.SetReportTimezone(ctx, zone)
}
