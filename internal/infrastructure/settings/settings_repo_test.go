package settings

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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
	require.Equal(t, "local", settings.Mode)
	require.Equal(t, "", settings.ServerURL)
	require.Equal(t, "Asia/Jakarta", settings.ReportTimezone)

	// File now exists on disk.
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Contains(t, string(data), "mode: local")
	require.Contains(t, string(data), "report_timezone: Asia/Jakarta")

	// Second read uses the file, not auto-create again.
	settings, err = repo.GetSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, "local", settings.Mode)
	require.Equal(t, "", settings.ServerURL)
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

	err := os.WriteFile(filePath, []byte("mode: local\nreport_timezone: Europe/London\nserver_url: \"\"\n"), 0o600)
	require.NoError(t, err)

	repo := NewSettingsRepository(filePath, true)
	settings, err := repo.GetSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, "local", settings.Mode)
	require.Equal(t, "", settings.ServerURL)
	require.Equal(t, "Europe/London", settings.ReportTimezone)
}

func TestSettingsRepository_RejectsInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "settings.yaml")

	err := os.WriteFile(filePath, []byte("mode: [broken\n"), 0o600)
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
	require.Contains(t, err.Error(), "mode")
}

func TestSettingsRepository_RejectsMissingTimezone(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "settings.yaml")

	err := os.WriteFile(filePath, []byte("mode: local\n"), 0o600)
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

	err := os.WriteFile(filePath, []byte("mode: local\nreport_timezone: Mars/Base\n"), 0o600)
	require.NoError(t, err)

	repo := NewSettingsRepository(filePath, true)
	_, err = repo.GetSettings(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "INVALID_TIMEZONE")
}

func TestSettingsRepository_RejectsInvalidStorageMode(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     string
	}{
		{
			name:        "invalid mode is rejected",
			yamlContent: "mode: invalid\nreport_timezone: UTC\nserver_url: \"\"\napi_key: \"\"\n",
			wantErr:     "MODE_UNAVAILABLE",
		},
		{
			name:        "remote without server_url is rejected",
			yamlContent: "mode: remote\nreport_timezone: UTC\nserver_url: \"\"\napi_key: \"secret\"\n",
			wantErr:     "INVALID_CONFIG",
		},
		{
			name:        "remote without api_key is rejected",
			yamlContent: "mode: remote\nreport_timezone: UTC\nserver_url: http://localhost:8080\napi_key: \"\"\n",
			wantErr:     "MISSING_API_KEY",
		},
		{
			name:        "remote with server_url and api_key is accepted",
			yamlContent: "mode: remote\nreport_timezone: UTC\nserver_url: http://localhost:8080\napi_key: \"secret\"\n",
			wantErr:     "",
		},
		{
			name:        "local with api_key is accepted (optional for local)",
			yamlContent: "mode: local\nreport_timezone: UTC\nserver_url: \"\"\napi_key: \"optional-key\"\n",
			wantErr:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			filePath := filepath.Join(dir, "settings.yaml")

			err := os.WriteFile(filePath, []byte(tt.yamlContent), 0o600)
			require.NoError(t, err)

			repo := NewSettingsRepository(filePath, true)
			settings, err := repo.GetSettings(context.Background())

			if tt.wantErr == "" {
				require.NoError(t, err)
				if strings.Contains(tt.yamlContent, "mode: remote") {
					require.Equal(t, "remote", settings.Mode)
					require.Equal(t, "http://localhost:8080", settings.ServerURL)
					require.Equal(t, "secret", settings.APIKey)
				} else {
					require.Equal(t, "local", settings.Mode)
					require.Equal(t, "optional-key", settings.APIKey)
				}
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}
