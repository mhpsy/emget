package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mhpsy/emget/internal/emby"
)

type resultsScreen struct {
	app    *App
	items  []emby.Item
	cursor int
	term   string
}

func newResultsScreen(a *App) screen {
	return &resultsScreen{app: a}
}

func (r *resultsScreen) setResults(items []emby.Item, term string) {
	r.items = items
	r.term = term
	r.cursor = 0
}

func (r *resultsScreen) Init() tea.Cmd { return nil }

type loadMovieDetailMsg struct {
	item *emby.Item
	err  error
}

func (r *resultsScreen) Update(msg tea.Msg) (tea.Cmd, screenID) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "up", "k":
			if r.cursor > 0 {
				r.cursor--
			}
		case "down", "j":
			if r.cursor < len(r.items)-1 {
				r.cursor++
			}
		case "enter":
			if len(r.items) == 0 {
				return nil, -1
			}
			selected := r.items[r.cursor]
			switch selected.Type {
			case emby.TypeMovie:
				return r.loadMovieDetail(selected.ID), -1
			case emby.TypeSeries:
				ds := r.app.screens[screenDetailSeries].(*detailSeriesScreen)
				ds.setSeries(&selected)
				return nil, screenDetailSeries
			default:
				return flash("unsupported item type: "+string(selected.Type), true), -1
			}
		case "esc":
			return nil, screenSearch
		}
	case loadMovieDetailMsg:
		if m.err != nil {
			return flash("load detail failed: "+m.err.Error(), true), -1
		}
		dm := r.app.screens[screenDetailMovie].(*detailMovieScreen)
		dm.setItem(m.item)
		return nil, screenDetailMovie
	}
	return nil, -1
}

func (r *resultsScreen) loadMovieDetail(id string) tea.Cmd {
	app := r.app
	return func() tea.Msg {
		item, err := app.client.GetItem(app.ctx, id)
		return loadMovieDetailMsg{item: item, err: err}
	}
}

func (r *resultsScreen) View() string {
	if len(r.items) == 0 {
		return infoStyle.Render(fmt.Sprintf("no results for %q\n\nesc: back", r.term))
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d results for %q\n\n", len(r.items), r.term))
	for i, it := range r.items {
		marker := "  "
		line := fmt.Sprintf("%s [%s] %s", marker, string(it.Type), it.Name)
		if it.ProductionYear > 0 {
			line += fmt.Sprintf(" (%d)", it.ProductionYear)
		}
		if i == r.cursor {
			line = selectedStyle.Render("> " + line[2:])
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n" + infoStyle.Render("↑/↓ move, enter: details, esc: back"))
	return b.String()
}
