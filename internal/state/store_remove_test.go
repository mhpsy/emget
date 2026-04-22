package state

import (
	"path/filepath"
	"testing"

	"github.com/mhpsy/emget/internal/downloader"
)

func TestRemoveWhere(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(filepath.Join(dir, "state.json"))
	s.Upsert(&downloader.Task{ID: "a", Status: downloader.StatusCompleted, DisplayName: "a"})
	s.Upsert(&downloader.Task{ID: "b", Status: downloader.StatusFailed, DisplayName: "b"})
	s.Upsert(&downloader.Task{ID: "c", Status: downloader.StatusQueued, DisplayName: "c"})

	removed := s.RemoveWhere(func(t *downloader.Task) bool {
		return t.Status == downloader.StatusCompleted
	})
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}
	if got := len(s.Tasks()); got != 2 {
		t.Errorf("len(Tasks()) = %d, want 2", got)
	}
	for _, tk := range s.Tasks() {
		if tk.ID == "a" {
			t.Errorf("'a' should have been removed")
		}
	}
}

func TestRemoveWhere_MatchAll(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(filepath.Join(dir, "state.json"))
	s.Upsert(&downloader.Task{ID: "a", DisplayName: "a"})
	s.Upsert(&downloader.Task{ID: "b", DisplayName: "b"})
	removed := s.RemoveWhere(func(*downloader.Task) bool { return true })
	if removed != 2 {
		t.Errorf("removed = %d, want 2", removed)
	}
	if got := len(s.Tasks()); got != 0 {
		t.Errorf("len = %d, want 0", got)
	}
}
