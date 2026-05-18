//go:build !dev

// Package static provides the compiled-asset filesystem served at /static.
// In production builds, assets are embedded; in dev builds they are read
// from disk via afero (see static_dev.go).
package static

import "embed"

// FS is the embedded filesystem containing the compiled static assets.
//
//go:embed *
var FS embed.FS
