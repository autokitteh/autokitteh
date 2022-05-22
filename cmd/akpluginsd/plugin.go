package main

import (
	"github.com/autokitteh/autokitteh/sdk/pluginsvc"

	"github.com/autokitteh/autokitteh/internal/pkg/plugins/githubplugin"
	"github.com/autokitteh/autokitteh/internal/pkg/plugins/googlesheetsplugin"
	"github.com/autokitteh/autokitteh/internal/pkg/plugins/slackplugin"
)

var version, commit, date string

func main() {
	pluginsvc.Run(
		&pluginsvc.Version{Version: version, Commit: commit, Date: date},
		githubplugin.Plugin,
		googlesheetsplugin.Plugin,
		slackplugin.Plugin,
	)
}
