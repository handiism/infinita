package bootstrap

import (
	"context"
	"fmt"
	"io"
	"time"

	transportcli "github.com/handiism/infinita/internal/transport/cli"
	transportclient "github.com/handiism/infinita/internal/transport/client"
)

// RunCLI starts the embedded server and runs the CLI client.
func RunCLI(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	b, err := NewBootstrap(ctx, args)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := b.Close(); closeErr != nil {
			_, _ = fmt.Fprintln(stderr, "database close error:", closeErr)
		}
	}()

	baseURL, err := startEmbeddedServer(ctx, b.Runtime)
	if err != nil {
		return err
	}

	defer shutdownServer(ctx, b.Runtime.Server, stderr)

	client := transportclient.New(baseURL)
	app := transportcli.NewApp(
		client, client, client, client, client,
		b.Paths.SettingsFile,
		stdout,
		stderr,
	)

	go logServerErrors(ctx, b.Runtime, stderr)

	return app.Command().Run(ctx, args)
}

// startEmbeddedServer starts the embedded server and waits for it to be ready.
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
	readyCtx, readyCancel := context.WithTimeout(ctx, serverGracePeriod*time.Second)
	defer readyCancel()
	if err := runtime.Server.WaitForReady(readyCtx, baseURL); err != nil {
		return "", fmt.Errorf("server readiness check failed: %w", err)
	}

	return baseURL, nil
}

// logServerErrors logs server errors to stderr until the context is cancelled.
func logServerErrors(ctx context.Context, runtime *Runtime, stderr io.Writer) {
	select {
	case serverErr := <-runtime.Server.Err():
		if serverErr != nil {
			_, _ = fmt.Fprintln(stderr, "server error:", serverErr)
		}
	case <-ctx.Done():
	}
}
