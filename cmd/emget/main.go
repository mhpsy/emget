package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/mhpsy/emget/internal/config"
	"github.com/mhpsy/emget/internal/downloader"
	"github.com/mhpsy/emget/internal/emby"
	"github.com/mhpsy/emget/internal/state"
	"github.com/mhpsy/emget/internal/tui"
)

func main() {
	if err := dispatch(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "emget:", err)
		os.Exit(1)
	}
}

func dispatch(args []string) error {
	if len(args) == 0 {
		return runTUI("")
	}
	cmd := args[0]
	rest := args[1:]
	switch cmd {
	case "tui":
		fs := flag.NewFlagSet("tui", flag.ExitOnError)
		configPath := fs.String("config", "", "path to config.yaml")
		_ = fs.Parse(rest)
		return runTUI(*configPath)
	case "version", "-v", "--version":
		return runVersion(os.Stdout)
	case "help", "-h", "--help":
		printUsage(os.Stdout)
		return nil
	case "config":
		fs := flag.NewFlagSet("config", flag.ExitOnError)
		configPath := fs.String("config", "", "path to config.yaml")
		pathsOnly := fs.Bool("paths-only", false, "print only paths, not contents")
		raw := fs.Bool("raw", false, "print raw file (password NOT redacted)")
		_ = fs.Parse(rest)
		cp := *configPath
		if cp == "" {
			dir, err := config.ConfigDir()
			if err != nil {
				return err
			}
			cp = filepath.Join(dir, "config.yaml")
		}
		return runConfig(runConfigOpts{configPath: cp, pathsOnly: *pathsOnly, raw: *raw, stdout: os.Stdout})
	default:
		printUsage(os.Stderr)
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func runTUI(configPath string) error {
	if configPath == "" {
		dir, err := config.ConfigDir()
		if err != nil {
			return err
		}
		configPath = filepath.Join(dir, "config.yaml")
	}

	cfg, err := config.Load(configPath)
	if errors.Is(err, config.ErrTemplateWritten) {
		fmt.Fprintf(os.Stderr, "emget: wrote template to %s — fill it in and re-run\n", configPath)
		return nil
	}
	if err != nil {
		return err
	}

	cleanup, err := config.SetupLogger(expandHome(cfg.Logging.File), cfg.Logging.Level)
	if err != nil {
		return fmt.Errorf("setup logger: %w", err)
	}
	defer cleanup()

	log := config.SLogger()
	log.Info("starting emget", "endpoint", cfg.Emby.Endpoint)

	// Context with signal cancellation
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Emby client + session
	httpClient := &http.Client{Timeout: 60 * time.Second}
	client := emby.NewClient(cfg.Emby.Endpoint, httpClient)

	sessionPath, err := sessionFilePath()
	if err != nil {
		return err
	}
	// Session handling:
	//   - Load cached session from disk if not expired.
	//   - If cache missing or expired, authenticate with config credentials.
	//   - In-TUI 401s (token revoked mid-session) are logged but not auto-recovered
	//     in v0.3 — surface a flash and let the user restart. Auto-recovery is
	//     tracked for a future release.
	if sess, err := config.LoadSession(sessionPath); err == nil && sess.ExpiresAt.After(time.Now()) {
		client.SetSession(embySession(sess))
		log.Info("loaded existing session", "user_id", sess.UserID, "expires_at", sess.ExpiresAt)
	} else {
		if err == nil {
			log.Info("session expired, will reauthenticate", "expired_at", sess.ExpiresAt)
			client.SetDeviceID(sess.DeviceID)
		} else {
			client.SetDeviceID(uuid.NewString())
		}
	}

	if !client.IsAuthenticated() {
		if err := authenticateAndSave(ctx, client, cfg, sessionPath); err != nil {
			return err
		}
	}

	// State store
	dataDir, err := config.DataDir()
	if err != nil {
		return err
	}
	store := state.NewStore(filepath.Join(dataDir, "state.json"))
	if err := store.Load(); err != nil {
		log.Error("state load failed", "err", err)
	}

	// Downloader + Queue
	dl := downloader.New(httpClient)
	dl.SetUserAgent(cfg.Download.UserAgent)
	queue := downloader.NewQueue(downloader.QueueConfig{
		Runner:     dl,
		Store:      store,
		Delay:      cfg.Download.InterDownloadDelay,
		Jitter:     cfg.Download.Jitter,
		MaxRetries: cfg.Download.MaxRetries,
		RetryBase:  cfg.Download.RetryBackoff,
	})
	queue.Start(ctx)
	defer queue.Stop()

	// Expand output dir
	cfg.Download.OutputDir = expandHome(cfg.Download.OutputDir)

	// TUI
	app := tui.NewApp(ctx, client, queue, store, tui.AppConfig{
		OutputDir:       cfg.Download.OutputDir,
		MoviesSubdir:    cfg.Download.MoviesSubdir,
		TVSubdir:        cfg.Download.TVSubdir,
		Languages:       cfg.Subtitles.PreferredLanguages,
		ResolutionOrder: cfg.Versions.ResolutionOrder,
		KeywordBoost:    cfg.Versions.KeywordBoost,
	})

	// Startup recovery: if state has unfinished tasks, show the recovery screen first.
	pending := unfinishedTasks(store)
	if len(pending) > 0 {
		log.Info("unfinished tasks detected", "count", len(pending))
		app.SetInitialScreen(tui.ScreenStartup())
		app.SeedStartup(pending)
	}

	program := tea.NewProgram(app, tea.WithContext(ctx))
	_, err = program.Run()
	if err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	log.Info("emget shutdown clean")
	return nil
}

func authenticateAndSave(ctx context.Context, client *emby.Client, cfg *config.Config, sessionPath string) error {
	log := config.SLogger()
	log.Info("authenticating", "username", cfg.Emby.Username)
	sess, err := client.Authenticate(ctx, cfg.Emby.Username, cfg.Emby.Password)
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}
	return config.SaveSession(sessionPath, &config.Session{
		AccessToken: sess.AccessToken,
		UserID:      sess.UserID,
		ServerID:    sess.ServerID,
		DeviceID:    sess.DeviceID,
		ExpiresAt:   sess.ExpiresAt,
	})
}

func sessionFilePath() (string, error) {
	dir, err := config.CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

func embySession(s *config.Session) *emby.Session {
	return &emby.Session{
		AccessToken: s.AccessToken,
		UserID:      s.UserID,
		ServerID:    s.ServerID,
		DeviceID:    s.DeviceID,
		ExpiresAt:   s.ExpiresAt,
	}
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		if h := os.Getenv("HOME"); h != "" {
			return filepath.Join(h, path[2:])
		}
	}
	return path
}

func unfinishedTasks(store *state.Store) []*downloader.Task {
	var out []*downloader.Task
	for _, t := range store.Tasks() {
		if t.Status != downloader.StatusCompleted {
			out = append(out, t)
		}
	}
	return out
}
