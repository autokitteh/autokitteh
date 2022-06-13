package templates

import "embed"

//go:embed *.html *.tmpl
var FS embed.FS
