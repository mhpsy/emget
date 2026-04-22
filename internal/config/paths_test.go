package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfigDir_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-config")
	t.Setenv("HOME", "/tmp/home")
	got, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join("/tmp/xdg-config", "emget") {
		t.Errorf("ConfigDir() = %q", got)
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	got, err = ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join("/tmp/home", ".config", "emget") {
		t.Errorf("ConfigDir() fallback = %q", got)
	}
}

func TestDataDir_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", "/tmp/home")
	got, err := DataDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/tmp/home", ".local", "share", "emget")
	if got != want {
		t.Errorf("DataDir() = %q want %q", got, want)
	}
}

func TestCacheDir_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	t.Setenv("XDG_CACHE_HOME", "/tmp/xdg-cache")
	t.Setenv("HOME", "/tmp/home")
	got, err := CacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join("/tmp/xdg-cache", "emget") {
		t.Errorf("CacheDir() = %q", got)
	}
}

func TestDataDir_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin only")
	}
	t.Setenv("HOME", "/Users/test")
	got, err := DataDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/Users/test", "Library", "Application Support", "emget")
	if got != want {
		t.Errorf("DataDir() = %q want %q", got, want)
	}
}
