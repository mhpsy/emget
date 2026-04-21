package downloader

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Runner interface {
	Run(ctx context.Context, t *Task, p ProgressFn) error
}

type Persister interface {
	Upsert(t *Task)
	Save() error
}

type QueueConfig struct {
	Runner     Runner
	Store      Persister
	Delay      time.Duration
	Jitter     time.Duration
	MaxRetries int
	RetryBase  time.Duration
}

type EventKind string

const (
	EventStarted   EventKind = "started"
	EventProgress  EventKind = "progress"
	EventCompleted EventKind = "completed"
	EventFailed    EventKind = "failed"
)

type Event struct {
	Kind       EventKind
	Task       *Task
	Downloaded int64
	Err        error
}

type Queue struct {
	cfg      QueueConfig
	tasks    chan *Task
	events   chan Event
	stopCh   chan struct{}
	wg       sync.WaitGroup
	mu       sync.Mutex
	canceled map[string]bool
}

func NewQueue(cfg QueueConfig) *Queue {
	return &Queue{
		cfg:      cfg,
		tasks:    make(chan *Task, 256),
		events:   make(chan Event, 256),
		stopCh:   make(chan struct{}),
		canceled: map[string]bool{},
	}
}

func (q *Queue) Enqueue(t *Task) {
	if t.Status == "" {
		t.Status = StatusQueued
	}
	if q.cfg.Store != nil {
		q.cfg.Store.Upsert(t)
		_ = q.cfg.Store.Save()
	}
	q.tasks <- t
}

func (q *Queue) Events() <-chan Event { return q.events }

func (q *Queue) Start(ctx context.Context) {
	q.wg.Add(1)
	go q.loop(ctx)
}

func (q *Queue) Stop() {
	close(q.stopCh)
	q.wg.Wait()
	close(q.events)
}

func (q *Queue) loop(ctx context.Context) {
	defer q.wg.Done()
	first := true
	for {
		select {
		case <-ctx.Done():
			return
		case <-q.stopCh:
			return
		case t := <-q.tasks:
			if !first {
				q.sleepThrottle(ctx)
				if ctx.Err() != nil {
					return
				}
			}
			first = false
			q.runOne(ctx, t)
		}
	}
}

func (q *Queue) sleepThrottle(ctx context.Context) {
	if q.cfg.Delay == 0 && q.cfg.Jitter == 0 {
		return
	}
	d := q.cfg.Delay
	if q.cfg.Jitter > 0 {
		d += time.Duration(rand.Int63n(int64(q.cfg.Jitter)))
	}
	select {
	case <-ctx.Done():
	case <-q.stopCh:
	case <-time.After(d):
	}
}

func (q *Queue) runOne(ctx context.Context, t *Task) {
	if q.isCanceled(t.ID) {
		t.Status = StatusFailed
		t.LastError = "canceled by user"
		q.upsert(t)
		q.emit(Event{Kind: EventFailed, Task: t, Err: fmt.Errorf("canceled")})
		return
	}
	t.Status = StatusDownloading
	t.Attempts++
	q.upsert(t)
	q.emit(Event{Kind: EventStarted, Task: t})

	lastSave := time.Now()
	onProgress := func(n int64) {
		t.Downloaded = n
		q.emit(Event{Kind: EventProgress, Task: t, Downloaded: n})
		if time.Since(lastSave) >= 2*time.Second {
			q.upsert(t)
			lastSave = time.Now()
		}
	}

	var err error
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				err = fmt.Errorf("panic in downloader: %v", rec)
			}
		}()
		err = runWithRetries(ctx, q.cfg.MaxRetries, q.cfg.RetryBase, nil, func(ctx context.Context) error {
			return q.cfg.Runner.Run(ctx, t, onProgress)
		})
	}()

	if err != nil {
		t.Status = StatusFailed
		t.LastError = err.Error()
		q.upsert(t)
		q.emit(Event{Kind: EventFailed, Task: t, Err: err})
		return
	}
	t.Status = StatusCompleted
	q.upsert(t)
	q.emit(Event{Kind: EventCompleted, Task: t})
}

func (q *Queue) upsert(t *Task) {
	if q.cfg.Store == nil {
		return
	}
	q.cfg.Store.Upsert(t)
	_ = q.cfg.Store.Save()
}

func (q *Queue) emit(e Event) {
	select {
	case q.events <- e:
	default:
		// drop if nobody listening; Events channel has 256 buffer
	}
}

// Retry re-enqueues a previously-failed task for another attempt.
// Resets its status and progress counters.
func (q *Queue) Retry(t *Task) {
	t.Status = StatusQueued
	t.LastError = ""
	t.Downloaded = 0
	if q.cfg.Store != nil {
		q.cfg.Store.Upsert(t)
		_ = q.cfg.Store.Save()
	}
	q.tasks <- t
}

// Cancel marks a queued task to be skipped when the runner reaches it.
// Has no effect on the currently-running task (use context cancel for that).
func (q *Queue) Cancel(taskID string) {
	q.mu.Lock()
	q.canceled[taskID] = true
	q.mu.Unlock()
}

func (q *Queue) isCanceled(taskID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.canceled[taskID]
}
