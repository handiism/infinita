package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
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
	transportserver "github.com/handiism/infinita/internal/transport/server"
)

const envDataDir = "INFINITA_DATA_DIR"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dataDir, err := resolveDataDir()
	if err != nil {
		log.Fatalf("determine data directory: %v", err)
	}

	db, err := sqlite.OpenDatabase(dataDir)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer func() { _ = db.Close() }()

	queries := sqlc.New(db)
	categoryRepo := sqlite.NewCategoryRepository(queries)
	transactionRepo := sqlite.NewTransactionRepository(queries)
	budgetRepo := sqlite.NewBudgetRepository(queries)
	settingRepo := sqlite.NewSettingRepository(queries)
	initialBalanceRepo := sqlite.NewInitialBalanceRepository(queries)

	if err := enforceLocalOnly(context.Background(), settingRepo); err != nil {
		log.Fatalf("startup validation: %v", err)
	}

	txnUsecase := usecase.NewTransactionUseCase(transactionRepo, categoryRepo)
	categoryUsecase := usecase.NewCategoryUseCase(categoryRepo)
	budgetUsecase := usecase.NewBudgetUseCase(budgetRepo, categoryRepo)
	reportUsecase := usecase.NewReportUseCase(transactionRepo, initialBalanceRepo, settingRepo)
	settingsUsecase := usecase.NewSettingsUseCase(settingRepo, initialBalanceRepo)

	handler := transportserver.NewHandler(
		txnUsecase,
		categoryUsecase,
		budgetUsecase,
		reportUsecase,
		settingsUsecase,
	)
	mux := transportserver.NewRouter(handler)
	server := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}

	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatalf("failed to create listener: %v", err)
		}
	}
	log.Printf("Starting server on http://%s", ln.Addr().String())

	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("Server stopped")
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
