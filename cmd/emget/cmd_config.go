package main

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/mhpsy/emget/internal/config"
)

type runConfigOpts struct {
	configPath string
	pathsOnly  bool
	raw        bool
	stdout     io.Writer
}

func runConfig(opts runConfigOpts) error {
	out := opts.stdout
	if out == nil {
		out = os.Stdout
	}

	configDir, _ := config.ConfigDir()
	cacheDir, _ := config.CacheDir()
	dataDir, _ := config.DataDir()

	fmt.Fprintf(out, "Config file: %s\n", opts.configPath)
	fmt.Fprintf(out, "Config dir:  %s\n", configDir)
	fmt.Fprintf(out, "Cache dir:   %s\n", cacheDir)
	fmt.Fprintf(out, "Data dir:    %s\n", dataDir)
	if opts.pathsOnly {
		return nil
	}

	fmt.Fprintln(out)
	if opts.raw {
		fmt.Fprintln(out, "# WARNING: raw mode — password is NOT redacted")
		data, err := os.ReadFile(opts.configPath)
		if err != nil {
			return fmt.Errorf("read raw config: %w", err)
		}
		out.Write(data)
		return nil
	}

	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.Emby.Password = "****"
	buf, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	out.Write(buf)
	return nil
}
