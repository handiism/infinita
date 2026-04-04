package usecase

import (
	"context"
	"fmt"

	"github.com/handiism/infinita/internal/application/port/output"
	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/domain/valueobject"
)

// SettingsUseCase implements settings-related operations.
type SettingsUseCase struct {
	settingsRepo       output.SettingsRepository
	initialBalanceRepo output.InitialBalanceRepository
}

// NewSettingsUseCase creates a new SettingsUseCase.
func NewSettingsUseCase(settingsRepo output.SettingsRepository, initialBalanceRepo output.InitialBalanceRepository) *SettingsUseCase {
	return &SettingsUseCase{settingsRepo: settingsRepo, initialBalanceRepo: initialBalanceRepo}
}

func (u *SettingsUseCase) SetInitialBalance(ctx context.Context, amount int64) (entity.InitialBalance, error) {
	if amount < 0 {
		return entity.InitialBalance{}, domainerror.ErrInvalidAmount.WithField("amount")
	}
	return u.initialBalanceRepo.Set(ctx, amount, valueobject.DefaultCurrencyCode)
}

func (u *SettingsUseCase) ResetInitialBalance(ctx context.Context) error {
	_, err := u.initialBalanceRepo.Set(ctx, 0, valueobject.DefaultCurrencyCode)
	if err != nil {
		return fmt.Errorf("reset initial balance: %w", err)
	}
	return nil
}
