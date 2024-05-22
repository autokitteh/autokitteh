package scheduler

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// uiPath is the URL root path of a simple web UI to interact with users.
	uiPath = "/scheduler/connect/"

	// savePath is the URL path for our handler to save a new autokitteh
	// connection, after the user submits its details via a web form.
	savePath = "/scheduler/save"
)

func Start(l *zap.Logger, mux *http.ServeMux, s sdkservices.Secrets, d sdkservices.Dispatcher) {
	// New connection UI + form submission handler.
	mux.Handle(uiPath, http.FileServer(http.FS(static.SchedulerWebContent)))
	mux.Handle(savePath, NewHTTPHandler(l, s, "scheduler"))

	// In-memory cron table to dispatch events.
	go func() {
		ticker := time.NewTicker(initInterval)
		defer ticker.Stop()
		for range ticker.C {
			detectNewConnections(l, s, d)
		}
	}()
}
