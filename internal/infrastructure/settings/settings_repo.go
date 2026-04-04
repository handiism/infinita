package settings

import (
	"context"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
)

var defaults = yamlSettings{
	Mode:           "local",
	ServerURL:      "",
	ReportTimezone: "Asia/Jakarta",
	APIKey:         "",
}

type yamlSettings struct {
	Mode           string `yaml:"mode"`
	ServerURL      string `yaml:"server_url"`
	ReportTimezone string `yaml:"report_timezone"`
	APIKey         string `yaml:"api_key"`
}

// SettingsRepository implements output.SettingsRepository using a YAML file.
type SettingsRepository struct {
	filePath      string
	isDefaultPath bool
}

func NewSettingsRepository(filePath string, isDefaultPath bool) *SettingsRepository {
	return &SettingsRepository{
		filePath:      filePath,
		isDefaultPath: isDefaultPath,
	}
}

func (r *SettingsRepository) GetSettings(ctx context.Context) (entity.Settings, error) {
	s, err := r.readOrCreate(ctx)
	if err != nil {
		return entity.Settings{}, err
	}
	return entity.Settings{
		Mode:           s.Mode,
		ServerURL:      s.ServerURL,
		ReportTimezone: s.ReportTimezone,
		APIKey:         s.APIKey,
	}, nil
}

func (r *SettingsRepository) SetMode(_ context.Context, mode string) error {
	if mode != "local" && mode != "remote" {
		return domainerror.ErrInvalidStorageMode.WithField("mode")
	}
	s, err := r.read()
	if err != nil {
		return err
	}
	s.Mode = mode
	return r.write(s)
}

// readOrCreate reads the settings file. If it does not exist and the path
// is the default one, it writes defaults to disk and returns them.
func (r *SettingsRepository) readOrCreate(_ context.Context) (yamlSettings, error) {
	s, err := r.read()
	if err == nil {
		return s, nil
	}
	if !os.IsNotExist(err) {
		return yamlSettings{}, err
	}
	if !r.isDefaultPath {
		return yamlSettings{}, fmt.Errorf("settings file not found: %s", r.filePath)
	}
	if writeErr := r.write(defaults); writeErr != nil {
		return yamlSettings{}, fmt.Errorf("create default settings: %w", writeErr)
	}
	return defaults, nil
}

func (r *SettingsRepository) read() (yamlSettings, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return yamlSettings{}, err
	}

	var s yamlSettings
	if err := yaml.Unmarshal(data, &s); err != nil {
		return yamlSettings{}, domainerror.ErrInvalidConfig.WithHint(fmt.Sprintf("YAML parse error: %v", err))
	}

	if err := validate(s); err != nil {
		return yamlSettings{}, err
	}

	return s, nil
}

func (r *SettingsRepository) write(s yamlSettings) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(r.filePath, data, 0o600)
}

func validate(s yamlSettings) error {
	if s.Mode == "" {
		return domainerror.ErrInvalidConfig.WithField("mode").WithHint("mode is required")
	}
	if s.Mode != "local" && s.Mode != "remote" {
		return domainerror.ErrInvalidStorageMode.WithField("mode")
	}
	if s.Mode == "remote" && s.ServerURL == "" {
		return domainerror.ErrInvalidConfig.WithField("server_url").WithHint("server_url is required for remote mode")
	}
	if s.Mode == "remote" && s.APIKey == "" {
		return domainerror.ErrMissingAPIKey.WithField("api_key").WithHint("api_key is required for remote mode")
	}

	if s.ReportTimezone == "" {
		return domainerror.ErrInvalidConfig.WithField("report_timezone").WithHint("report_timezone is required")
	}
	if _, err := time.LoadLocation(s.ReportTimezone); err != nil {
		return domainerror.ErrInvalidTimezone.WithField("report_timezone").WithHint("provide a valid IANA timezone")
	}

	return nil
}
