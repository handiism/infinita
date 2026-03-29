package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

func (stubSettingsRepo) SetAnalyticsOptIn(context.Context, bool) error {
	return nil
}

func (stubSettingsRepo) GetAnalyticsOptIn(context.Context) (bool, error) {
	return false, nil
}

func TestResolveDataDirUsesEnvironmentOverride(t *testing.T) {
	t.Setenv(envDataDir, "/tmp/infinita-test")

	got, err := resolveDataDir()
	if err != nil {
		t.Fatalf("resolveDataDir() error = %v", err)
	}
	if got != "/tmp/infinita-test" {
		t.Fatalf("resolveDataDir() = %q, want %q", got, "/tmp/infinita-test")
	}
}

func TestResolveDataDirUsesUserConfigDir(t *testing.T) {
	t.Setenv(envDataDir, "")

	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Skipf("os.UserConfigDir() error: %v", err)
	}
	want := filepath.Join(configDir, "infinita")

	got, err := resolveDataDir()
	if err != nil {
		t.Fatalf("resolveDataDir() error = %v", err)
	}
	if got != want {
		t.Fatalf("resolveDataDir() = %q, want %q", got, want)
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
