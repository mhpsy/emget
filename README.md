# emget

A Go TUI CLI for downloading movies (and, coming soon, TV series) from an Emby server. Strictly serial downloads with Range-based resume and retry-with-backoff.

## Status

**v0.2.0 — Movies and TV series.** Log rotation, startup recovery, and automatic re-auth land in v0.3.

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

**Movie detail:**
3. `Tab` to switch between versions and subtitles panes
4. `↑/↓` to move; `space` toggles subtitle selection
5. `d` enqueues video + selected subtitles

**Series detail:**
6. `↑/↓` to navigate seasons/episodes
7. `Tab` or `Enter` on a season to expand its episodes (lazy-loaded)
8. `Space` on an episode toggles selection; `Space` on a season toggles all its loaded episodes
9. `d` enqueues every selected episode using `versions` and `subtitles` rules from your config; episodes with no matching version are reported as skipped
10. `p` opens progress/queue panel; `esc` returns to search
11. `ctrl+c` quits

## Series matching rules

When you press `d` on the Detail Series screen, emget picks one version and zero-or-more subtitle streams per selected episode automatically, using rules from your `config.yaml`:

```yaml
subtitles:
  preferred_languages: [zho, chi, eng]   # external subs matching these languages, in this order

versions:
  resolution_order: [2160, 1080, 720, 480]   # preferred heights, best first
  keyword_boost: [BluRay, REMUX, WEB-DL]     # tie-breakers (case-insensitive substring match on source name)
```

Scoring per MediaSource is `resolution_score × 1000 + keyword_score`. If an episode has no MediaSource whose resolution appears in `resolution_order` (and which has zero keyword hits), that episode is skipped — not failed. The final flash message reports the skip count.

Subtitle streams are filtered to external subs only; non-matching languages are ignored.

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
- [ ] Series search returns Series-type results alongside movies
- [ ] Opening a series loads season list; expanding a season loads episodes lazily
- [ ] `Space` on an episode toggles an `[x]`; `Space` on a season with all episodes checked clears them all
- [ ] Pressing `d` enqueues one video + N subtitle tasks per selected episode, with correct TV naming under `<output>/TV/<Series>/Season NN/...`
- [ ] An episode with no version matching `resolution_order` is counted in the "skipped" tally rather than failing
- [ ] TV subtitle files land next to the video with `.lang.ext` suffix, matching Emby/Plex/Jellyfin conventions

## Paths

- Config: `$XDG_CONFIG_HOME/emget/config.yaml` (default `~/.config/emget/config.yaml`)
- Session cache: `$XDG_CACHE_HOME/emget/session.json`
- Task state: `$XDG_DATA_HOME/emget/state.json`
- Logs: path configurable via `logging.file` (default `~/.local/share/emget/emget.log`)
