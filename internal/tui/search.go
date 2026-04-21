package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mhpsy/emget/internal/emby"
)

type searchScreen struct {
	app   *App
	input textinput.Model
	busy  bool
}

type searchResultMsg struct {
	items []emby.Item
	err   error
}

func newSearchScreen(a *App) screen {
	ti := textinput.New()
	ti.Placeholder = "Search movies, series..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 60
	return &searchScreen{app: a, input: ti}
}

func (s *searchScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (s *searchScreen) Update(msg tea.Msg) (tea.Cmd, screenID) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "enter":
			if !s.busy && s.input.Value() != "" {
				s.busy = true
				return s.submit(s.input.Value()), -1
			}
		case "esc":
			// nothing to escape to on search screen
		}
	case searchResultMsg:
		s.busy = false
		if m.err != nil {
			return flash("search failed: "+m.err.Error(), true), -1
		}
		// FIXME(task-19): restore once resultsScreen exists
		// rs := s.app.screens[screenResults].(*resultsScreen)
		// rs.setResults(m.items, s.input.Value())
		_ = m.items
		return nil, screenResults
	}
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return cmd, -1
}

func (s *searchScreen) View() string {
	v := "Search:\n\n" + s.input.View()
	if s.busy {
		v += "\n\n" + infoStyle.Render("searching...")
	} else {
		v += "\n\n" + infoStyle.Render("enter to search, q to quit")
	}
	return v
}

func (s *searchScreen) submit(term string) tea.Cmd {
	app := s.app
	return func() tea.Msg {
		items, err := app.client.Search(app.ctx, term,
			[]emby.ItemType{emby.TypeMovie, emby.TypeSeries}, 50)
		return searchResultMsg{items: items, err: err}
	}
}
