# emget

> TUI + CLI Emby media downloader. Single-connection, resumable, versioned.
>
> Languages: **English** · [中文](README.zh-CN.md)

[![CI](https://github.com/mhpsy/emget/actions/workflows/ci.yml/badge.svg)](https://github.com/mhpsy/emget/actions/workflows/ci.yml)

A Go CLI that downloads movies and TV series from an Emby server. Strictly serial downloads with Range-based resume and retry-with-backoff. Ships with a bubbletea TUI and four standalone subcommands for scripting.

## Install

### From source

```sh
go install github.com/mhpsy/emget/cmd/emget@latest
```

Or from a clone:

```sh
make build   # produces bin/emget with version ldflags
```

Requires Go 1.24+.

### From a release

Download the archive for your platform from the [releases page](https://github.com/mhpsy/emget/releases) and extract `emget` (or `emget.exe`) into a directory on your `PATH`.

Available platforms: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64.

## First-time setup

On first run, emget writes a config template and exits:

```sh
emget
# emget: wrote template to <path> — fill it in and re-run
```

Edit the template (point `emby.*` at your server), then re-run.

### Config paths per OS

| OS      | Config                                        | Cache                        | Data                                      |
|---------|-----------------------------------------------|------------------------------|-------------------------------------------|
| Linux   | `$XDG_CONFIG_HOME/emget` or `~/.config/emget` | `~/.cache/emget`             | `~/.local/share/emget`                    |
| macOS   | `~/Library/Application Support/emget`         | `~/Library/Caches/emget`     | `~/Library/Application Support/emget`     |
| Windows | `%AppData%\emget`                             | `%LocalAppData%\emget\cache` | `%AppData%\emget`                         |

Run `emget config --paths-only` to print the exact paths on your machine.

## Usage

```
emget                  Launch the TUI (default)
emget tui              Launch the TUI (explicit)
emget tasks [flags]    List queued/active/finished tasks
emget clean [flags]    Remove tasks from the state store
emget config [flags]   Print config paths and contents
emget version          Print version and build info
emget help             Show help
```

### TUI keys

- `↑/↓` or `k/j` — move cursor
- `PgUp / PgDn` — half-page
- `Home / g`, `End / G` — jump to first / last
- `Enter` — open detail / expand season
- `Space` — toggle selection (multi-select screens)
- `Tab` — expand / switch pane
- `d` — enqueue selected
- `p` — open Progress screen (global)
- `esc` — back
- `Ctrl+C` — quit

Progress screen adds:
- `r` — retry a failed task
- `x` — cancel a queued task

### `emget tasks`

```sh
emget tasks                        # all tasks, grouped by status
emget tasks --status=failed        # only failed
emget tasks --status=queued,completed
emget tasks --format=json          # raw JSON array
```

### `emget clean`

```sh
emget clean                # prompt, then remove all tasks
emget clean --yes          # no prompt
emget clean --completed-only
emget clean --failed-only
```

Mutually-exclusive flags return an error.

### `emget config`

```sh
emget config               # print paths + parsed YAML (password redacted)
emget config --paths-only  # only print paths
emget config --raw         # print file as-is (password NOT redacted; warning banner)
```

## Series matching rules

When you press `d` on the Detail Series screen, emget picks one version and zero-or-more subtitle streams per selected episode automatically, using rules from `config.yaml`:

```yaml
subtitles:
  preferred_languages: [zho, chi, eng]   # external subs matching these languages, in this order

versions:
  resolution_order: [2160, 1080, 720, 480]   # preferred heights, best first
  keyword_boost: [BluRay, REMUX, WEB-DL]     # tie-breakers (case-insensitive)
```

Scoring per MediaSource is `resolution_score × 1000 + keyword_score`. Episodes with no matching version are skipped (not failed); the flash message reports the skip count.

## Startup recovery

If emget exits with unfinished downloads, the next launch shows a recovery screen:

- `[Y]` resume — re-enqueue unfinished tasks (continues from `.part` files via Range)
- `[N]` clear — wipe the state file and start fresh
- `[esc]` skip — keep state as-is, start at the search screen

The log file rotates once it exceeds 10 MiB (keeps one previous generation as `<log>.1`).

## Features by version

- **v0.1** — movies: search, version pick, external subtitle multi-select, serial downloads with resume + retry
- **v0.2** — series: expandable seasons, episode multi-select, matcher-driven enqueue
- **v0.3** — hardening: startup recovery, log rotation, retry/cancel bindings, panic recovery
- **v0.4** — CLI subcommands (`tasks` / `clean` / `config` / `version`), Windows + macOS support, list scrolling, GitHub-built multi-platform releases

## Developing

```sh
make build    # compile bin/emget with version ldflags
make test     # run unit tests
make vet      # go vet ./...
```

## License

MIT — `LICENSE` file pending.
