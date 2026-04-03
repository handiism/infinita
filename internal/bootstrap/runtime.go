package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/handiism/infinita/internal/application/port/output"
	"github.com/handiism/infinita/internal/application/usecase"
	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/infrastructure/database/sqlite"
	"github.com/handiism/infinita/internal/infrastructure/database/sqlite/sqlc"
	infinitasettings "github.com/handiism/infinita/internal/infrastructure/settings"
	transportserver "github.com/handiism/infinita/internal/transport/server"
)

const envDataDir = "INFINITA_DATA_DIR"
const envSettingsFile = "INFINITA_SETTINGS_FILE"

type Runtime struct {
	Server *transportserver.Server
	db     interface{ Close() error }
}

func ResolveDataDir() (string, error) {
	if env := os.Getenv(envDataDir); env != "" {
		return env, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "infinita"), nil
}

func ResolveSettingsFile(dataDir string) string {
	if env := os.Getenv(envSettingsFile); env != "" {
		return env
	}
	return filepath.Join(dataDir, "settings.yaml")
}

func NewRuntime(ctx context.Context, dataDir string, settingsFile string) (*Runtime, error) {
	db, err := sqlite.OpenDatabase(dataDir)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	queries := sqlc.New(db)
	categoryRepo := sqlite.NewCategoryRepository(queries)
	transactionRepo := sqlite.NewTransactionRepository(queries)
	budgetRepo := sqlite.NewBudgetRepository(queries)
	settingRepo := infinitasettings.NewSettingsRepository(settingsFile)
	initialBalanceRepo := sqlite.NewInitialBalanceRepository(queries)

	if err := enforceLocalOnly(ctx, settingRepo); err != nil {
		_ = db.Close()
		return nil, err
	}

	txnUsecase := usecase.NewTransactionUseCase(transactionRepo, categoryRepo)
	categoryUsecase := usecase.NewCategoryUseCase(categoryRepo)
	budgetUsecase := usecase.NewBudgetUseCase(budgetRepo, categoryRepo)
	reportUsecase := usecase.NewReportUseCase(transactionRepo, initialBalanceRepo, settingRepo)
	settingsUsecase := usecase.NewSettingsUseCase(settingRepo, initialBalanceRepo)

	server := transportserver.New(
		txnUsecase,
		categoryUsecase,
		budgetUsecase,
		reportUsecase,
		settingsUsecase,
	)

	return &Runtime{
		Server: server,
		db:     db,
	}, nil
}

func (r *Runtime) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}

func enforceLocalOnly(ctx context.Context, settingsRepo output.SettingsRepository) error {
	settings, err := settingsRepo.GetSettings(ctx)
	if err != nil {
		return err
	}
	if settings.StorageMode != "local" {
		return domainerror.ErrInvalidStorageMode.WithField("storage_mode").WithHint(fmt.Sprintf("unsupported configured mode '%s'; storage mode must remain local in MVP", settings.StorageMode))
	}
	return nil
}
