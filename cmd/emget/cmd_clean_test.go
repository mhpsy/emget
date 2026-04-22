package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mhpsy/emget/internal/downloader"
	"github.com/mhpsy/emget/internal/state"
)

func seedStore(t *testing.T) (*state.Store, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	s := state.NewStore(path)
	s.Upsert(&downloader.Task{ID: "a", Status: downloader.StatusCompleted, DisplayName: "a"})
	s.Upsert(&downloader.Task{ID: "b", Status: downloader.StatusFailed, DisplayName: "b"})
	s.Upsert(&downloader.Task{ID: "c", Status: downloader.StatusQueued, DisplayName: "c"})
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	return s, path
}

func TestRunClean_All_Yes(t *testing.T) {
	_, path := seedStore(t)
	var out bytes.Buffer
	err := runClean(runCleanOpts{statePath: path, yes: true, stdout: &out, stdin: strings.NewReader("")})
	if err != nil {
		t.Fatal(err)
	}
	s := state.NewStore(path)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if got := len(s.Tasks()); got != 0 {
		t.Errorf("tasks remaining = %d, want 0", got)
	}
	if !strings.Contains(out.String(), "removed 3") {
		t.Errorf("output missing 'removed 3': %s", out.String())
	}
}

func TestRunClean_CompletedOnly(t *testing.T) {
	_, path := seedStore(t)
	var out bytes.Buffer
	err := runClean(runCleanOpts{statePath: path, completedOnly: true, yes: true, stdout: &out})
	if err != nil {
		t.Fatal(err)
	}
	s := state.NewStore(path)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if got := len(s.Tasks()); got != 2 {
		t.Errorf("tasks remaining = %d, want 2", got)
	}
	for _, tk := range s.Tasks() {
		if tk.ID == "a" {
			t.Errorf("completed task 'a' should be gone")
		}
	}
}

func TestRunClean_PromptNo(t *testing.T) {
	_, path := seedStore(t)
	var out bytes.Buffer
	err := runClean(runCleanOpts{statePath: path, stdout: &out, stdin: strings.NewReader("n\n")})
	if err != nil {
		t.Fatal(err)
	}
	s := state.NewStore(path)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if got := len(s.Tasks()); got != 3 {
		t.Errorf("tasks remaining = %d, want 3 (user said no)", got)
	}
	if !strings.Contains(out.String(), "aborted") {
		t.Errorf("expected 'aborted' in output: %s", out.String())
	}
}

func TestRunClean_MutuallyExclusiveFlags(t *testing.T) {
	// Regression: both flags must error at dispatch; runClean itself
	// silently picks completedOnly first, which is the wrong UX.
	// This test documents that dispatch (not runClean) is the guard.
	// runClean called with both flags set behaves like completedOnly:
	_, path := seedStore(t)
	var out bytes.Buffer
	err := runClean(runCleanOpts{
		statePath:     path,
		completedOnly: true,
		failedOnly:    true,
		yes:           true,
		stdout:        &out,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Both flags set = completedOnly wins = 1 removed
	if !strings.Contains(out.String(), "removed 1") {
		t.Errorf("expected completedOnly behavior when both flags set, got: %s", out.String())
	}
}
