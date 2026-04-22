package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

var ErrTemplateWritten = errors.New("config: template written; fill it in and re-run")

type Config struct {
	Emby      EmbyConfig      `yaml:"emby"`
	Download  DownloadConfig  `yaml:"download"`
	Subtitles SubtitlesConfig `yaml:"subtitles"`
	Versions  VersionsConfig  `yaml:"versions"`
	Logging   LoggingConfig   `yaml:"logging"`
}

type EmbyConfig struct {
	Endpoint string `yaml:"endpoint"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type DownloadConfig struct {
	OutputDir          string        `yaml:"output_dir"`
	MoviesSubdir       string        `yaml:"movies_subdir"`
	TVSubdir           string        `yaml:"tv_subdir"`
	InterDownloadDelay time.Duration `yaml:"inter_download_delay"`
	Jitter             time.Duration `yaml:"jitter"`
	MaxRetries         int           `yaml:"max_retries"`
	RetryBackoff       time.Duration `yaml:"retry_backoff"`
	UserAgent          string        `yaml:"user_agent"`
}

type SubtitlesConfig struct {
	PreferredLanguages []string `yaml:"preferred_languages"`
}

type VersionsConfig struct {
	ResolutionOrder []int    `yaml:"resolution_order"`
	KeywordBoost    []string `yaml:"keyword_boost"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

const templateYAML = `# emget configuration
emby:
  endpoint: https://your-emby.example.com
  username: your-username
  password: your-password

download:
  output_dir: ~/Media
  movies_subdir: Movies
  tv_subdir: TV
  inter_download_delay: 3s
  jitter: 2s
  max_retries: 3
  retry_backoff: 1s
  user_agent: emget/0.1.0

subtitles:
  preferred_languages: [zho, chi, eng]

versions:
  resolution_order: [2160, 1080, 720, 480]
  keyword_boost: [BluRay, REMUX, WEB-DL]

logging:
  level: info
  file: ~/.local/share/emget/emget.log
`

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if werr := writeTemplate(path); werr != nil {
			return nil, fmt.Errorf("config: write template: %w", werr)
		}
		return nil, ErrTemplateWritten
	}
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}
	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func writeTemplate(path string) error {
	if err := EnsureDir(dirname(path)); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(templateYAML), 0o600)
}

func dirname(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[:i]
		}
	}
	return "."
}

func (c *Config) applyDefaults() {
	if c.Download.MoviesSubdir == "" {
		c.Download.MoviesSubdir = "Movies"
	}
	if c.Download.TVSubdir == "" {
		c.Download.TVSubdir = "TV"
	}
	if c.Download.InterDownloadDelay == 0 {
		c.Download.InterDownloadDelay = 3 * time.Second
	}
	if c.Download.Jitter == 0 {
		c.Download.Jitter = 2 * time.Second
	}
	if c.Download.MaxRetries == 0 {
		c.Download.MaxRetries = 3
	}
	if c.Download.RetryBackoff == 0 {
		c.Download.RetryBackoff = 1 * time.Second
	}
	if c.Download.UserAgent == "" {
		c.Download.UserAgent = "emget/0.1.0"
	}
	if len(c.Subtitles.PreferredLanguages) == 0 {
		c.Subtitles.PreferredLanguages = []string{"zho", "chi", "eng"}
	}
	if len(c.Versions.ResolutionOrder) == 0 {
		c.Versions.ResolutionOrder = []int{2160, 1080, 720, 480}
	}
	if len(c.Versions.KeywordBoost) == 0 {
		c.Versions.KeywordBoost = []string{"BluRay", "REMUX", "WEB-DL"}
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.File == "" {
		if dir, err := DataDir(); err == nil {
			c.Logging.File = filepath.Join(dir, "emget.log")
		} else {
			c.Logging.File = "emget.log"
		}
	}
}

func (c *Config) validate() error {
	if c.Emby.Endpoint == "" {
		return errors.New("config: emby.endpoint is required")
	}
	if c.Emby.Username == "" {
		return errors.New("config: emby.username is required")
	}
	if c.Emby.Password == "" {
		return errors.New("config: emby.password is required")
	}
	if c.Download.OutputDir == "" {
		return errors.New("config: download.output_dir is required")
	}
	return nil
}
