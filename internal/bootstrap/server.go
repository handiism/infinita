package bootstrap

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

func RunServer(ctx context.Context, stdout io.Writer, stderr io.Writer) error {
	paths, err := ResolvePaths(nil)
	if err != nil {
		return fmt.Errorf("determine runtime paths: %w", err)
	}

	runtime, err := NewRuntime(ctx, paths.DataDir, paths.SettingsFile)
	if err != nil {
		return fmt.Errorf("runtime: %w", err)
	}
	defer func() {
		if closeErr := runtime.Close(); closeErr != nil {
			_, _ = fmt.Fprintln(stderr, "database close error:", closeErr)
		}
	}()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	_, _ = fmt.Fprintf(stdout, "Starting server on http://%s\n", ln.Addr().String())

	if _, err := runtime.Server.StartOnListener(ctx, ln); err != nil {
		_ = ln.Close()
		return fmt.Errorf("server start error: %w", err)
	}

	select {
	case <-ctx.Done():
	case err := <-runtime.Server.Err():
		return fmt.Errorf("server error: %w", err)
	}

	_, _ = fmt.Fprintln(stdout, "Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), serverLifecycleTimeout*time.Second)
	defer shutdownCancel()

	if err := runtime.Server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	_, _ = fmt.Fprintln(stdout, "Server stopped")
	return nil
}
