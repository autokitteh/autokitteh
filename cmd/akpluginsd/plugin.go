package main

import (
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/pluginsvc"

	"gitlab.com/softkitteh/autokitteh/internal/pkg/plugins/githubplugin"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/plugins/googlesheetsplugin"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/plugins/slackplugin"
)

func main() {
	pluginsvc.Run(
		githubplugin.Plugin,
		googlesheetsplugin.Plugin,
		slackplugin.Plugin,
	)
}
