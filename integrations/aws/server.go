package aws

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// uiPath is the URL root path of a simple web UI to interact with users.
	uiPath = "/aws/connect/"

	// savePath is the URL path for our handler to save a new autokitteh
	// connection, after the user submits its details via a web form.
	savePath = "/aws/save"
)

func Start(mux *http.ServeMux) {
	// New connection UI + form submission handler.
	mux.Handle(uiPath, http.FileServer(http.FS(static.AWSWebContent)))
	mux.Handle(savePath, NewHTTPHandler())
}
