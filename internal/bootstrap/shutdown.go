package bootstrap

import (
	"context"
	"fmt"
	"io"
	"time"

	transportserver "github.com/handiism/infinita/internal/transport/server"
)

// shutdownServer gracefully shuts down the server, logging any errors to stderr.
func shutdownServer(ctx context.Context, srv *transportserver.Server, stderr io.Writer) {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), serverGracePeriod*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		_, _ = fmt.Fprintln(stderr, "server shutdown error:", err)
	}
}
