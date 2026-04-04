package settings

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSettingsRepository_AutoCreatesDefaultFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "settings.yaml")
	repo := NewSettingsRepository(filePath, true)
	ctx := context.Background()

	// File does not exist yet.
	_, err := os.ReadFile(filePath)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))

	// GetSettings auto-creates the file with defaults.
	settings, err := repo.GetSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, "local", settings.StorageMode)
	require.Equal(t, "Asia/Jakarta", settings.ReportTimezone)

	// File now exists on disk.
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Contains(t, string(data), "storage_mode: local")
	require.Contains(t, string(data), "report_timezone: Asia/Jakarta")

	// Second read uses the file, not auto-create again.
	settings, err = repo.GetSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, "local", settings.StorageMode)
	require.Equal(t, "Asia/Jakarta", settings.ReportTimezone)
}

func TestSettingsRepository_DoesNotAutoCreateNonDefaultPath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "settings.yaml")
	repo := NewSettingsRepository(filePath, false)
	ctx := context.Background()

	// File does not exist and should not be auto-created.
	_, err := repo.GetSettings(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "settings file not found")

	// File still does not exist on disk.
	_, err = os.ReadFile(filePath)
	require.True(t, os.IsNotExist(err))
}

func TestSettingsRepository_ReadsExistingFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "settings.yaml")

	err := os.WriteFile(filePath, []byte("storage_mode: local\nreport_timezone: Europe/London\n"), 0o600)
	require.NoError(t, err)

	repo := NewSettingsRepository(filePath, true)
	settings, err := repo.GetSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, "local", settings.StorageMode)
	require.Equal(t, "Europe/London", settings.ReportTimezone)
}

func TestSettingsRepository_SetReportTimezone(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "settings.yaml")
	repo := NewSettingsRepository(filePath, true)
	ctx := context.Background()

	// First call auto-creates file, then SetReportTimezone updates it.
	settings, err := repo.GetSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, "Asia/Jakarta", settings.ReportTimezone)

	require.NoError(t, repo.SetReportTimezone(ctx, "UTC"))

	settings, err = repo.GetSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, "UTC", settings.ReportTimezone)
}

func TestSettingsRepository_RejectsInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "settings.yaml")

	err := os.WriteFile(filePath, []byte("storage_mode: [broken\n"), 0o600)
	require.NoError(t, err)

	repo := NewSettingsRepository(filePath, true)
	_, err = repo.GetSettings(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "INVALID_CONFIG")
}

func TestSettingsRepository_RejectsMissingStorageMode(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "settings.yaml")

	err := os.WriteFile(filePath, []byte("report_timezone: UTC\n"), 0o600)
	require.NoError(t, err)

	repo := NewSettingsRepository(filePath, true)
	_, err = repo.GetSettings(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "INVALID_CONFIG")
	require.Contains(t, err.Error(), "storage_mode")
}

func TestSettingsRepository_RejectsMissingTimezone(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "settings.yaml")

	err := os.WriteFile(filePath, []byte("storage_mode: local\n"), 0o600)
	require.NoError(t, err)

	repo := NewSettingsRepository(filePath, true)
	_, err = repo.GetSettings(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "INVALID_CONFIG")
	require.Contains(t, err.Error(), "report_timezone")
}

func TestSettingsRepository_RejectsInvalidTimezone(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "settings.yaml")

	err := os.WriteFile(filePath, []byte("storage_mode: local\nreport_timezone: Mars/Base\n"), 0o600)
	require.NoError(t, err)

	repo := NewSettingsRepository(filePath, true)
	_, err = repo.GetSettings(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "INVALID_TIMEZONE")
}

func TestSettingsRepository_RejectsNonLocalStorageMode(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "settings.yaml")

	err := os.WriteFile(filePath, []byte("storage_mode: remote\nreport_timezone: UTC\n"), 0o600)
	require.NoError(t, err)

	repo := NewSettingsRepository(filePath, true)
	_, err = repo.GetSettings(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "STORAGE_MODE_UNAVAILABLE")
}
