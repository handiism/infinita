package bootstrap

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/handiism/infinita/internal/domain/entity"
)

func TestResolveDataDirUsesEnvironmentOverride(t *testing.T) {
	t.Setenv(envDataDir, "/tmp/infinita-bootstrap-test")

	got, err := ResolveDataDir()
	if err != nil {
		t.Fatalf("ResolveDataDir() error = %v", err)
	}
	if got != "/tmp/infinita-bootstrap-test" {
		t.Fatalf("ResolveDataDir() = %q, want %q", got, "/tmp/infinita-bootstrap-test")
	}
}

func TestResolveDataDirUsesUserConfigDir(t *testing.T) {
	t.Setenv(envDataDir, "")

	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Skipf("os.UserConfigDir() error: %v", err)
	}
	want := filepath.Join(configDir, "infinita")

	got, err := ResolveDataDir()
	if err != nil {
		t.Fatalf("ResolveDataDir() error = %v", err)
	}
	if got != want {
		t.Fatalf("ResolveDataDir() = %q, want %q", got, want)
	}
}

func TestResolveSettingsFileUsesEnvironmentOverride(t *testing.T) {
	t.Setenv(envSettingsFile, "/tmp/custom-settings.yaml")

	got := ResolveSettingsFile("/default/dir")
	if got != "/tmp/custom-settings.yaml" {
		t.Fatalf("ResolveSettingsFile() = %q, want %q", got, "/tmp/custom-settings.yaml")
	}
}

func TestResolveSettingsFileFallsBackToDefault(t *testing.T) {
	t.Setenv(envSettingsFile, "")

	got := ResolveSettingsFile("/default/dir")
	want := filepath.Join("/default/dir", "settings.yaml")
	if got != want {
		t.Fatalf("ResolveSettingsFile() = %q, want %q", got, want)
	}
}

func TestResolvePathsUsesConfigArgumentOverride(t *testing.T) {
	t.Setenv(envDataDir, "/tmp/infinita-bootstrap-test")
	t.Setenv(envSettingsFile, "")

	got, err := ResolvePaths([]string{"infinita", "--config", "/tmp/override.yaml"})
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}

	if got.DataDir != "/tmp/infinita-bootstrap-test" {
		t.Fatalf("ResolvePaths().DataDir = %q, want %q", got.DataDir, "/tmp/infinita-bootstrap-test")
	}
	if got.SettingsFile != "/tmp/override.yaml" {
		t.Fatalf("ResolvePaths().SettingsFile = %q, want %q", got.SettingsFile, "/tmp/override.yaml")
	}
	if got.IsDefaultConfig {
		t.Fatalf("ResolvePaths().IsDefaultConfig = true, want false (override was provided)")
	}
}

func TestResolvePathsFromDataDirFallsBackToResolvedSettingsFile(t *testing.T) {
	t.Setenv(envSettingsFile, "")

	got := ResolvePathsFromDataDir("/tmp/infinita-bootstrap-test", nil)
	want := filepath.Join("/tmp/infinita-bootstrap-test", "settings.yaml")

	if got.DataDir != "/tmp/infinita-bootstrap-test" {
		t.Fatalf("ResolvePathsFromDataDir().DataDir = %q, want %q", got.DataDir, "/tmp/infinita-bootstrap-test")
	}
	if got.SettingsFile != want {
		t.Fatalf("ResolvePathsFromDataDir().SettingsFile = %q, want %q", got.SettingsFile, want)
	}
	if !got.IsDefaultConfig {
		t.Fatalf("ResolvePathsFromDataDir().IsDefaultConfig = false, want true")
	}
}

