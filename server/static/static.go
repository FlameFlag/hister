package static

import "embed"

//go:embed app
//go:embed app/_app
var FS embed.FS
