package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mhpsy/emget/internal/downloader"
)

type progressScreen struct {
	app   *App
	tasks map[string]*downloader.Task
	order []string
}

func newProgressScreen(a *App) screen {
	p := &progressScreen{
		app:   a,
		tasks: map[string]*downloader.Task{},
	}
	return p
}

type queueEventMsg downloader.Event
type tickMsg time.Time

func (p *progressScreen) Init() tea.Cmd { return tick() }

func tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (p *progressScreen) Update(msg tea.Msg) (tea.Cmd, screenID) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "esc":
			return nil, screenSearch
		}
	case queueEventMsg:
		p.applyEvent(downloader.Event(m))
	case tickMsg:
		// drain any pending events
		for done := false; !done; {
			select {
			case ev := <-p.app.queue.Events():
				p.applyEvent(ev)
			default:
				done = true
			}
		}
		return tick(), -1
	}
	return nil, -1
}

func (p *progressScreen) applyEvent(ev downloader.Event) {
	if ev.Task == nil {
		return
	}
	id := ev.Task.ID
	if _, ok := p.tasks[id]; !ok {
		p.order = append(p.order, id)
	}
	// copy to avoid aliasing mutations
	tcopy := *ev.Task
	p.tasks[id] = &tcopy
}

func (p *progressScreen) View() string {
	var b strings.Builder
	b.WriteString("Downloads:\n\n")
	if len(p.order) == 0 {
		b.WriteString(infoStyle.Render("(no tasks yet — enqueue from a movie detail)\n"))
	} else {
		for _, id := range p.order {
			t := p.tasks[id]
			b.WriteString(renderTask(t) + "\n")
		}
	}
	b.WriteString("\n" + infoStyle.Render("esc: back"))
	return b.String()
}

func renderTask(t *downloader.Task) string {
	status := string(t.Status)
	bar := ""
	if t.TotalSize > 0 {
		frac := float64(t.Downloaded) / float64(t.TotalSize)
		if frac > 1 {
			frac = 1
		}
		width := 30
		filled := int(frac * float64(width))
		bar = fmt.Sprintf("[%s%s] %5.1f%%", strings.Repeat("#", filled), strings.Repeat(".", width-filled), frac*100)
	} else {
		bar = fmt.Sprintf("%d bytes", t.Downloaded)
	}
	line := fmt.Sprintf("  %-11s %s  %s", status, bar, t.DisplayName)
	switch t.Status {
	case downloader.StatusFailed:
		line = errorStyle.Render(line) + " — " + t.LastError
	case downloader.StatusCompleted:
		line = checkedStyle.Render(line)
	case downloader.StatusDownloading:
		line = selectedStyle.Render(line)
	}
	return line
}
