package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	version = "v0.4.0"
	commit = "abc1234"
	date = "2026-04-22"
	defer func() {
		version = "dev"
		commit = "none"
		date = "unknown"
	}()
	var buf bytes.Buffer
	if err := runVersion(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"v0.4.0", "abc1234", "2026-04-22"} {
		if !strings.Contains(out, want) {
			t.Errorf("runVersion output missing %q: %s", want, out)
		}
	}
}
