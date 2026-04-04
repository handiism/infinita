package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	domainerror "github.com/handiism/infinita/internal/domain/error"
	transportcli "github.com/handiism/infinita/internal/transport/cli"
	transportclient "github.com/handiism/infinita/internal/transport/client"
	"github.com/handiism/infinita/internal/version"
)

// cliFlags holds flags parsed from raw CLI args before cobra initialization.
type cliFlags struct {
	mode      string
	serverURL string
	apiKey    string
	config    string
}

// parseCLIFlags extracts known flags from raw CLI args before cobra initialization.
// This is necessary because the storage mode decision (local vs remote) must be made
// before creating the CLI app, which determines whether to start an embedded server
// or connect to a remote server.
func parseCLIFlags(args []string) cliFlags {
	var flags cliFlags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--mode":
			if i+1 < len(args) {
				flags.mode = args[i+1]
				i++
			}
		case "--server-url":
			if i+1 < len(args) {
				flags.serverURL = args[i+1]
				i++
			}
		case "--api-key":
			if i+1 < len(args) {
				flags.apiKey = args[i+1]
				i++
			}
		case "--config":
			if i+1 < len(args) {
				flags.config = args[i+1]
				i++
			}
		}
	}
	return flags
}

// RunCLI starts the embedded server and runs the CLI client.
func RunCLI(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	// Handle --version early, before any heavy initialization
	for _, arg := range args {
		if arg == "--version" || arg == "-v" {
			fmt.Fprintln(stdout, version.Version)
			return nil
		}
	}

	flags := parseCLIFlags(args)

	b, err := NewBootstrap(ctx, args)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := b.Close(); closeErr != nil {
			_, _ = fmt.Fprintln(stderr, "database close error:", closeErr)
		}
	}()

	// Override settings file path if --config was provided
	if flags.config != "" {
		b.Paths.SettingsFile = flags.config
	}

	// Get settings to determine storage mode
	settings, err := b.Runtime.Settings.GetSettings(ctx)
	if err != nil {
		return err
	}

	// Apply CLI flag overrides
	if flags.mode != "" {
		settings.Mode = flags.mode
	}
	if flags.serverURL != "" {
		settings.ServerURL = flags.serverURL
	}
	if flags.apiKey != "" {
		settings.APIKey = flags.apiKey
	}

	// Validate overridden settings
	if err := validateStorageMode(settings); err != nil {
		return err
	}

	var baseURL string
	var app *transportcli.App

	if settings.Mode == "remote" && settings.ServerURL != "" {
		// Remote mode: connect to remote server
		baseURL = settings.ServerURL
		client := transportclient.New(baseURL, settings.APIKey)
		app = transportcli.NewApp(
			client, client, client, client,
			b.Paths.SettingsFile,
			stdout,
			stderr,
		)
	} else {
		// Local mode: start embedded server
		baseURL, err = startEmbeddedServer(ctx, b.Runtime)
		if err != nil {
			return err
		}

		defer shutdownServer(ctx, b.Runtime.Server, stderr)

		client := transportclient.New(baseURL, settings.APIKey)
		app = transportcli.NewApp(
			client, client, client, client,
			b.Paths.SettingsFile,
			stdout,
			stderr,
		)

		go logServerErrors(ctx, b.Runtime, stderr)
	}

	rootCmd := app.Command()
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(args[1:]) // Skip program name

	return rootCmd.Execute()
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

// HandleExitError extracts the exit code from a domainerror.ExitError.
// Returns 2 for domain/validation errors and unknown commands, 3 for runtime errors.
func HandleExitError(err error) int {
	var exitErr domainerror.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Code
	}
	// Cobra's built-in unknown command error
	if err != nil && strings.Contains(err.Error(), "unknown command") {
		return 2
	}
	return 3
}
