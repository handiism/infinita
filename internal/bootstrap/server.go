package bootstrap

import (
	"context"
	"fmt"
	"io"
	"net"
)

// RunServer starts the server on a dynamically chosen port and blocks until context cancellation.
func RunServer(ctx context.Context, stdout io.Writer, stderr io.Writer) error {
	b, err := NewBootstrap(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := b.Close(); closeErr != nil {
			_, _ = fmt.Fprintln(stderr, "database close error:", closeErr)
		}
	}()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	_, _ = fmt.Fprintf(stdout, "Starting server on http://%s\n", ln.Addr().String())

	if _, err := b.Runtime.Server.StartOnListener(ctx, ln); err != nil {
		_ = ln.Close()
		return fmt.Errorf("server start error: %w", err)
	}

	select {
	case <-ctx.Done():
	case err := <-b.Runtime.Server.Err():
		return fmt.Errorf("server error: %w", err)
	}

	_, _ = fmt.Fprintln(stdout, "Shutting down server...")
	if err := b.Runtime.Server.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	_, _ = fmt.Fprintln(stdout, "Server stopped")
	return nil
}
