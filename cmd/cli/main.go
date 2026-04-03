package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/handiism/infinita/internal/application/port/output"
	"github.com/handiism/infinita/internal/application/usecase"
	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/infrastructure/database/sqlite"
	"github.com/handiism/infinita/internal/infrastructure/database/sqlite/sqlc"
	transportcli "github.com/handiism/infinita/internal/transport/cli"
	transportclient "github.com/handiism/infinita/internal/transport/client"
	transportserver "github.com/handiism/infinita/internal/transport/server"
	ucli "github.com/urfave/cli/v3"
)

const envDataDir = "INFINITA_DATA_DIR"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dataDir, err := resolveDataDir()
	if err != nil {
		exitRuntime(fmt.Errorf("determine data directory: %w", err))
	}

	db, err := sqlite.OpenDatabase(dataDir)
	if err != nil {
		exitRuntime(fmt.Errorf("database: %w", err))
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("close database: %v", closeErr)
		}
	}()

	queries := sqlc.New(db)
	categoryRepo := sqlite.NewCategoryRepository(queries)
	transactionRepo := sqlite.NewTransactionRepository(queries)
	budgetRepo := sqlite.NewBudgetRepository(queries)
	settingRepo := sqlite.NewSettingRepository(queries)
	initialBalanceRepo := sqlite.NewInitialBalanceRepository(queries)

	if err := enforceLocalOnly(context.Background(), settingRepo); err != nil {
		exitRuntime(err)
	}

	// Create use cases for the server
	txnUsecase := usecase.NewTransactionUseCase(transactionRepo, categoryRepo)
	categoryUsecase := usecase.NewCategoryUseCase(categoryRepo)
	budgetUsecase := usecase.NewBudgetUseCase(budgetRepo, categoryRepo)
	reportUsecase := usecase.NewReportUseCase(transactionRepo, initialBalanceRepo, settingRepo)
	settingsUsecase := usecase.NewSettingsUseCase(settingRepo, initialBalanceRepo)

	// Start embedded HTTP server
	server := transportserver.New(
		txnUsecase,
		categoryUsecase,
		budgetUsecase,
		reportUsecase,
		settingsUsecase,
	)

	portCh, err := server.Start(ctx)
	if err != nil {
		exitRuntime(fmt.Errorf("start server: %w", err))
	}

	var port int
	select {
	case port = <-portCh:
	case <-ctx.Done():
		exitRuntime(fmt.Errorf("server startup cancelled"))
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Wait for server to be ready
	readyCtx, readyCancel := context.WithTimeout(ctx, 5*time.Second)
	defer readyCancel()
	if err := server.WaitForReady(readyCtx, baseURL); err != nil {
		exitRuntime(fmt.Errorf("server readiness check failed: %w", err))
	}

	// Create HTTP client that implements use case interfaces
	client := transportclient.New(baseURL)

	app := transportcli.NewApp(
		client,
		client,
		client,
		client,
		client,
		os.Stdout,
		os.Stderr,
	)

	go func() {
		select {
		case serverErr := <-server.Err():
			if serverErr != nil {
				fmt.Fprintln(os.Stderr, "server error:", serverErr)
				cancel()
			}
		case <-ctx.Done():
		}
	}()

	err = app.Command().Run(ctx, os.Args)

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
		fmt.Fprintln(os.Stderr, "server shutdown error:", shutdownErr)
	}

	if err != nil {
		ucli.HandleExitCoder(err)
		os.Exit(exitCode(err))
	}
}

func exitRuntime(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(3)
}

func resolveDataDir() (string, error) {
	if env := os.Getenv(envDataDir); env != "" {
		return env, nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "infinita"), nil
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

func exitCode(err error) int {
	var clientErr *transportclient.ClientError
	if errors.As(err, &clientErr) {
		return clientErr.ExitCode()
	}
	return 3
}
