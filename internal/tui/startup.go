package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mhpsy/emget/internal/downloader"
)

type StartupDecision int

const (
	StartupDecisionPending StartupDecision = iota
	StartupDecisionResume
	StartupDecisionClear
	StartupDecisionSkip
)

type startupScreen struct {
	app      *App
	pending  []*downloader.Task
	decision StartupDecision
}

func newStartupScreen(a *App) screen {
	return &startupScreen{app: a}
}

func (s *startupScreen) SetPending(tasks []*downloader.Task) {
	s.pending = tasks
	s.decision = StartupDecisionPending
}

func (s *startupScreen) Init() tea.Cmd { return nil }

func (s *startupScreen) Update(msg tea.Msg) (tea.Cmd, screenID) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "y", "Y":
			s.decision = StartupDecisionResume
			return s.applyDecision(), screenSearch
		case "n", "N":
			s.decision = StartupDecisionClear
			return s.applyDecision(), screenSearch
		case "esc":
			s.decision = StartupDecisionSkip
			return nil, screenSearch
		}
	}
	return nil, -1
}

func (s *startupScreen) applyDecision() tea.Cmd {
	switch s.decision {
	case StartupDecisionResume:
		count := 0
		for _, t := range s.pending {
			if t.Status == downloader.StatusQueued || t.Status == downloader.StatusFailed {
				t.LastError = ""
				t.Status = downloader.StatusQueued
				s.app.queue.Retry(t)
				count++
			}
		}
		return flash(fmt.Sprintf("resumed %d task(s)", count), false)
	case StartupDecisionClear:
		// Task 9 will wire actual store clearing via an interface.
		return flash("state will be cleared", false)
	}
	return nil
}

func (s *startupScreen) View() string {
	var b strings.Builder
	b.WriteString("Recovery\n\n")
	if len(s.pending) == 0 {
		b.WriteString(infoStyle.Render("no unfinished tasks\n"))
	} else {
		fmt.Fprintf(&b, "Found %d unfinished task(s) from a previous session:\n\n", len(s.pending))
		maxShow := 5
		for i, t := range s.pending {
			if i >= maxShow {
				fmt.Fprintf(&b, "  ... and %d more\n", len(s.pending)-maxShow)
				break
			}
			fmt.Fprintf(&b, "  [%s] %s\n", string(t.Status), t.DisplayName)
		}
		b.WriteString("\n")
	}
	b.WriteString(infoStyle.Render("[Y]es resume  [N]o clear  [esc] skip"))
	return b.String()
}
