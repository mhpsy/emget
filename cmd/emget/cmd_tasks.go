package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/mhpsy/emget/internal/downloader"
	"github.com/mhpsy/emget/internal/state"
)

type runTasksOpts struct {
	statePath string
	statuses  string // comma-separated; empty = all
	format    string // "table" or "json"
	stdout    io.Writer
}

// statusGroupOrder controls the order status groups are printed.
// StatusDownloading is not listed because state.Store.Load rewrites
// downloading → queued on startup, so it is never observed here.
var statusGroupOrder = []downloader.Status{
	downloader.StatusQueued,
	downloader.StatusFailed,
	downloader.StatusCompleted,
}

func runTasks(opts runTasksOpts) error {
	out := opts.stdout
	if out == nil {
		out = os.Stdout
	}

	s := state.NewStore(opts.statePath)
	if err := s.Load(); err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	tasks := s.Tasks()

	var allowed map[downloader.Status]bool
	if opts.statuses != "" {
		allowed = map[downloader.Status]bool{}
		for _, raw := range strings.Split(opts.statuses, ",") {
			allowed[downloader.Status(strings.TrimSpace(raw))] = true
		}
	}
	filtered := tasks[:0:0]
	for _, t := range tasks {
		if allowed != nil && !allowed[t.Status] {
			continue
		}
		filtered = append(filtered, t)
	}

	switch opts.format {
	case "json":
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(filtered)
	case "", "table":
		return printTasksTable(out, filtered)
	default:
		return fmt.Errorf("unknown format: %s", opts.format)
	}
}

func printTasksTable(w io.Writer, tasks []*downloader.Task) error {
	if len(tasks) == 0 {
		fmt.Fprintln(w, "(no tasks)")
		return nil
	}
	groups := map[downloader.Status][]*downloader.Task{}
	for _, t := range tasks {
		groups[t.Status] = append(groups[t.Status], t)
	}

	first := true
	for _, st := range statusGroupOrder {
		g := groups[st]
		if len(g) == 0 {
			continue
		}
		if !first {
			fmt.Fprintln(w)
		}
		first = false
		fmt.Fprintf(w, "== %s (%d) ==\n", st, len(g))
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "ID\tTYPE\tNAME\tPROGRESS\tSIZE")
		for _, t := range g {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
				shortID(t.ID), t.Kind, t.DisplayName, formatProgress(t), formatSize(t.TotalSize))
		}
		tw.Flush()
	}
	return nil
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

func formatProgress(t *downloader.Task) string {
	if t.TotalSize <= 0 {
		if t.Downloaded > 0 {
			return fmt.Sprintf("%d B", t.Downloaded)
		}
		return "-"
	}
	pct := float64(t.Downloaded) / float64(t.TotalSize) * 100
	if pct > 100 {
		pct = 100
	}
	return fmt.Sprintf("%.1f%%", pct)
}

func formatSize(n int64) string {
	if n <= 0 {
		return "-"
	}
	const (
		KiB = 1024
		MiB = 1024 * KiB
		GiB = 1024 * MiB
	)
	switch {
	case n >= GiB:
		return fmt.Sprintf("%.2f GiB", float64(n)/GiB)
	case n >= MiB:
		return fmt.Sprintf("%.1f MiB", float64(n)/MiB)
	case n >= KiB:
		return fmt.Sprintf("%.1f KiB", float64(n)/KiB)
	default:
		return fmt.Sprintf("%d B", n)
	}
}
