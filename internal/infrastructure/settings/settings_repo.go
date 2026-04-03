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
	StorageMode:    "local",
	ReportTimezone: "Asia/Jakarta",
}

type yamlSettings struct {
	StorageMode    string `yaml:"storage_mode"`
	ReportTimezone string `yaml:"report_timezone"`
}

// SettingsRepository implements output.SettingsRepository using a YAML file.
type SettingsRepository struct {
	filePath string
}

func NewSettingsRepository(filePath string) *SettingsRepository {
	return &SettingsRepository{
		filePath: filePath,
	}
}

func (r *SettingsRepository) GetSettings(ctx context.Context) (entity.Settings, error) {
	s, err := r.readOrCreate(ctx)
	if err != nil {
		return entity.Settings{}, err
	}
	return entity.Settings{
		StorageMode:    s.StorageMode,
		ReportTimezone: s.ReportTimezone,
	}, nil
}

func (r *SettingsRepository) SetStorageMode(_ context.Context, mode string) error {
	if mode != "local" {
		return domainerror.ErrInvalidStorageMode.WithField("storage_mode").WithHint("storage mode must remain 'local' in MVP")
	}
	s, err := r.read()
	if err != nil {
		return err
	}
	s.StorageMode = mode
	return r.write(s)
}

func (r *SettingsRepository) SetReportTimezone(_ context.Context, timezone string) error {
	s, err := r.read()
	if err != nil {
		return err
	}
	s.ReportTimezone = timezone
	return r.write(s)
}

// readOrCreate reads the settings file. If it does not exist, it writes
// defaults to disk and returns them.
func (r *SettingsRepository) readOrCreate(_ context.Context) (yamlSettings, error) {
	s, err := r.read()
	if err == nil {
		return s, nil
	}
	if !os.IsNotExist(err) {
		return yamlSettings{}, err
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
	if s.StorageMode == "" {
		return domainerror.ErrInvalidConfig.WithField("storage_mode").WithHint("storage_mode is required")
	}
	if s.StorageMode != "local" {
		return domainerror.ErrInvalidStorageMode.WithField("storage_mode").WithHint("storage mode must remain 'local' in MVP")
	}

	if s.ReportTimezone == "" {
		return domainerror.ErrInvalidConfig.WithField("report_timezone").WithHint("report_timezone is required")
	}
	if _, err := time.LoadLocation(s.ReportTimezone); err != nil {
		return domainerror.ErrInvalidTimezone.WithField("report_timezone").WithHint("provide a valid IANA timezone")
	}

	return nil
}
