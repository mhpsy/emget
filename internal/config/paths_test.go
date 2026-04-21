package config

import (
	"path/filepath"
	"testing"
)

func TestConfigDir_UsesXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-config")
	t.Setenv("HOME", "/home/irrelevant")
	got, err := ConfigDir()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	want := filepath.Join("/tmp/xdg-config", "emget")
	if got != want {
		t.Fatalf("ConfigDir = %q, want %q", got, want)
	}
}

func TestConfigDir_FallsBackToHomeDotConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/home/u")
	got, _ := ConfigDir()
	want := "/home/u/.config/emget"
	if got != want {
		t.Fatalf("ConfigDir = %q, want %q", got, want)
	}
}

func TestCacheDir_UsesXDGCacheHome(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "/tmp/xdg-cache")
	got, _ := CacheDir()
	want := "/tmp/xdg-cache/emget"
	if got != want {
		t.Fatalf("CacheDir = %q, want %q", got, want)
	}
}

func TestCacheDir_FallsBackToHomeDotCache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("HOME", "/home/u")
	got, _ := CacheDir()
	want := "/home/u/.cache/emget"
	if got != want {
		t.Fatalf("CacheDir = %q, want %q", got, want)
	}
}

func TestDataDir_UsesXDGDataHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/xdg-data")
	got, _ := DataDir()
	want := "/tmp/xdg-data/emget"
	if got != want {
		t.Fatalf("DataDir = %q, want %q", got, want)
	}
}

func TestDataDir_FallsBackToHomeDotLocalShare(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", "/home/u")
	got, _ := DataDir()
	want := "/home/u/.local/share/emget"
	if got != want {
		t.Fatalf("DataDir = %q, want %q", got, want)
	}
}
