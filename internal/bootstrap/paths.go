package bootstrap

import "strings"

type Paths struct {
	DataDir      string
	SettingsFile string
}

func ResolvePaths(args []string) (Paths, error) {
	dataDir, err := ResolveDataDir()
	if err != nil {
		return Paths{}, err
	}

	return ResolvePathsFromDataDir(dataDir, args), nil
}

func ResolvePathsFromDataDir(dataDir string, args []string) Paths {
	settingsFile := ResolveSettingsFile(dataDir)
	if override, ok := resolveSettingsFileArg(args); ok {
		settingsFile = override
	}

	return Paths{
		DataDir:      dataDir,
		SettingsFile: settingsFile,
	}
}

func resolveSettingsFileArg(args []string) (string, bool) {
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
