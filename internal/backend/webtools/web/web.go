package web

import (
	"embed"
	"io/fs"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

//go:embed webmessages
var messages embed.FS

var Messages fs.FS

//go:embed webterminal
var terminal embed.FS

var Terminal fs.FS

func init() {
	Messages = kittehs.Must1(fs.Sub(messages, "webmessages"))
	Terminal = kittehs.Must1(fs.Sub(terminal, "webterminal"))
}
