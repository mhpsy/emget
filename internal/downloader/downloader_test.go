package downloader

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownload_FullWriteAndRename(t *testing.T) {
	body := []byte("hello world, movie content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "" {
			t.Errorf("unexpected Range on fresh download: %s", r.Header.Get("Range"))
		}
		w.Header().Set("Content-Length", "26")
		w.Write(body)
	}))
	defer srv.Close()

	dir := t.TempDir()
	out := filepath.Join(dir, "movie.mkv")
	d := New(srv.Client())
	var progress int64
	task := &Task{URL: srv.URL, OutputPath: out}
	err := d.Run(context.Background(), task, func(n int64) { progress = n })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(out + ".part"); !os.IsNotExist(err) {
		t.Errorf(".part should be removed after rename")
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(body) {
		t.Errorf("content mismatch: %q vs %q", data, body)
	}
	if progress == 0 {
		t.Error("progress callback never fired")
	}
}

func TestDownload_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		io.WriteString(w, "err")
	}))
	defer srv.Close()

	dir := t.TempDir()
	d := New(srv.Client())
	task := &Task{URL: srv.URL, OutputPath: filepath.Join(dir, "movie.mkv")}
	err := d.Run(context.Background(), task, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
