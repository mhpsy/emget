package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mhpsy/emget/internal/downloader"
)

const stateVersion = 1

type fileFormat struct {
	Version int                `json:"version"`
	Tasks   []*downloader.Task `json:"tasks"`
}

type Store struct {
	path  string
	mu    sync.Mutex
	tasks map[string]*downloader.Task
	order []string
}

func NewStore(path string) *Store {
	return &Store{path: path, tasks: make(map[string]*downloader.Task)}
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var f fileFormat
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	s.tasks = make(map[string]*downloader.Task, len(f.Tasks))
	s.order = s.order[:0]
	for _, t := range f.Tasks {
		if t.Status == downloader.StatusDownloading {
			t.Status = downloader.StatusQueued
		}
		s.tasks[t.ID] = t
		s.order = append(s.order, t.ID)
	}
	return nil
}

func (s *Store) Save() error {
	s.mu.Lock()
	tasks := make([]*downloader.Task, 0, len(s.order))
	for _, id := range s.order {
		tasks = append(tasks, s.tasks[id])
	}
	s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(fileFormat{Version: stateVersion, Tasks: tasks}, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) Upsert(t *downloader.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[t.ID]; !ok {
		s.order = append(s.order, t.ID)
	}
	t.UpdatedAt = time.Now().UTC()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = t.UpdatedAt
	}
	s.tasks[t.ID] = t
}

func (s *Store) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
	for i, x := range s.order {
		if x == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
}

// Clear removes all tasks from the store. Call Save() to persist.
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = make(map[string]*downloader.Task)
	s.order = nil
}

func (s *Store) Tasks() []*downloader.Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*downloader.Task, 0, len(s.order))
	for _, id := range s.order {
		out = append(out, s.tasks[id])
	}
	return out
}
