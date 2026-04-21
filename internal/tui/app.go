package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mhpsy/emget/internal/downloader"
	"github.com/mhpsy/emget/internal/emby"
)

type screenID int

const (
	screenSearch screenID = iota
	screenResults
	screenDetailMovie
	screenProgress
)

type screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Cmd, screenID) // returns next-screen id; -1 means stay
	View() string
}

type App struct {
	ctx    context.Context
	client *emby.Client
	queue  *downloader.Queue
	cfg    AppConfig

	current screenID
	screens map[screenID]screen

	width, height int
	flash         string
	flashErr      bool
}

type AppConfig struct {
	OutputDir       string
	MoviesSubdir    string
	TVSubdir        string
	Languages       []string // preferred subtitle languages (external only)
	ResolutionOrder []int    // preferred video heights, most-preferred first
	KeywordBoost    []string // source-name keywords used as tie-breakers
}

func NewApp(ctx context.Context, client *emby.Client, queue *downloader.Queue, cfg AppConfig) *App {
	a := &App{
		ctx:     ctx,
		client:  client,
		queue:   queue,
		cfg:     cfg,
		current: screenSearch,
		screens: map[screenID]screen{},
	}
	a.screens[screenSearch] = newSearchScreen(a)
	a.screens[screenResults] = newResultsScreen(a)
	a.screens[screenDetailMovie] = newDetailMovieScreen(a)
	a.screens[screenProgress] = newProgressScreen(a)
	return a
}

// bubbletea Model interface

func (a *App) Init() tea.Cmd {
	return a.screens[a.current].Init()
}

type switchScreenMsg struct{ to screenID }

type flashMsg struct {
	text  string
	isErr bool
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = m.Width
		a.height = m.Height
	case tea.KeyMsg:
		// ctrl+c is always quit; "q" and other keys are delegated to the active screen
		if m.String() == "ctrl+c" {
			return a, tea.Quit
		}
		if m.String() == "p" {
			a.current = screenProgress
			return a, a.screens[a.current].Init()
		}
	case switchScreenMsg:
		a.current = m.to
		return a, a.screens[a.current].Init()
	case flashMsg:
		a.flash = m.text
		a.flashErr = m.isErr
		return a, nil
	}
	cmd, next := a.screens[a.current].Update(msg)
	if next != -1 {
		a.current = next
		return a, a.screens[a.current].Init()
	}
	return a, cmd
}

func (a *App) View() string {
	top := titleStyle.Render("emget")
	body := a.screens[a.current].View()
	bar := ""
	if a.flash != "" {
		style := infoStyle
		if a.flashErr {
			style = errorStyle
		}
		bar = "\n" + style.Render(a.flash)
	}
	return top + "\n" + body + bar
}

func flash(msg string, isErr bool) tea.Cmd {
	return func() tea.Msg { return flashMsg{text: msg, isErr: isErr} }
}
