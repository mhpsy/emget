package config

import (
	"errors"
	"os"
	"path/filepath"
)

const appName = "emget"

func ConfigDir() (string, error) { return xdgDir("XDG_CONFIG_HOME", ".config") }
func CacheDir() (string, error)  { return xdgDir("XDG_CACHE_HOME", ".cache") }
func DataDir() (string, error)   { return xdgDir("XDG_DATA_HOME", ".local/share") }

func xdgDir(envVar, homeSubdir string) (string, error) {
	if v := os.Getenv(envVar); v != "" {
		return filepath.Join(v, appName), nil
	}
	home := os.Getenv("HOME")
	if home == "" {
		return "", errors.New("neither " + envVar + " nor HOME is set")
	}
	return filepath.Join(home, homeSubdir, appName), nil
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
