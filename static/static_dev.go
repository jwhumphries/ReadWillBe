//go:build dev

package static

import (
	"io/fs"

	"github.com/spf13/afero"
)

// FS is the on-disk filesystem rooted at ./static (dev builds only).
var FS fs.FS = afero.NewIOFS(afero.NewBasePathFs(afero.NewOsFs(), "static"))
