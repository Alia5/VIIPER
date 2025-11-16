package configpaths

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// DefaultConfigDir returns the platform-specific configuration directory for VIIPER.
func DefaultConfigDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if appdata := os.Getenv("AppData"); appdata != "" {
			return filepath.Join(appdata, "VIIPER"), nil
		}
		return "", errors.New("AppData not set")
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "viiper"), nil
		}
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, ".config", "viiper"), nil
		}
		return "", errors.New("HOME not set")
	}
}

// DefaultConfigPath returns the default config file path for the given format using base name "config".
func DefaultConfigPath(format string) (string, error) {
	return DefaultNamedConfigPath("config", format)
}

// DefaultNamedConfigPath returns the default config file path for the given format and base name (e.g., "server").
func DefaultNamedConfigPath(baseName, format string) (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}
	ext := "json"
	switch format {
	case "yaml", "yml":
		ext = "yaml"
	case "toml":
		ext = "toml"
	}
	return filepath.Join(dir, baseName+"."+ext), nil
}

// EnsureDir ensures the directory for a given file path exists.
func EnsureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0o755)
}

// ConfigCandidatePaths builds candidate paths for config files per format.
// If userPath is provided, it is prioritized and routed to the matching loader by extension.
func ConfigCandidatePaths(userPath string) (jsonPaths, yamlPaths, tomlPaths []string) {
	add := func(slice *[]string, p string) { *slice = append(*slice, p) }

	if userPath != "" {
		switch ext := filepath.Ext(userPath); ext {
		case ".json":
			add(&jsonPaths, userPath)
		case ".yaml", ".yml":
			add(&yamlPaths, userPath)
		case ".toml":
			add(&tomlPaths, userPath)
		default:
			add(&jsonPaths, userPath)
		}
	}

	// Working directory candidates
	wd, _ := os.Getwd()
	for _, base := range []string{"viiper", "config", "server", "proxy"} {
		add(&jsonPaths, filepath.Join(wd, base+".json"))
		add(&yamlPaths, filepath.Join(wd, base+".yaml"))
		add(&yamlPaths, filepath.Join(wd, base+".yml"))
		add(&tomlPaths, filepath.Join(wd, base+".toml"))
	}

	// Config home
	if dir, err := DefaultConfigDir(); err == nil {
		for _, base := range []string{"config", "server", "proxy"} {
			add(&jsonPaths, filepath.Join(dir, base+".json"))
			add(&yamlPaths, filepath.Join(dir, base+".yaml"))
			add(&yamlPaths, filepath.Join(dir, base+".yml"))
			add(&tomlPaths, filepath.Join(dir, base+".toml"))
		}
	}

	// System-wide (unix)
	if runtime.GOOS != "windows" {
		for _, base := range []string{"config", "server", "proxy"} {
			add(&jsonPaths, filepath.Join("/etc/viiper", base+".json"))
			add(&yamlPaths, filepath.Join("/etc/viiper", base+".yaml"))
			add(&yamlPaths, filepath.Join("/etc/viiper", base+".yml"))
			add(&tomlPaths, filepath.Join("/etc/viiper", base+".toml"))
		}
	}

	return
}
