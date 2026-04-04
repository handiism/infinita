package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/handiism/infinita/internal/bootstrap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(bootstrap.HandleExitError(err))
	}
}

func run(ctx context.Context) error {
	return bootstrap.RunCLI(ctx, os.Args, os.Stdout, os.Stderr)
}
