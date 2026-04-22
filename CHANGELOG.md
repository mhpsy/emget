# Changelog

All notable changes to emget are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] — 2026-04-22

### Added
- CLI subcommand dispatcher with four new commands: `emget tasks`, `emget clean`, `emget config`, `emget version`.
- `emget tasks` — lists tasks grouped by status (queued / failed / completed). Supports `--status=<list>` filter and `--format=json` for scripting.
- `emget clean` — removes tasks from the state store. Supports `--completed-only`, `--failed-only`, `--yes`. Mutually-exclusive filter flags return an error.
- `emget config` — prints config / cache / data paths and the parsed YAML with the Emby password redacted to `****`. `--paths-only` skips content; `--raw` dumps the file verbatim with a warning banner.
- `emget version` — prints ldflag-injected version, short commit, build date, Go version, and OS/arch.
- Windows and macOS support for config / cache / data directories, using `os.UserConfigDir` + `os.UserCacheDir` + a platform-specific `DataDir`.
- TUI list scrolling on the results, progress, and series-detail screens. New keys: `PgUp` / `PgDn` (half-page), `Home` / `g` (first), `End` / `G` (last). Indicators `↑ N more` / `↓ N more` shown when content overflows.
- GitHub Actions CI workflow running `go vet` / `go test` / `go build` on Ubuntu, macOS, and Windows.
- GitHub Actions release workflow driven by goreleaser: on `v*` tag push, produces archives for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, plus `checksums.txt`.
- Bilingual README (`README.md` + `README.zh-CN.md`).

### Changed
- `Makefile` `build` target now injects `main.version` / `main.commit` / `main.date` via `-ldflags`, sourced from `git describe` / `git rev-parse` / `date`.
- `internal/config/config.go` uses `filepath.Join` when computing the default log path, for Windows compatibility.
- Config file password is redacted at the struct level (not via string substitution), preventing leaks across all YAML formats.
- Root `README.md` rewritten to document all four subcommands, cross-platform paths, and install from release.

### Fixed
- `internal/downloader/naming_test.go` — normalized path separators via `filepath.ToSlash` at assertion boundary; previously the tests hardcoded forward slashes and failed on Windows.
- `internal/config/session_test.go` — gated POSIX `0o600` permission assertion behind `runtime.GOOS != "windows"`; Windows ignores Unix mode bits on file creation.

## [0.3.0] — 2026-04-21

### Added
- Startup recovery screen. If state contains unfinished tasks at launch, user is prompted with `[Y]` resume (re-enqueue), `[N]` clear (wipe state), `[esc]` skip (leave state untouched).
- Log rotation when `emget.log` exceeds 10 MiB; previous generation kept as `emget.log.1`.
- Retry + cancel bindings on the Progress screen: `[r]` retries a failed task, `[x]` cancels a queued task.
- Goroutine panic recovery inside `Queue.runOne`: panics in the downloader runner are captured, the task is marked failed, and the queue continues with the next task.
- Session expiry check on load: expired cached sessions trigger re-authentication with the config credentials, reusing the cached `DeviceID`.
- `state.Store.Clear()` method; used by the startup recovery screen when user declines to resume.

### Changed
- `state.Store.Load()` rewrites tasks with status `downloading` to `queued` on startup, so interrupted downloads resume as queued items.

## [0.2.0] — 2026-04-21

### Added
- TV series support: `Series` result type is now routed to a dedicated Detail Series screen.
- Expandable seasons with lazy-loaded episodes. `Tab` / `Enter` on a season toggles expand/collapse.
- Multi-select for episodes: `space` toggles an episode; `space` on a season toggles all its loaded episodes (tri-state indicators `[x]` / `[~]` / `[ ]`).
- Matcher-driven enqueue. `d` key picks one version and zero-or-more external subtitles per selected episode based on `config.yaml` rules; episodes with no matching version are skipped (not failed).
- `matcher.PickVersion` — scores by `resolution_score × 1000 + keyword_score`; returns `ErrNoMatch` when nothing fits.
- `matcher.PickSubtitles` — stable sort by preferred language order.
- `emby.GetSeasons` and `emby.GetEpisodes` endpoints.
- `downloader.TVPaths` — builds `<output>/TV/<Series>/Season NN/<Series> - SXXEXX - <Title>.<ext>` paths with filename sanitization.
- Extended `AppConfig` with TV subdir, language preferences, resolution order, and keyword-boost fields.

### Changed
- Results screen now branches on item type: movies open Detail Movie, series open Detail Series.

## [0.1.0] — 2026-04-21

Initial MVP release (movies only).

### Added
- Bubbletea-based TUI with five screens: Search, Results, Detail Movie, Progress, plus a transient splash.
- YAML config (`~/.config/emget/config.yaml`) loaded via XDG paths; template written on first run.
- Session persistence with atomic write and `0o600` perms.
- slog-based file logger with configurable level.
- Emby REST client (hand-rolled): `/Users/AuthenticateByName`, `/Users/{id}/Items` (search), `/Items/{id}`, `/Items/{id}/Download`, `/Videos/{id}/{src}/Subtitles/{idx}/Stream.{ext}`.
- Serial download queue (single goroutine, no parallel workers) with inter-download throttle + jitter.
- HTTP Range-based resume using `.part` files and atomic `os.Rename` on completion.
- Smart retry classifier: 4xx not retried, 5xx/network/timeout retried up to N times with exponential backoff + jitter.
- `downloader.MoviePaths` — builds `<output>/Movies/<Title> (<Year>)/<Title> (<Year>).<ext>` with filename sanitization.
- Atomic JSON state store (`~/.local/share/emget/state.json`) with mutex and `.tmp` + rename write.
- Movie detail screen: single-select version + multi-select external subtitles; `d` enqueues video plus selected subtitles.
- Progress screen polling queue events via tick-based channel drain.

[Unreleased]: https://github.com/mhpsy/emget/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/mhpsy/emget/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/mhpsy/emget/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/mhpsy/emget/releases/tag/v0.2.0
[0.1.0]: https://github.com/mhpsy/emget/releases/tag/v0.2.0
