package config

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	loggerMu sync.Mutex
	logger   *slog.Logger
	logFile  *os.File
)

func SetupLogger(filePath, level string) (cleanup func(), err error) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	if err := EnsureDir(dirname(filePath)); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	opts := &slog.HandlerOptions{Level: parseLevel(level)}
	h := slog.NewJSONHandler(io.MultiWriter(f), opts)
	logger = slog.New(h)
	logFile = f

	return func() {
		loggerMu.Lock()
		defer loggerMu.Unlock()
		if logFile != nil {
			_ = logFile.Sync()
			_ = logFile.Close()
			logFile = nil
		}
	}, nil
}

func SLogger() *slog.Logger {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	if logger == nil {
		return slog.Default()
	}
	return logger
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
