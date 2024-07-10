package gemini

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// savePath is the URL path for our handler to save a new autokitteh
	// connection, after the user submits its details via a web form.
	savePath = "/googlegemini/save"
)

func Start(l *zap.Logger, mux *http.ServeMux) {
	// New connection UI + form submission handler.
	uiPath := "GET " + desc.ConnectionURL().Path + "/"
	mux.Handle(uiPath, http.FileServer(http.FS(static.ChatGPTWebContent)))

	mux.Handle("POST "+savePath, NewHTTPHandler(l))
}
