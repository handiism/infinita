package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/handiism/infinita/internal/bootstrap"
	transportclient "github.com/handiism/infinita/internal/transport/client"
	ucli "github.com/urfave/cli/v3"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		ucli.HandleExitCoder(err)
		os.Exit(exitCode(err))
	}
}

func run(ctx context.Context) error {
	return bootstrap.RunCLI(ctx, os.Args, os.Stdout, os.Stderr)
}

func exitCode(err error) int {
	var clientErr *transportclient.ClientError
	if errors.As(err, &clientErr) {
		return clientErr.ExitCode()
	}
	return 3
}
