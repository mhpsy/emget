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
