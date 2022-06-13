package static

import "embed"

//go:embed *.js *.css
var FS embed.FS
