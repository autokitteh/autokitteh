package dashboard

import (
	"embed"
	"io/fs"
)

//go:embed build/*
var raw embed.FS

var FS fs.FS

func init() {
	var err error
	if FS, err = fs.Sub(raw, "build"); err != nil {
		panic(err)
	}
}
