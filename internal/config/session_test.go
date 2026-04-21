package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSession_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")
	s := Session{
		AccessToken: "tok",
		UserID:      "uid",
		ServerID:    "sid",
		DeviceID:    "did",
		ExpiresAt:   time.Now().Add(1 * time.Hour).UTC().Truncate(time.Second),
	}
	if err := SaveSession(path, &s); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("perm = %v, want 0o600", info.Mode().Perm())
	}
	got, err := LoadSession(path)
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	if got.AccessToken != s.AccessToken || got.UserID != s.UserID || got.DeviceID != s.DeviceID {
		t.Errorf("roundtrip mismatch: %+v vs %+v", got, s)
	}
	if !got.ExpiresAt.Equal(s.ExpiresAt) {
		t.Errorf("expiresAt mismatch")
	}
}

func TestLoadSession_Missing(t *testing.T) {
	_, err := LoadSession(filepath.Join(t.TempDir(), "nope.json"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("want IsNotExist, got %v", err)
	}
}
