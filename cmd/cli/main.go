package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/handiism/infinita/internal/bootstrap"
	transportcli "github.com/handiism/infinita/internal/transport/cli"
	transportclient "github.com/handiism/infinita/internal/transport/client"
	ucli "github.com/urfave/cli/v3"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dataDir, err := resolveDataDir()
	if err != nil {
		exitRuntime(fmt.Errorf("determine data directory: %w", err))
	}

	// Pre-parse --config before urfave/cli runs, since runtime needs it for bootstrap.
	settingsFile := resolveSettingsFile(dataDir, os.Args)

	runtime, err := bootstrap.NewRuntime(context.Background(), dataDir, settingsFile)
	if err != nil {
		exitRuntime(fmt.Errorf("database: %w", err))
	}
	defer func() {
		if closeErr := runtime.Close(); closeErr != nil {
			fmt.Fprintln(os.Stderr, "database close error:", closeErr)
		}
	}()

	portCh, err := runtime.Server.Start(ctx)
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
	if err := runtime.Server.WaitForReady(readyCtx, baseURL); err != nil {
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
		settingsFile,
		os.Stdout,
		os.Stderr,
	)

	go func() {
		select {
		case serverErr := <-runtime.Server.Err():
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
	if shutdownErr := runtime.Server.Shutdown(shutdownCtx); shutdownErr != nil {
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
	return bootstrap.ResolveDataDir()
}

func resolveSettingsFile(dataDir string, args []string) string {
	for i, arg := range args {
		if arg == "--config" && i+1 < len(args) {
			return args[i+1]
		}
		if after, ok := strings.CutPrefix(arg, "--config="); ok {
			return after
		}
	}
	return bootstrap.ResolveSettingsFile(dataDir)
}

func exitCode(err error) int {
	var clientErr *transportclient.ClientError
	if errors.As(err, &clientErr) {
		return clientErr.ExitCode()
	}
	return 3
}
