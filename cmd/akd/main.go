package main

import "github.com/autokitteh/autokitteh/internal/app/aksvc"

var version, commit, date string

func main() { aksvc.Run(&aksvc.Version{Version: version, Commit: commit, Date: date}) }
