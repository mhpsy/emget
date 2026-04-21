package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupLogger_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	cleanup, err := SetupLogger(logPath, "debug")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	SLogger().Info("hello", "k", "v")
	cleanup()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "hello") {
		t.Errorf("log does not contain message: %s", data)
	}
	if !strings.Contains(string(data), `"k":"v"`) {
		t.Errorf("log does not contain attr: %s", data)
	}
}

func TestSetupLogger_RotatesWhenLargerThan10MiB(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "emget.log")
	big := make([]byte, 10*1024*1024+1)
	if err := os.WriteFile(path, big, 0o644); err != nil {
		t.Fatal(err)
	}

	cleanup, err := SetupLogger(path, "info")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	rotatedInfo, err := os.Stat(path + ".1")
	if err != nil {
		t.Fatalf("rotated file missing: %v", err)
	}
	if rotatedInfo.Size() < int64(len(big)) {
		t.Errorf("rotated size = %d, want >= %d", rotatedInfo.Size(), len(big))
	}
	currentInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if currentInfo.Size() >= int64(len(big)) {
		t.Errorf("current log should be small, got %d", currentInfo.Size())
	}
}

func TestSetupLogger_NoRotateWhenSmall(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "emget.log")
	if err := os.WriteFile(path, []byte("existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cleanup, err := SetupLogger(path, "info")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if _, err := os.Stat(path + ".1"); !os.IsNotExist(err) {
		t.Errorf(".1 should not exist when log is small (err=%v)", err)
	}
}
