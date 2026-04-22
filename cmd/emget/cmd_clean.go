package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mhpsy/emget/internal/downloader"
	"github.com/mhpsy/emget/internal/state"
)

type runCleanOpts struct {
	statePath     string
	completedOnly bool
	failedOnly    bool
	yes           bool
	stdout        io.Writer
	stdin         io.Reader
}

func runClean(opts runCleanOpts) error {
	out := opts.stdout
	if out == nil {
		out = os.Stdout
	}
	in := opts.stdin
	if in == nil {
		in = os.Stdin
	}

	s := state.NewStore(opts.statePath)
	if err := s.Load(); err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	pred := func(*downloader.Task) bool { return true }
	scope := "ALL tasks"
	switch {
	case opts.completedOnly:
		pred = func(t *downloader.Task) bool { return t.Status == downloader.StatusCompleted }
		scope = "completed tasks"
	case opts.failedOnly:
		pred = func(t *downloader.Task) bool { return t.Status == downloader.StatusFailed }
		scope = "failed tasks"
	}

	if !opts.yes {
		fmt.Fprintf(out, "Remove %s from %s? [y/N] ", scope, opts.statePath)
		r := bufio.NewReader(in)
		line, _ := r.ReadString('\n')
		answer := strings.ToLower(strings.TrimSpace(line))
		if answer != "y" && answer != "yes" {
			fmt.Fprintln(out, "aborted")
			return nil
		}
	}

	n := s.RemoveWhere(pred)
	if err := s.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}
	fmt.Fprintf(out, "removed %d task(s)\n", n)
	return nil
}
