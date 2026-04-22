package main

import (
	"fmt"
	"io"
	"runtime"
)

func runVersion(w io.Writer) error {
	fmt.Fprintf(w, "emget %s (commit %s, built %s)\n", version, commit, date)
	fmt.Fprintf(w, "%s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return nil
}
