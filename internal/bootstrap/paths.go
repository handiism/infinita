package bootstrap

import "strings"

// Paths holds the resolved filesystem paths used by the application.
type Paths struct {
	DataDir         string
	SettingsFile    string
	IsDefaultConfig bool
}

// ResolvePaths resolves the data and settings paths from CLI arguments and environment.
func ResolvePaths(args []string) (Paths, error) {
	dataDir, err := ResolveDataDir()
	if err != nil {
		return Paths{}, err
	}

	return ResolvePathsFromDataDir(dataDir, args), nil
}

// ResolvePathsFromDataDir resolves paths using the provided dataDir, with optional CLI overrides.
func ResolvePathsFromDataDir(dataDir string, args []string) Paths {
	settingsFile := ResolveSettingsFile(dataDir)
	isDefaultConfig := true
	if override, ok := parseConfigArg(args); ok {
		settingsFile = override
		isDefaultConfig = false
	}

	return Paths{
		DataDir:         dataDir,
		SettingsFile:    settingsFile,
		IsDefaultConfig: isDefaultConfig,
	}
}

// parseConfigArg extracts --config value from CLI args, returning the path and true,
// or empty string and false if not found. Supports both "--config value" and "--config=value" forms.
func parseConfigArg(args []string) (string, bool) {
	for i, arg := range args {
		if arg == "--config" && i+1 < len(args) {
			return args[i+1], true
		}
		if after, ok := strings.CutPrefix(arg, "--config="); ok {
			return after, true
		}
	}

	return "", false
}
