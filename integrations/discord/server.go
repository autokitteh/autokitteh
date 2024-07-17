package discord

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// savePath is the URL path for our handler to save a new autokitteh
	// connection, after the user submits its details via a web form.
	// savePath = "/discord/save"

	// oauthPath is the URL path for our handler to save new OAuth-based connections.
	oauthPath = "/discord/oauth"
)

func Start(l *zap.Logger, mux *http.ServeMux, vars sdkservices.Vars, o sdkservices.OAuth) {
	// New connection UI + form submission handler.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	mux.Handle(uiPath, http.FileServer(http.FS(static.DiscordWebContent)))

	mux.HandleFunc("GET "+oauthPath, NewHTTPHandler(l, o, vars).handleOAuth)
}
