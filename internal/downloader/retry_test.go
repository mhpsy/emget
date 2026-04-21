package downloader

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestShouldRetry(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"auth", &HTTPError{Status: 401}, false},
		{"forbidden", &HTTPError{Status: 403}, false},
		{"not_found", &HTTPError{Status: 404}, false},
		{"400", &HTTPError{Status: 400}, false},
		{"500", &HTTPError{Status: 500}, true},
		{"503", &HTTPError{Status: 503}, true},
		{"plain_network", errors.New("connection reset"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldRetry(tc.err); got != tc.want {
				t.Errorf("shouldRetry(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestRunWithRetries_SucceedsAfterRetries(t *testing.T) {
	attempts := 0
	op := func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return &HTTPError{Status: 503}
		}
		return nil
	}
	sleeps := []time.Duration{}
	sleepFn := func(d time.Duration) { sleeps = append(sleeps, d) }

	err := runWithRetries(context.Background(), 3, 1*time.Millisecond, sleepFn, op)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d", attempts)
	}
	if len(sleeps) != 2 {
		t.Errorf("sleeps = %v", sleeps)
	}
}

func TestRunWithRetries_GivesUpAfterMax(t *testing.T) {
	attempts := 0
	op := func(ctx context.Context) error {
		attempts++
		return &HTTPError{Status: 503}
	}
	err := runWithRetries(context.Background(), 2, 1*time.Millisecond, func(time.Duration) {}, op)
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if attempts != 3 { // initial + 2 retries
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestRunWithRetries_StopsImmediatelyOnNonRetryable(t *testing.T) {
	attempts := 0
	op := func(ctx context.Context) error {
		attempts++
		return &HTTPError{Status: 401}
	}
	err := runWithRetries(context.Background(), 5, 1*time.Millisecond, func(time.Duration) {}, op)
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (no retries for 401)", attempts)
	}
}
