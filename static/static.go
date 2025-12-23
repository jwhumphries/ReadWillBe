//go:build !dev

package static

import "embed"

//go:embed *
var FS embed.FS
