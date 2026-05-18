// Package buildinfo exposes the build-time version tag for the readwillbe binary.
package buildinfo

// Tag is the version string set at build time via -ldflags. Defaults to "dev".
var Tag = "dev"
