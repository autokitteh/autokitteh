package webplatform

import (
	"embed"
	"io/fs"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

//go:embed dist/*
var distFS embed.FS

func FS() fs.FS { return kittehs.Must1(fs.Sub(distFS, "dist")) }
