package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mhpsy/emget/internal/emby"
)

type detailSeriesScreen struct {
	app            *App
	series         *emby.Item
	seasons        []emby.Item
	loadingSeasons bool
	loadErr        string
	cursor         int
}

func newDetailSeriesScreen(a *App) screen {
	return &detailSeriesScreen{app: a}
}

func (d *detailSeriesScreen) setSeries(it *emby.Item) {
	d.series = it
	d.seasons = nil
	d.loadErr = ""
	d.loadingSeasons = true
	d.cursor = 0
}

type seasonsLoadedMsg struct {
	seasons []emby.Item
	err     error
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
	case tea.KeyMsg:
		switch m.String() {
		case "esc":
			return nil, screenResults
		case "up", "k":
			if d.cursor > 0 {
				d.cursor--
			}
		case "down", "j":
			if d.cursor < len(d.seasons)-1 {
				d.cursor++
			}
		}
	}
	return nil, -1
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
	if len(d.seasons) == 0 {
		b.WriteString(infoStyle.Render("(no seasons)\n"))
		b.WriteString("\n" + infoStyle.Render("esc: back"))
		return b.String()
	}

	for i, s := range d.seasons {
		line := fmt.Sprintf("  %s", s.Name)
		if i == d.cursor {
			line = selectedStyle.Render("> " + line[2:])
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n" + infoStyle.Render("↑/↓ move, esc: back"))
	return b.String()
}
