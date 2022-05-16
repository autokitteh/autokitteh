package main

import (
	"github.com/autokitteh/autokitteh/pkg/autokitteh/pluginsvc"

	"github.com/autokitteh/autokitteh/internal/pkg/plugins/githubplugin"
	"github.com/autokitteh/autokitteh/internal/pkg/plugins/googlesheetsplugin"
	"github.com/autokitteh/autokitteh/internal/pkg/plugins/slackplugin"
)

func main() {
	pluginsvc.Run(
		githubplugin.Plugin,
		googlesheetsplugin.Plugin,
		slackplugin.Plugin,
	)
}
