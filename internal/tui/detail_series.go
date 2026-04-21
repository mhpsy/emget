package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mhpsy/emget/internal/emby"
)

type detailSeriesScreen struct {
	app    *App
	series *emby.Item

	seasons        []emby.Item
	loadingSeasons bool
	loadErr        string

	episodesBySeason map[string][]emby.Item
	loadingEpisodes  map[string]bool
	episodeErr       map[string]string
	expanded         map[string]bool

	cursor int // index into the flattened view
}

type row struct {
	kind     rowKind
	seasonID string
	season   *emby.Item
	episode  *emby.Item
}

type rowKind int

const (
	rowSeason rowKind = iota
	rowEpisode
	rowEpisodeLoading
	rowEpisodeError
)

func newDetailSeriesScreen(a *App) screen {
	return &detailSeriesScreen{
		app:              a,
		episodesBySeason: map[string][]emby.Item{},
		loadingEpisodes:  map[string]bool{},
		episodeErr:       map[string]string{},
		expanded:         map[string]bool{},
	}
}

func (d *detailSeriesScreen) setSeries(it *emby.Item) {
	d.series = it
	d.seasons = nil
	d.loadErr = ""
	d.loadingSeasons = true
	d.episodesBySeason = map[string][]emby.Item{}
	d.loadingEpisodes = map[string]bool{}
	d.episodeErr = map[string]string{}
	d.expanded = map[string]bool{}
	d.cursor = 0
}

type seasonsLoadedMsg struct {
	seasons []emby.Item
	err     error
}

type episodesLoadedMsg struct {
	seasonID string
	episodes []emby.Item
	err      error
}

func (d *detailSeriesScreen) Init() tea.Cmd {
	if d.series == nil || !d.loadingSeasons {
		return nil
	}
	app := d.app
	id := d.series.ID
	return func() tea.Msg {
		seasons, err := app.client.GetSeasons(app.ctx, id)
		return seasonsLoadedMsg{seasons: seasons, err: err}
	}
}

func (d *detailSeriesScreen) Update(msg tea.Msg) (tea.Cmd, screenID) {
	switch m := msg.(type) {
	case seasonsLoadedMsg:
		d.loadingSeasons = false
		if m.err != nil {
			d.loadErr = m.err.Error()
			return nil, -1
		}
		d.seasons = m.seasons
		return nil, -1
	case episodesLoadedMsg:
		d.loadingEpisodes[m.seasonID] = false
		if m.err != nil {
			d.episodeErr[m.seasonID] = m.err.Error()
			return nil, -1
		}
		d.episodesBySeason[m.seasonID] = m.episodes
		return nil, -1
	case tea.KeyMsg:
		return d.handleKey(m)
	}
	return nil, -1
}

func (d *detailSeriesScreen) handleKey(m tea.KeyMsg) (tea.Cmd, screenID) {
	rows := d.rows()
	switch m.String() {
	case "esc":
		return nil, screenResults
	case "up", "k":
		if d.cursor > 0 {
			d.cursor--
		}
	case "down", "j":
		if d.cursor < len(rows)-1 {
			d.cursor++
		}
	case "tab", "enter":
		if d.cursor < len(rows) && rows[d.cursor].kind == rowSeason {
			return d.toggleSeason(rows[d.cursor].seasonID), -1
		}
	}
	return nil, -1
}

func (d *detailSeriesScreen) toggleSeason(seasonID string) tea.Cmd {
	if d.expanded[seasonID] {
		d.expanded[seasonID] = false
		return nil
	}
	d.expanded[seasonID] = true
	if _, cached := d.episodesBySeason[seasonID]; cached {
		return nil
	}
	if d.loadingEpisodes[seasonID] {
		return nil
	}
	d.loadingEpisodes[seasonID] = true
	d.episodeErr[seasonID] = ""
	app := d.app
	seriesID := d.series.ID
	return func() tea.Msg {
		episodes, err := app.client.GetEpisodes(app.ctx, seriesID, seasonID)
		return episodesLoadedMsg{seasonID: seasonID, episodes: episodes, err: err}
	}
}

func (d *detailSeriesScreen) rows() []row {
	rows := make([]row, 0, len(d.seasons)*2)
	for i := range d.seasons {
		s := &d.seasons[i]
		rows = append(rows, row{kind: rowSeason, seasonID: s.ID, season: s})
		if !d.expanded[s.ID] {
			continue
		}
		if d.loadingEpisodes[s.ID] {
			rows = append(rows, row{kind: rowEpisodeLoading, seasonID: s.ID})
			continue
		}
		if errMsg := d.episodeErr[s.ID]; errMsg != "" {
			rows = append(rows, row{kind: rowEpisodeError, seasonID: s.ID})
			continue
		}
		for j := range d.episodesBySeason[s.ID] {
			ep := &d.episodesBySeason[s.ID][j]
			rows = append(rows, row{kind: rowEpisode, seasonID: s.ID, episode: ep})
		}
	}
	return rows
}

func (d *detailSeriesScreen) View() string {
	if d.series == nil {
		return infoStyle.Render("no series loaded")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s", d.series.Name)
	if d.series.ProductionYear > 0 {
		fmt.Fprintf(&b, " (%d)", d.series.ProductionYear)
	}
	b.WriteString("\n\n")

	if d.loadingSeasons {
		b.WriteString(infoStyle.Render("loading seasons...\n"))
		b.WriteString("\n" + infoStyle.Render("esc: back"))
		return b.String()
	}
	if d.loadErr != "" {
		b.WriteString(errorStyle.Render("load failed: " + d.loadErr + "\n"))
		b.WriteString("\n" + infoStyle.Render("esc: back"))
		return b.String()
	}
	rows := d.rows()
	if len(rows) == 0 {
		b.WriteString(infoStyle.Render("(no seasons)\n"))
		b.WriteString("\n" + infoStyle.Render("esc: back"))
		return b.String()
	}
	for i, r := range rows {
		line := d.renderRow(r)
		if i == d.cursor {
			line = selectedStyle.Render("> " + strings.TrimPrefix(line, "  "))
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n" + infoStyle.Render("tab/enter: expand, ↑/↓ move, esc: back"))
	return b.String()
}

func (d *detailSeriesScreen) renderRow(r row) string {
	switch r.kind {
	case rowSeason:
		indicator := "▸"
		if d.expanded[r.seasonID] {
			indicator = "▼"
		}
		return fmt.Sprintf("  %s %s", indicator, r.season.Name)
	case rowEpisode:
		code := fmt.Sprintf("S%02dE%02d", r.episode.ParentIndexNumber, r.episode.IndexNumber)
		return fmt.Sprintf("      %s - %s", code, r.episode.Name)
	case rowEpisodeLoading:
		return infoStyle.Render("      loading episodes...")
	case rowEpisodeError:
		return errorStyle.Render("      load failed: " + d.episodeErr[r.seasonID])
	}
	return ""
}
