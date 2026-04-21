# emget

A Go TUI CLI for downloading movies (and, coming soon, TV series) from an Emby server. Strictly serial downloads with Range-based resume and retry-with-backoff.

## Status

**v0.1.0 MVP — movies only.** Series support lands in v0.2.

## Install / build

```bash
make build   # produces bin/emget
```

Requires Go 1.22+.

## Configure

First run writes a template:

```bash
./bin/emget
# emget: wrote template to ~/.config/emget/config.yaml — fill it in and re-run
```

Edit `~/.config/emget/config.yaml`:

```yaml
emby:
  endpoint: https://your-emby.example.com
  username: your-username
  password: your-password

download:
  output_dir: ~/Media
  # (optional overrides; see template for full list)
```

## Usage

```bash
./bin/emget
```

1. Type a search term, press Enter
2. `↑/↓` to navigate results, Enter to open detail
3. `Tab` to switch between versions and subtitles panes
4. `↑/↓` to move; `space` toggles subtitle selection
5. `d` enqueues video + selected subtitles
6. `p` opens progress/queue panel; `esc` returns to search
7. `q` quits

## Manual E2E verification

Before cutting a release, verify against a real Emby server:

- [ ] Config template writes cleanly when missing
- [ ] Login succeeds with valid credentials; session cached at `~/.cache/emget/session.json`
- [ ] Login fails with clear error on bad password
- [ ] Search for a known movie title returns results
- [ ] Movie detail shows multiple versions (if available) with sizes
- [ ] External subtitles list includes expected languages
- [ ] Enqueue: small file downloads successfully to `<output>/Movies/<Title> (<Year>)/...`
- [ ] Resume: kill CLI mid-download, restart, confirm `.part` continues from correct byte offset
- [ ] Retry: temporarily block network, observe backoff in log; on restore, download completes
- [ ] Ctrl+C twice: clean shutdown, `state.json` reflects final downloaded bytes

## Paths

- Config: `$XDG_CONFIG_HOME/emget/config.yaml` (default `~/.config/emget/config.yaml`)
- Session cache: `$XDG_CACHE_HOME/emget/session.json`
- Task state: `$XDG_DATA_HOME/emget/state.json`
- Logs: path configurable via `logging.file` (default `~/.local/share/emget/emget.log`)
