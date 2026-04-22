package main

// These are set via -ldflags at build time. See Makefile and
// .goreleaser.yml.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)
