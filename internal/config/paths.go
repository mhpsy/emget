package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

const appName = "emget"

// ConfigDir returns the per-user config directory for emget.
//
//	Linux:   $XDG_CONFIG_HOME/emget   or   $HOME/.config/emget
//	macOS:   $HOME/Library/Application Support/emget
//	Windows: %AppData%\emget
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName), nil
}

// CacheDir returns the per-user cache directory for emget.
//
//	Linux:   $XDG_CACHE_HOME/emget   or   $HOME/.cache/emget
//	macOS:   $HOME/Library/Caches/emget
//	Windows: %LocalAppData%\emget\cache
func CacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(base, appName, "cache"), nil
	}
	return filepath.Join(base, appName), nil
}

// DataDir returns the per-user persistent data directory for emget.
//
//	Linux:   $XDG_DATA_HOME/emget   or   $HOME/.local/share/emget
//	macOS:   $HOME/Library/Application Support/emget
//	Windows: %AppData%\emget
func DataDir() (string, error) {
	switch runtime.GOOS {
	case "darwin", "windows":
		return ConfigDir()
	default:
		if v := os.Getenv("XDG_DATA_HOME"); v != "" {
			return filepath.Join(v, appName), nil
		}
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return "", errors.New("DataDir: HOME not set")
		}
		return filepath.Join(home, ".local", "share", appName), nil
	}
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
