package cli_test

import (
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestCLIDisplaysListHeader(t *testing.T) {
	bin := buildCLIBinary(t)
	dataDir := filepath.Join(t.TempDir(), "data")
	cmd := exec.Command(bin, "list")
	cmd.Env = append(os.Environ(), "INFINITA_DATA_DIR="+dataDir)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	require.Contains(t, string(output), "ID       Date")
}

func TestCLIIntegrationCommands(t *testing.T) {
	bin := buildCLIBinary(t)
	dataDir := filepath.Join(t.TempDir(), "data")

	out, code := runCLI(t, bin, dataDir, "add", "--type", "expense", "--amount", "123.45", "--category", "Food", "--date", "2024-01-15", "--description", "lunch")
	require.Equal(t, 0, code)
	require.Contains(t, out, "Transaction recorded")

	out, code = runCLI(t, bin, dataDir, "list")
	require.Equal(t, 0, code)
	require.Contains(t, out, "Food")

	out, code = runCLI(t, bin, dataDir, "budget", "set", "--category", "Food", "--amount", "400.00", "--month", "2024-01")
	require.Equal(t, 0, code)
	require.Contains(t, out, "Budget stored")

	out, code = runCLI(t, bin, dataDir, "budget", "status", "--month", "2024-01")
	require.Equal(t, 0, code)
	require.Contains(t, out, "Food")

	out, code = runCLI(t, bin, dataDir, "report", "daily", "--date", "2024-01-15")
	require.Equal(t, 0, code)
	require.Contains(t, out, "Daily report 2024-01-15")

	out, code = runCLI(t, bin, dataDir, "report", "monthly", "--month", "2024-01")
	require.Equal(t, 0, code)
	require.Contains(t, out, "Monthly report 2024-01")

	out, code = runCLI(t, bin, dataDir, "settings", "show")
	require.Equal(t, 0, code)
	require.Contains(t, out, "Storage mode")

	out, code = runCLI(t, bin, dataDir, "settings", "set-initial-balance", "--amount", "500.00")
	require.Equal(t, 0, code)
	require.Contains(t, out, "Initial balance updated")

	out, code = runCLI(t, bin, dataDir, "settings", "reset-initial-balance")
	require.Equal(t, 0, code)
	require.Contains(t, out, "Initial balance reset")

	dbPath := filepath.Join(dataDir, "infinita.db")
	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close sqlite db: %v", err)
		}
	})

	var initialBalanceMinor int64
	err = db.QueryRow(`SELECT initial_balance_minor FROM initial_balance WHERE id = 1`).Scan(&initialBalanceMinor)
	require.NoError(t, err)
	require.Zero(t, initialBalanceMinor)

	out, code = runCLI(t, bin, dataDir, "settings", "report-timezone", "--timezone", "UTC")
	require.Equal(t, 0, code)
	require.Contains(t, out, "Report timezone updated")

	out, code = runCLI(t, bin, dataDir, "add", "--type", "expense", "--amount", "10.00", "--category", "Food", "--date", "2024-01-16")
	require.Equal(t, 0, code)
	require.Contains(t, out, "Transaction recorded")

	out, code = runCLI(t, bin, dataDir, "list")
	require.Equal(t, 0, code)
	require.Contains(t, out, "2024-01-16")
}

func TestCLIValidationExitCode(t *testing.T) {
	bin := buildCLIBinary(t)
	dataDir := filepath.Join(t.TempDir(), "data")

	_, code := runCLI(t, bin, dataDir, "add", "--type", "expense", "--amount", "abc", "--category", "Food", "--date", "2024-01-15")
	require.Equal(t, 2, code)

	out, code := runCLI(t, bin, dataDir, "unknown")
	require.Equal(t, 2, code)
	require.Contains(t, out, "unknown command")

	out, code = runCLI(t, bin, dataDir, "settings", "report-timezone", "--timezone", "Mars/Base")
	require.Equal(t, 2, code)
	require.Contains(t, out, "INVALID_TIMEZONE")
}

func TestCLIRuntimeExitCode(t *testing.T) {
	bin := buildCLIBinary(t)
	baseDir := t.TempDir()
	blockedPath := filepath.Join(baseDir, "blocked")
	require.NoError(t, os.WriteFile(blockedPath, []byte("not a directory"), 0o600))

	_, code := runCLI(t, bin, blockedPath, "list")
	require.Equal(t, 3, code)
}

func TestCLIRejectsNonLocalStorageMode(t *testing.T) {
	bin := buildCLIBinary(t)
	dataDir := filepath.Join(t.TempDir(), "data")

	_, code := runCLI(t, bin, dataDir, "list")
	require.Equal(t, 0, code)

	settingsPath := filepath.Join(dataDir, "settings.yaml")
	err := os.WriteFile(settingsPath, []byte("storage_mode: remote\nreport_timezone: Asia/Jakarta\n"), 0o600)
	require.NoError(t, err)

	out, code := runCLI(t, bin, dataDir, "list")
	require.Equal(t, 3, code)
	require.Contains(t, out, "STORAGE_MODE_UNAVAILABLE")
}

func runCLI(t *testing.T, bin, dataDir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "INFINITA_DATA_DIR="+dataDir)
	output, _ := cmd.CombinedOutput()
	code := 0
	if cmd.ProcessState != nil {
		code = cmd.ProcessState.ExitCode()
	}
	return string(output), code
}

func buildCLIBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "infinita")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/cli")
	cmd.Dir = filepath.Join("..", "..", "..")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build cli: %v - %s", err, output)
	}
	return bin
}
