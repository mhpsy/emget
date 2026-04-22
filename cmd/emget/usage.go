package main

import (
	"fmt"
	"io"
)

const usageText = `emget — Emby command-line downloader

Usage:
  emget                  Launch the TUI (default)
  emget tui              Launch the TUI (explicit)
  emget tasks [flags]    List queued/active/finished tasks
  emget clean [flags]    Remove tasks from the state store
  emget config [flags]   Print config paths and contents
  emget version          Print version and build info
  emget help | -h        Show this help

Run 'emget <command> -h' for command-specific flags.
`

func printUsage(w io.Writer) {
	fmt.Fprint(w, usageText)
}
