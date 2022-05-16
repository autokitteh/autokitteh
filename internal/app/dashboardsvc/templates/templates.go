package templates

import "embed"

//go:embed *.html *.css *.tmpl
var FS embed.FS