func TestParseConfigArg(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		want   string
		wantOK bool
	}{
		{
			name:   "no config arg",
			args:   []string{"infinita", "list"},
			want:   "",
			wantOK: false,
		},
		{
			name:   "config with separate value",
			args:   []string{"infinita", "--config", "/path/to/config.yaml", "list"},
			want:   "/path/to/config.yaml",
			wantOK: true,
		},
		{
			name:   "config with equals syntax",
			args:   []string{"infinita", "--config=/path/to/config.yaml", "list"},
			want:   "/path/to/config.yaml",
			wantOK: true,
		},
		{
			name:   "config at end",
			args:   []string{"infinita", "list", "--config=/final.yaml"},
			want:   "/final.yaml",
			wantOK: true,
		},
		{
			name:   "config without value",
			args:   []string{"infinita", "--config"},
			want:   "",
			wantOK: false,
		},
		{
			name:   "config equals without value",
			args:   []string{"infinita", "--config="},
			want:   "",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseConfigArg(tt.args)
			if got != tt.want {
				t.Errorf("parseConfigArg() got = %q, want %q", got, tt.want)
			}
			if ok != tt.wantOK {
				t.Errorf("parseConfigArg() ok = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}

func TestBootstrapClose_NilReceiver(t *testing.T) {
	var b *Bootstrap
	if err := b.Close(); err != nil {
		t.Errorf("Bootstrap.Close() on nil Bootstrap = %v, want nil", err)
	}
}

func TestBootstrapClose_NilRuntime(t *testing.T) {
	b := &Bootstrap{Runtime: nil}
	if err := b.Close(); err != nil {
		t.Errorf("Bootstrap.Close() on nil Runtime = %v, want nil", err)
	}
}

func TestRuntimeClose_NilCloser(t *testing.T) {
	r := &Runtime{closer: nil}
	if err := r.Close(); err != nil {
		t.Errorf("Runtime.Close() on nil closer = %v, want nil", err)
	}
}

func TestValidateStorageMode(t *testing.T) {
	tests := []struct {
		name     string
		settings entity.Settings
		wantErr  string
	}{
		{
			name:     "local mode allowed",
			settings: entity.Settings{Mode: "local"},
		},
		{
			name:     "non local/remote rejected",
			settings: entity.Settings{Mode: "invalid"},
			wantErr:  "MODE_UNAVAILABLE",
		},
		{
			name:     "remote with server_url and api_key allowed",
			settings: entity.Settings{Mode: "remote", ServerURL: "http://localhost:8080", APIKey: "secret"},
		},
		{
			name:     "remote without server_url rejected",
			settings: entity.Settings{Mode: "remote", ServerURL: "", APIKey: "secret"},
			wantErr:  "INVALID_CONFIG",
		},
		{
			name:     "remote without api_key rejected",
			settings: entity.Settings{Mode: "remote", ServerURL: "http://localhost:8080", APIKey: ""},
			wantErr:  "MISSING_API_KEY",
		},
		{
			name:     "remote without server_url and api_key rejected",
			settings: entity.Settings{Mode: "remote", ServerURL: "", APIKey: ""},
			wantErr:  "INVALID_CONFIG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStorageMode(tt.settings)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateStorageMode() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateStorageMode() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestNewRuntimeCreatesUsableServer(t *testing.T) {
	dataDir := t.TempDir()
	settingsFile := ResolveSettingsFile(dataDir)
	runtime, err := NewRuntime(context.Background(), dataDir, settingsFile, true)
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	defer func() {
		if err := runtime.Close(); err != nil {
			t.Fatalf("runtime.Close() error = %v", err)
		}
	}()

	if runtime.Server == nil {
		t.Fatal("NewRuntime() returned nil Server")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	portCh, err := runtime.Server.Start(ctx)
	if err != nil {
		t.Fatalf("runtime.Server.Start() error = %v", err)
	}

	var port int
	select {
	case port = <-portCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server port")
	}

	readyCtx, readyCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer readyCancel()
	if err := runtime.Server.WaitForReady(readyCtx, "http://127.0.0.1:"+strconv.Itoa(port)); err != nil {
		t.Fatalf("runtime.Server.WaitForReady() error = %v", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := runtime.Server.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("runtime.Server.Shutdown() error = %v", err)
	}
}
