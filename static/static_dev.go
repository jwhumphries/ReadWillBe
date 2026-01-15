//go:build dev

package static

import (
	"io/fs"

	"github.com/spf13/afero"
)

var FS fs.FS = afero.NewIOFS(afero.NewBasePathFs(afero.NewOsFs(), "static"))
