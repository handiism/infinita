package bootstrap

import (
	"context"
	"fmt"
	"io"
	"time"

	transportcli "github.com/handiism/infinita/internal/transport/cli"
	transportclient "github.com/handiism/infinita/internal/transport/client"
)

func RunCLI(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	paths, err := ResolvePaths(args)
	if err != nil {
		return fmt.Errorf("determine runtime paths: %w", err)
	}

	runtime, err := NewRuntime(ctx, paths.DataDir, paths.SettingsFile)
	if err != nil {
		return fmt.Errorf("initialize runtime: %w", err)
	}
	defer func() {
		if closeErr := runtime.Close(); closeErr != nil {
			_, _ = fmt.Fprintln(stderr, "database close error:", closeErr)
		}
	}()

	baseURL, err := startEmbeddedServer(ctx, runtime)
	if err != nil {
		return err
	}

	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), serverLifecycleTimeout*time.Second)
		defer shutdownCancel()
		if shutdownErr := runtime.Server.Shutdown(shutdownCtx); shutdownErr != nil {
			_, _ = fmt.Fprintln(stderr, "server shutdown error:", shutdownErr)
		}
	}()

	client := transportclient.New(baseURL)
	app := transportcli.NewApp(
		client, client, client, client, client,
		paths.SettingsFile,
		stdout,
		stderr,
	)

	go logServerErrors(ctx, runtime, stderr)

	return app.Command().Run(ctx, args)
}

func startEmbeddedServer(ctx context.Context, runtime *Runtime) (string, error) {
	portCh, err := runtime.Server.Start(ctx)
	if err != nil {
		return "", fmt.Errorf("start server: %w", err)
	}

	var port int
	select {
	case port = <-portCh:
	case <-ctx.Done():
		return "", fmt.Errorf("server startup cancelled")
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	readyCtx, readyCancel := context.WithTimeout(ctx, serverLifecycleTimeout*time.Second)
	defer readyCancel()
	if err := runtime.Server.WaitForReady(readyCtx, baseURL); err != nil {
		return "", fmt.Errorf("server readiness check failed: %w", err)
	}

	return baseURL, nil
}

func logServerErrors(ctx context.Context, runtime *Runtime, stderr io.Writer) {
	select {
	case serverErr := <-runtime.Server.Err():
		if serverErr != nil {
			_, _ = fmt.Fprintln(stderr, "server error:", serverErr)
		}
	case <-ctx.Done():
	}
}
