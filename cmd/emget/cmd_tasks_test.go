package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mhpsy/emget/internal/downloader"
	"github.com/mhpsy/emget/internal/state"
)

func TestRunTasks_Table(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	s := state.NewStore(path)
	s.Upsert(&downloader.Task{ID: "aa11bb22cc33", Status: downloader.StatusCompleted, DisplayName: "Movie A", Kind: downloader.KindVideo, TotalSize: 1024, Downloaded: 1024})
	s.Upsert(&downloader.Task{ID: "dd44ee55ff66", Status: downloader.StatusQueued, DisplayName: "Movie B", Kind: downloader.KindVideo})
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err := runTasks(runTasksOpts{statePath: path, format: "table", stdout: &buf})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"queued", "completed", "Movie A", "Movie B", "aa11bb22"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRunTasks_StatusFilter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	s := state.NewStore(path)
	s.Upsert(&downloader.Task{ID: "a", Status: downloader.StatusCompleted, DisplayName: "done"})
	s.Upsert(&downloader.Task{ID: "b", Status: downloader.StatusQueued, DisplayName: "pending"})
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err := runTasks(runTasksOpts{statePath: path, statuses: "queued", format: "table", stdout: &buf})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "pending") {
		t.Errorf("expected 'pending' in output: %s", out)
	}
	if strings.Contains(out, "done") {
		t.Errorf("did not expect 'done' in output: %s", out)
	}
}

func TestRunTasks_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	s := state.NewStore(path)
	s.Upsert(&downloader.Task{ID: "x", Status: downloader.StatusCompleted, DisplayName: "X"})
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err := runTasks(runTasksOpts{statePath: path, format: "json", stdout: &buf})
	if err != nil {
		t.Fatal(err)
	}
	var got []downloader.Task
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("bad json output: %v\n%s", err, buf.String())
	}
	if len(got) != 1 || got[0].DisplayName != "X" {
		t.Errorf("unexpected json: %+v", got)
	}
}
