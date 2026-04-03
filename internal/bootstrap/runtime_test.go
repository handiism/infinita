package bootstrap

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/handiism/infinita/internal/domain/entity"
)

type stubSettingsRepo struct {
	settings entity.Settings
	err      error
}

func (s stubSettingsRepo) GetSettings(context.Context) (entity.Settings, error) {
	if s.err != nil {
		return entity.Settings{}, s.err
	}
	return s.settings, nil
}

func (stubSettingsRepo) SetStorageMode(context.Context, string) error {
	return nil
}

func (stubSettingsRepo) SetReportTimezone(context.Context, string) error {
	return nil
}

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

func TestEnforceLocalOnly(t *testing.T) {
	tests := []struct {
		name    string
		repo    stubSettingsRepo
		wantErr string
	}{
		{
			name: "local mode allowed",
			repo: stubSettingsRepo{settings: entity.Settings{StorageMode: "local"}},
		},
		{
			name:    "repository error",
			repo:    stubSettingsRepo{err: errors.New("boom")},
			wantErr: "boom",
		},
		{
			name:    "non local rejected",
			repo:    stubSettingsRepo{settings: entity.Settings{StorageMode: "remote"}},
			wantErr: "STORAGE_MODE_UNAVAILABLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := enforceLocalOnly(context.Background(), tt.repo)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("enforceLocalOnly() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("enforceLocalOnly() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestNewRuntimeCreatesUsableServer(t *testing.T) {
	dataDir := t.TempDir()
	settingsFile := ResolveSettingsFile(dataDir)
	runtime, err := NewRuntime(context.Background(), dataDir, settingsFile)
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
