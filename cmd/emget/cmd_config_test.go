package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunConfig_PathsOnly(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(`emby:
  endpoint: https://x
  username: u
  password: p
download:
  output_dir: ~/Media
`), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := runConfig(runConfigOpts{configPath: cfgPath, pathsOnly: true, stdout: &buf}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, cfgPath) {
		t.Errorf("expected config path in output, got: %s", out)
	}
	if strings.Contains(out, "password:") {
		t.Errorf("--paths-only should not print contents: %s", out)
	}
}

func TestRunConfig_Redacts(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(`emby:
  endpoint: https://x
  username: u
  password: secret123
download:
  output_dir: ~/Media
`), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := runConfig(runConfigOpts{configPath: cfgPath, stdout: &buf}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "secret123") {
		t.Errorf("password leaked: %s", out)
	}
	// Assert the password FIELD itself contains the redaction marker,
	// not just that **** appears anywhere in output.
	if !strings.Contains(out, "password: '****'") &&
		!strings.Contains(out, `password: "****"`) &&
		!strings.Contains(out, "password: ****") {
		t.Errorf("expected password field redacted to ****, got: %s", out)
	}
}
