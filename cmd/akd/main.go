package main

import (
	"github.com/autokitteh/autokitteh/internal/app/aksvc"
	"strings"
)

var version, commit, date string

func init() {
	version = strings.TrimPrefix(version, "v")
}

func main() { aksvc.Run(&aksvc.Version{Version: version, Commit: commit, Date: date}) }
