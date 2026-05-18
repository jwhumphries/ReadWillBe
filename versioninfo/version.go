// Package versioninfo exposes the build-time version tag for the readwillbe binary.
package versioninfo

// Tag is the version string set at build time via -ldflags. Defaults to "dev".
var Tag = "dev"
