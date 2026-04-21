package downloader

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	var he *HTTPError
	if errors.As(err, &he) {
		if he.Status >= 400 && he.Status < 500 {
			return false
		}
		return true
	}
	return true // network errors, io errors, etc.
}

type SleepFn func(time.Duration)

// runWithRetries runs op up to maxRetries+1 times (initial attempt + retries).
// Backoff is baseBackoff * 2^attempt + jitter of ±500ms.
func runWithRetries(ctx context.Context, maxRetries int, baseBackoff time.Duration, sleep SleepFn, op func(context.Context) error) error {
	if sleep == nil {
		sleep = time.Sleep
	}
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := op(ctx)
		if err == nil {
			return nil
		}
		lastErr = err
		if !shouldRetry(err) {
			return err
		}
		if attempt == maxRetries {
			break
		}
		delay := baseBackoff << attempt
		jitter := time.Duration(rand.Int63n(int64(500*time.Millisecond))) - 250*time.Millisecond
		finalDelay := delay + jitter
		if finalDelay <= 0 {
			finalDelay = 1 * time.Nanosecond
		}
		sleep(finalDelay)
	}
	return lastErr
}
