package state

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/mhpsy/emget/internal/downloader"
)

func TestStore_LoadMissingFileReturnsEmpty(t *testing.T) {
	s := NewStore(filepath.Join(t.TempDir(), "state.json"))
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if len(s.Tasks()) != 0 {
		t.Errorf("tasks = %d, want 0", len(s.Tasks()))
	}
}

func TestStore_SaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s := NewStore(path)
	t1 := &downloader.Task{
		ID:         "a",
		Kind:       downloader.KindVideo,
		URL:        "http://x/a",
		OutputPath: "/tmp/a.mkv",
		Status:     downloader.StatusDownloading,
		Downloaded: 1024,
		CreatedAt:  time.Now().UTC().Truncate(time.Second),
	}
	s.Upsert(t1)
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	s2 := NewStore(path)
	if err := s2.Load(); err != nil {
		t.Fatal(err)
	}
	if len(s2.Tasks()) != 1 {
		t.Fatalf("tasks = %d", len(s2.Tasks()))
	}
	got := s2.Tasks()[0]
	if got.ID != "a" || got.Downloaded != 1024 {
		t.Errorf("roundtrip mismatch: %+v", got)
	}
}

func TestStore_ResumeDownloadingToQueuedOnLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s := NewStore(path)
	s.Upsert(&downloader.Task{ID: "x", Status: downloader.StatusDownloading})
	s.Save()

	s2 := NewStore(path)
	s2.Load()
	if s2.Tasks()[0].Status != downloader.StatusQueued {
		t.Errorf("status = %s, want queued", s2.Tasks()[0].Status)
	}
}

func TestStore_Delete(t *testing.T) {
	s := NewStore(filepath.Join(t.TempDir(), "state.json"))
	s.Upsert(&downloader.Task{ID: "a"})
	s.Upsert(&downloader.Task{ID: "b"})
	s.Delete("a")
	if len(s.Tasks()) != 1 || s.Tasks()[0].ID != "b" {
		t.Errorf("tasks = %+v", s.Tasks())
	}
}
