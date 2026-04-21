package downloader

import (
	"context"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// fake store just captures Upsert calls
type fakeStore struct {
	upserts atomic.Int32
	saves   atomic.Int32
}

func (f *fakeStore) Upsert(t *Task)  { f.upserts.Add(1) }
func (f *fakeStore) Save() error     { f.saves.Add(1); return nil }

// fake downloader that just marks task as "done" and returns err or nil
type fakeRunner struct {
	err error
	ran atomic.Int32
}

func (f *fakeRunner) Run(ctx context.Context, t *Task, p ProgressFn) error {
	f.ran.Add(1)
	if f.err != nil {
		return f.err
	}
	if p != nil {
		p(int64(len("done")))
	}
	return nil
}

func TestQueue_ExecutesTasksSerially(t *testing.T) {
	fs := &fakeStore{}
	fr := &fakeRunner{}
	q := NewQueue(QueueConfig{
		Runner: fr,
		Store:  fs,
		Delay:  0, // no throttle for test
		Jitter: 0,
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q.Start(ctx)
	defer q.Stop()

	q.Enqueue(&Task{ID: "t1", OutputPath: filepath.Join(t.TempDir(), "a")})
	q.Enqueue(&Task{ID: "t2", OutputPath: filepath.Join(t.TempDir(), "b")})

	// wait for both to complete
	timeout := time.After(2 * time.Second)
	for fr.ran.Load() < 2 {
		select {
		case <-timeout:
			t.Fatalf("only %d tasks ran", fr.ran.Load())
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestQueue_ThrottleBetweenTasks(t *testing.T) {
	fs := &fakeStore{}
	fr := &fakeRunner{}
	q := NewQueue(QueueConfig{
		Runner: fr,
		Store:  fs,
		Delay:  50 * time.Millisecond,
		Jitter: 0,
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q.Start(ctx)
	defer q.Stop()

	start := time.Now()
	q.Enqueue(&Task{ID: "a"})
	q.Enqueue(&Task{ID: "b"})
	q.Enqueue(&Task{ID: "c"})

	timeout := time.After(2 * time.Second)
	for fr.ran.Load() < 3 {
		select {
		case <-timeout:
			t.Fatalf("only %d tasks ran", fr.ran.Load())
		case <-time.After(10 * time.Millisecond):
		}
	}
	elapsed := time.Since(start)
	// between 3 tasks there are 2 delays; expect at least 2*50ms = 100ms
	if elapsed < 100*time.Millisecond {
		t.Errorf("elapsed = %v, want >= 100ms", elapsed)
	}
}

func TestQueue_FailedTaskMarkedFailed(t *testing.T) {
	fs := &fakeStore{}
	fr := &fakeRunner{err: errors.New("boom")}
	q := NewQueue(QueueConfig{
		Runner: fr,
		Store:  fs,
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make([]Event, 0, 4)
	done := make(chan struct{})
	go func() {
		for ev := range q.Events() {
			events = append(events, ev)
			if ev.Kind == EventFailed {
				close(done)
				return
			}
		}
	}()

	q.Start(ctx)
	defer q.Stop()
	q.Enqueue(&Task{ID: "t1"})

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("no EventFailed received")
	}
}

func TestQueue_ContextCancel_StopsLoop(t *testing.T) {
	fr := &fakeRunner{}
	fs := &fakeStore{}
	q := NewQueue(QueueConfig{Runner: fr, Store: fs})
	ctx, cancel := context.WithCancel(context.Background())
	q.Start(ctx)

	q.Enqueue(&Task{ID: "a"})
	// wait for first task to run
	timeout := time.After(1 * time.Second)
	for fr.ran.Load() < 1 {
		select {
		case <-timeout:
			t.Fatal("first task never ran")
		case <-time.After(10 * time.Millisecond):
		}
	}

	cancel()
	done := make(chan struct{})
	go func() { q.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("queue did not stop within 1s after ctx cancel")
	}
}

type panicRunner struct{}

func (p *panicRunner) Run(ctx context.Context, t *Task, _ ProgressFn) error {
	panic("boom")
}

func TestQueue_RecoversFromPanicInRunner(t *testing.T) {
	fs := &fakeStore{}
	q := NewQueue(QueueConfig{
		Runner: &panicRunner{},
		Store:  fs,
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q.Start(ctx)
	defer q.Stop()

	failedEvent := make(chan Event, 4)
	go func() {
		for ev := range q.Events() {
			if ev.Kind == EventFailed {
				select {
				case failedEvent <- ev:
				default:
				}
			}
		}
	}()

	q.Enqueue(&Task{ID: "p1"})
	q.Enqueue(&Task{ID: "p2"})

	seen := 0
	timeout := time.After(2 * time.Second)
	for seen < 2 {
		select {
		case <-failedEvent:
			seen++
		case <-timeout:
			t.Fatalf("only %d failed events received", seen)
		}
	}
}
