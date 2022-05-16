package cuemod

import (
	"embed"
)

//go:embed module.cue pkg
var FS embed.FS
