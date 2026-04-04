package bootstrap

import (
	"testing"
)

func TestParseCLIFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantMode string
		wantURL  string
		wantKey  string
		wantCfg  string
	}{
		{
			name:     "no flags",
			args:     []string{"infinita", "list"},
			wantMode: "",
			wantURL:  "",
			wantKey:  "",
			wantCfg:  "",
		},
		{
			name:     "mode flag",
			args:     []string{"infinita", "--mode", "remote", "list"},
			wantMode: "remote",
			wantURL:  "",
			wantKey:  "",
			wantCfg:  "",
		},
		{
			name:     "server-url flag",
			args:     []string{"infinita", "--server-url", "http://localhost:8080", "list"},
			wantMode: "",
			wantURL:  "http://localhost:8080",
			wantKey:  "",
			wantCfg:  "",
		},
		{
			name:     "api-key flag",
			args:     []string{"infinita", "--api-key", "secret123", "list"},
			wantMode: "",
			wantURL:  "",
			wantKey:  "secret123",
			wantCfg:  "",
		},
		{
			name:     "config flag",
			args:     []string{"infinita", "--config", "/path/to/config.yaml", "list"},
			wantMode: "",
			wantURL:  "",
			wantKey:  "",
			wantCfg:  "/path/to/config.yaml",
		},
		{
			name:     "all flags together",
			args:     []string{"infinita", "--mode", "remote", "--server-url", "http://localhost:8080", "--api-key", "secret", "--config", "/path/to/config.yaml"},
			wantMode: "remote",
			wantURL:  "http://localhost:8080",
			wantKey:  "secret",
			wantCfg:  "/path/to/config.yaml",
		},
		{
			name:     "flags with command",
			args:     []string{"infinita", "add", "--mode", "local", "--amount", "100"},
			wantMode: "local",
			wantURL:  "",
			wantKey:  "",
			wantCfg:  "",
		},
		{
			name:     "mode at end",
			args:     []string{"infinita", "list", "--mode", "local"},
			wantMode: "local",
			wantURL:  "",
			wantKey:  "",
			wantCfg:  "",
		},
		{
			name:     "mode without value",
			args:     []string{"infinita", "--mode"},
			wantMode: "",
			wantURL:  "",
			wantKey:  "",
			wantCfg:  "",
		},
		{
			name:     "server-url without value",
			args:     []string{"infinita", "--server-url"},
			wantMode: "",
			wantURL:  "",
			wantKey:  "",
			wantCfg:  "",
		},
		{
			name:     "api-key without value",
			args:     []string{"infinita", "--api-key"},
			wantMode: "",
			wantURL:  "",
			wantKey:  "",
			wantCfg:  "",
		},
		{
			name:     "config without value",
			args:     []string{"infinita", "--config"},
			wantMode: "",
			wantURL:  "",
			wantKey:  "",
			wantCfg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCLIFlags(tt.args)
			if got.mode != tt.wantMode {
				t.Errorf("parseCLIFlags().mode = %q, want %q", got.mode, tt.wantMode)
			}
			if got.serverURL != tt.wantURL {
				t.Errorf("parseCLIFlags().serverURL = %q, want %q", got.serverURL, tt.wantURL)
			}
			if got.apiKey != tt.wantKey {
				t.Errorf("parseCLIFlags().apiKey = %q, want %q", got.apiKey, tt.wantKey)
			}
			if got.config != tt.wantCfg {
				t.Errorf("parseCLIFlags().config = %q, want %q", got.config, tt.wantCfg)
			}
		})
	}
}
