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

// Environment variable names for path overrides.
const (
	envDataDir        = "INFINITA_DATA_DIR"
	envSettingsFile   = "INFINITA_SETTINGS_FILE"
	serverGracePeriod = 5 // seconds
)

// DBCloser is the interface for database-like resources that can be closed.
type DBCloser interface {
	Close() error
}

// Runtime holds the server and its resources.
type Runtime struct {
	Server *transportserver.Server
	closer DBCloser
}

// Bootstrap holds the resolved paths and runtime for a running application.
type Bootstrap struct {
	Paths   Paths
	Runtime *Runtime
}

// NewBootstrap resolves paths and creates a Runtime. The caller must call
// Bootstrap.Close when done.
func NewBootstrap(ctx context.Context, args []string) (*Bootstrap, error) {
	paths, err := ResolvePaths(args)
	if err != nil {
		return nil, fmt.Errorf("determine runtime paths: %w", err)
	}

	runtime, err := NewRuntime(ctx, paths.DataDir, paths.SettingsFile)
	if err != nil {
		return nil, fmt.Errorf("initialize runtime: %w", err)
	}

	return &Bootstrap{
		Paths:   paths,
		Runtime: runtime,
	}, nil
}

// Close releases all resources held by the bootstrap.
func (b *Bootstrap) Close() error {
	if b == nil || b.Runtime == nil {
		return nil
	}
	return b.Runtime.Close()
}

// ResolveDataDir returns the data directory, checking INFINITA_DATA_DIR first.
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

// ResolveSettingsFile returns the settings file path, checking INFINITA_SETTINGS_FILE first.
func ResolveSettingsFile(dataDir string) string {
	if env := os.Getenv(envSettingsFile); env != "" {
		return env
	}
	return filepath.Join(dataDir, "settings.yaml")
}

// NewRuntime creates a new Runtime with the database, repositories, use cases, and server.
func NewRuntime(ctx context.Context, dataDir string, settingsFile string) (*Runtime, error) {
	db, queries, err := openDatabase(dataDir)
	if err != nil {
		return nil, err
	}

	repos := wireRepositories(queries, settingsFile)

	if err := enforceLocalOnly(ctx, repos.Settings); err != nil {
		_ = db.Close()
		return nil, err
	}

	usecases := wireUseCases(repos)

	server := transportserver.New(
		usecases.Transaction,
		usecases.Category,
		usecases.Budget,
		usecases.Report,
		usecases.Settings,
	)

	return &Runtime{
		Server: server,
		closer: db,
	}, nil
}

// Close closes the underlying database connection.
func (r *Runtime) Close() error {
	if r == nil || r.closer == nil {
		return nil
	}
	return r.closer.Close()
}

// openDatabase opens the SQLite database and returns the connection, queries, and any error.
func openDatabase(dataDir string) (DBCloser, *sqlc.Queries, error) {
	db, err := sqlite.OpenDatabase(dataDir)
	if err != nil {
		return nil, nil, fmt.Errorf("database: %w", err)
	}
	queries := sqlc.New(db)
	return db, queries, nil
}

// repositories holds all repositories used by the application.
type repositories struct {
	Category       *sqlite.CategoryRepository
	Transaction    *sqlite.TransactionRepository
	Budget         *sqlite.BudgetRepository
	Settings       *infinitasettings.SettingsRepository
	InitialBalance *sqlite.InitialBalanceRepository
}

// wireRepositories creates all repositories with the given database queries and settings path.
func wireRepositories(queries *sqlc.Queries, settingsFile string) *repositories {
	return &repositories{
		Category:       sqlite.NewCategoryRepository(queries),
		Transaction:    sqlite.NewTransactionRepository(queries),
		Budget:         sqlite.NewBudgetRepository(queries),
		Settings:       infinitasettings.NewSettingsRepository(settingsFile),
		InitialBalance: sqlite.NewInitialBalanceRepository(queries),
	}
}

// useCases holds all use cases used by the application.
type useCases struct {
	Transaction *usecase.TransactionUseCase
	Category    *usecase.CategoryUseCase
	Budget      *usecase.BudgetUseCase
	Report      *usecase.ReportUseCase
	Settings    *usecase.SettingsUseCase
}

// wireUseCases creates all use cases with the given repositories.
func wireUseCases(repos *repositories) *useCases {
	return &useCases{
		Transaction: usecase.NewTransactionUseCase(repos.Transaction, repos.Category),
		Category:    usecase.NewCategoryUseCase(repos.Category),
		Budget:      usecase.NewBudgetUseCase(repos.Budget, repos.Category),
		Report:      usecase.NewReportUseCase(repos.Transaction, repos.InitialBalance, repos.Settings),
		Settings:    usecase.NewSettingsUseCase(repos.Settings, repos.InitialBalance),
	}
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
