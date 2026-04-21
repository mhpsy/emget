package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const partSuffix = ".part"

type Downloader struct {
	http          *http.Client
	userAgent     string
	progressBatch int64
}

type ProgressFn func(downloaded int64)

func New(httpClient *http.Client) *Downloader {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Downloader{
		http:          httpClient,
		userAgent:     "emget/0.1.0",
		progressBatch: 1 << 20, // 1 MiB
	}
}

func (d *Downloader) SetUserAgent(ua string) { d.userAgent = ua }

// Run downloads task.URL to task.OutputPath. If a <path>.part file already
// exists, resumes via Range. On success, renames .part to the final path.
func (d *Downloader) Run(ctx context.Context, task *Task, onProgress ProgressFn) error {
	if err := os.MkdirAll(filepath.Dir(task.OutputPath), 0o755); err != nil {
		return err
	}
	partPath := task.OutputPath + partSuffix

	existing := int64(0)
	if info, err := os.Stat(partPath); err == nil {
		existing = info.Size()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", task.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", d.userAgent)
	if existing > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existing))
	}

	resp, err := d.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return &HTTPError{Status: resp.StatusCode, URL: task.URL}
	}
	if existing > 0 && resp.StatusCode == 200 {
		// server ignored Range; truncate and restart
		if err := os.Truncate(partPath, 0); err != nil {
			return err
		}
		existing = 0
	}

	if resp.ContentLength > 0 {
		task.TotalSize = existing + resp.ContentLength
	}

	f, err := os.OpenFile(partPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	downloaded := existing
	buf := make([]byte, 32<<10)
	nextProgressAt := downloaded + d.progressBatch

	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return werr
			}
			downloaded += int64(n)
			task.Downloaded = downloaded
			if onProgress != nil && downloaded >= nextProgressAt {
				onProgress(downloaded)
				nextProgressAt = downloaded + d.progressBatch
			}
		}
		if rerr != nil {
			if errors.Is(rerr, io.EOF) {
				break
			}
			return rerr
		}
	}

	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if onProgress != nil {
		onProgress(downloaded)
	}
	return os.Rename(partPath, task.OutputPath)
}

type HTTPError struct {
	Status int
	URL    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("downloader: %s returned %d", e.URL, e.Status)
}
