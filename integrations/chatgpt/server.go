package chatgpt

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// uiPath is the URL root path of a simple web UI to interact with users.
	uiPath = "/chatgpt/connect/"

	// savePath is the URL path for our handler to save a new autokitteh
	// connection, after the user submits its details via a web form.
	savePath = "/chatgpt/save"
)

func Start(l *zap.Logger, mux *http.ServeMux, s sdkservices.Secrets) {
	// New connection UI + form submission handler.
	mux.Handle(uiPath, http.FileServer(http.FS(static.ChatGPTWebContent)))
	mux.Handle(savePath, NewHTTPHandler(l, s, "chatgpt"))
}
