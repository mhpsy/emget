package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_MissingFileWritesTemplateAndReturnsErrMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	_, err := Load(path)
	if err == nil || err != ErrTemplateWritten {
		t.Fatalf("want ErrTemplateWritten, got %v", err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("template not written: %v", statErr)
	}
	data, _ := os.ReadFile(path)
	if len(data) == 0 {
		t.Fatal("template file is empty")
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
emby:
  endpoint: https://example.com
  username: user
  password: pw
download:
  output_dir: /tmp/media
  movies_subdir: Movies
  tv_subdir: TV
  inter_download_delay: 3s
  jitter: 2s
  max_retries: 3
  retry_backoff: 1s
  user_agent: emget/0.1.0
subtitles:
  preferred_languages: [zho, eng]
versions:
  resolution_order: [2160, 1080, 720]
  keyword_boost: [BluRay, REMUX]
logging:
  level: info
  file: /tmp/emget.log
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Emby.Endpoint != "https://example.com" {
		t.Errorf("endpoint: %q", cfg.Emby.Endpoint)
	}
	if cfg.Download.InterDownloadDelay != 3*time.Second {
		t.Errorf("delay: %v", cfg.Download.InterDownloadDelay)
	}
	if cfg.Download.Jitter != 2*time.Second {
		t.Errorf("jitter: %v", cfg.Download.Jitter)
	}
	if len(cfg.Versions.ResolutionOrder) != 3 {
		t.Errorf("resolution order: %v", cfg.Versions.ResolutionOrder)
	}
}

func TestLoad_AppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
emby:
  endpoint: https://example.com
  username: user
  password: pw
download:
  output_dir: /tmp/media
`
	os.WriteFile(path, []byte(content), 0o600)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Download.MoviesSubdir != "Movies" {
		t.Errorf("movies default: %q", cfg.Download.MoviesSubdir)
	}
	if cfg.Download.InterDownloadDelay != 3*time.Second {
		t.Errorf("delay default: %v", cfg.Download.InterDownloadDelay)
	}
	if cfg.Download.MaxRetries != 3 {
		t.Errorf("max_retries default: %d", cfg.Download.MaxRetries)
	}
	if cfg.Download.UserAgent == "" {
		t.Error("user_agent default empty")
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("log level default: %q", cfg.Logging.Level)
	}
}

func TestLoad_RequiredFieldsMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("emby: {endpoint: ''}"), 0o600)

	if _, err := Load(path); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}
