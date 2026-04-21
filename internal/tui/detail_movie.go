package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/mhpsy/emget/internal/downloader"
	"github.com/mhpsy/emget/internal/emby"
)

type dmFocus int

const (
	dmFocusVersions dmFocus = iota
	dmFocusSubtitles
)

type detailMovieScreen struct {
	app        *App
	item       *emby.Item
	focus      dmFocus
	versionIdx int
	subCursor  int
	subChecked map[int]bool // keyed by MediaStream.Index within chosen source
}

func newDetailMovieScreen(a *App) screen {
	return &detailMovieScreen{app: a, subChecked: map[int]bool{}}
}

func (d *detailMovieScreen) setItem(it *emby.Item) {
	d.item = it
	d.focus = dmFocusVersions
	d.versionIdx = 0
	d.subCursor = 0
	d.subChecked = map[int]bool{}
}

func (d *detailMovieScreen) Init() tea.Cmd { return nil }

func (d *detailMovieScreen) currentSource() *emby.MediaSource {
	if d.item == nil || len(d.item.MediaSources) == 0 {
		return nil
	}
	if d.versionIdx < 0 || d.versionIdx >= len(d.item.MediaSources) {
		return nil
	}
	return &d.item.MediaSources[d.versionIdx]
}

func (d *detailMovieScreen) externalSubs() []emby.MediaStream {
	src := d.currentSource()
	if src == nil {
		return nil
	}
	out := make([]emby.MediaStream, 0, len(src.MediaStreams))
	for _, s := range src.MediaStreams {
		if s.Type == "Subtitle" && s.IsExternal {
			out = append(out, s)
		}
	}
	return out
}

func (d *detailMovieScreen) Update(msg tea.Msg) (tea.Cmd, screenID) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil, -1
	}
	switch km.String() {
	case "esc":
		return nil, screenResults
	case "tab":
		if d.focus == dmFocusVersions {
			d.focus = dmFocusSubtitles
			d.subCursor = 0
		} else {
			d.focus = dmFocusVersions
		}
	case "up", "k":
		if d.focus == dmFocusVersions && d.versionIdx > 0 {
			d.versionIdx--
			d.subChecked = map[int]bool{} // reset subs when version changes
			d.subCursor = 0
		}
		if d.focus == dmFocusSubtitles && d.subCursor > 0 {
			d.subCursor--
		}
	case "down", "j":
		if d.focus == dmFocusVersions && d.versionIdx < len(d.item.MediaSources)-1 {
			d.versionIdx++
			d.subChecked = map[int]bool{}
			d.subCursor = 0
		}
		if d.focus == dmFocusSubtitles {
			if d.subCursor < len(d.externalSubs())-1 {
				d.subCursor++
			}
		}
	case " ", "space":
		if d.focus == dmFocusSubtitles {
			subs := d.externalSubs()
			if d.subCursor < len(subs) {
				idx := subs[d.subCursor].Index
				d.subChecked[idx] = !d.subChecked[idx]
			}
		}
	case "d":
		return d.enqueue(), screenProgress
	}
	return nil, -1
}

func (d *detailMovieScreen) enqueue() tea.Cmd {
	src := d.currentSource()
	if src == nil {
		return flash("no media source selected", true)
	}
	ext := src.Container
	if ext == "" {
		ext = "mkv"
	}
	videoPath, _ := downloader.MoviePaths(
		d.app.cfg.OutputDir, d.app.cfg.MoviesSubdir,
		d.item.Name, d.item.ProductionYear, ext, "", "")

	parentID := uuid.NewString()
	videoTask := &downloader.Task{
		ID:          parentID,
		Kind:        downloader.KindVideo,
		ItemID:      d.item.ID,
		SourceID:    src.ID,
		StreamIndex: -1,
		DisplayName: fmt.Sprintf("%s (%d) — %s", d.item.Name, d.item.ProductionYear, src.Name),
		URL:         d.app.client.VideoDownloadURL(d.item.ID, src.ID),
		OutputPath:  videoPath,
		Status:      downloader.StatusQueued,
		CreatedAt:   time.Now().UTC(),
	}
	d.app.queue.Enqueue(videoTask)

	count := 1
	subs := d.externalSubs()
	for _, s := range subs {
		if !d.subChecked[s.Index] {
			continue
		}
		subExt := s.Codec
		if subExt == "" {
			subExt = "srt"
		}
		_, subPath := downloader.MoviePaths(
			d.app.cfg.OutputDir, d.app.cfg.MoviesSubdir,
			d.item.Name, d.item.ProductionYear, ext, subLang(s), subExt)
		subTask := &downloader.Task{
			ID:          uuid.NewString(),
			ParentID:    parentID,
			Kind:        downloader.KindSubtitle,
			ItemID:      d.item.ID,
			SourceID:    src.ID,
			StreamIndex: s.Index,
			DisplayName: fmt.Sprintf("%s — subtitle %s", d.item.Name, subLang(s)),
			URL:         d.app.client.SubtitleDownloadURL(d.item.ID, src.ID, s.Index, subExt),
			OutputPath:  subPath,
			Status:      downloader.StatusQueued,
			CreatedAt:   time.Now().UTC(),
		}
		d.app.queue.Enqueue(subTask)
		count++
	}
	return flash(fmt.Sprintf("enqueued %d task(s)", count), false)
}

func subLang(s emby.MediaStream) string {
	if s.Language != "" {
		return s.Language
	}
	return "und"
}

func (d *detailMovieScreen) View() string {
	if d.item == nil {
		return infoStyle.Render("no item loaded")
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s (%d)\n\n", d.item.Name, d.item.ProductionYear))
	b.WriteString("Versions:\n")
	for i, src := range d.item.MediaSources {
		line := fmt.Sprintf("  [%d] %s  (%.2f GB)", i, src.Name, float64(src.Size)/(1<<30))
		if i == d.versionIdx && d.focus == dmFocusVersions {
			line = selectedStyle.Render("> " + line[2:])
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\nSubtitles (external):\n")
	subs := d.externalSubs()
	if len(subs) == 0 {
		b.WriteString(infoStyle.Render("  (none)\n"))
	} else {
		for i, s := range subs {
			check := "[ ]"
			if d.subChecked[s.Index] {
				check = checkedStyle.Render("[x]")
			}
			title := s.Title
			if title == "" {
				title = s.DisplayTitle
			}
			line := fmt.Sprintf("  %s %s (%s)", check, subLang(s), title)
			if i == d.subCursor && d.focus == dmFocusSubtitles {
				line = selectedStyle.Render("> " + line[2:])
			}
			b.WriteString(line + "\n")
		}
	}
	b.WriteString("\n" + infoStyle.Render("tab: switch focus, ↑/↓ move, space: toggle sub, d: enqueue, esc: back"))
	return b.String()
}
