package sdkintegrations

import (
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ConnectionInit struct {
	Writer       http.ResponseWriter
	Request      *http.Request
	Integration  sdktypes.Integration
	ConnectionID string
	Origin       string
	logger       *zap.Logger
}

func NewConnectionInit(l *zap.Logger, w http.ResponseWriter, r *http.Request, i sdktypes.Integration) (ConnectionInit, *zap.Logger) {
	cid := r.FormValue("cid")
	origin := r.FormValue("origin")

	l = l.With(
		zap.String("urlPath", r.URL.Path),
		zap.String("connectionID", cid),
		zap.String("origin", origin),
	)

	return ConnectionInit{
		Writer:       w,
		Request:      r,
		Integration:  i,
		ConnectionID: cid,
		Origin:       origin,
		logger:       l,
	}, l
}

// Abort is the same as [AbortWithStatus] with HTTP 400 (Bad Request).
func (c ConnectionInit) Abort(err string) {
	c.AbortWithStatus(http.StatusBadRequest, err)
}

// Abort aborts the connection initialization flow due to
// a runtime error. It encodes the error status and message,
// and redirects the user to the last HTTP response, based on its
// origin (local AK server / local VS Code extension / SaaS web UI).
func (c ConnectionInit) AbortWithStatus(status int, err string) {
	origin := c.Request.FormValue("origin")
	switch origin {
	case "vscode":
		u := "vscode://autokitteh.autokitteh?cid=%s&status=%d&error=%s"
		u = fmt.Sprintf(u, c.ConnectionID, status, url.QueryEscape(err))
		http.Redirect(c.Writer, c.Request, u, http.StatusFound)
	case "web":
		http.Error(c.Writer, err, status)
	default: // Local server ("cli", "dash", etc.)
		u := c.Integration.ConnectionURL().Path + "/error.html?cid=%s&origin=%s&error=%s"
		u = fmt.Sprintf(u, c.ConnectionID, origin, url.QueryEscape(err))
		http.Redirect(c.Writer, c.Request, u, http.StatusFound)
	}
}

// Finalize finalizes all integration-specific connection flows.
// It encodes their resulting details and redirects the user to the
// final HTTP handler ("post-init") that saves the data in the connection's
// scope, and redirects the user to the last HTTP response, based on its
// origin (local AK server / local VS Code extension / SaaS web UI).
func (c ConnectionInit) Finalize(data []sdktypes.Var) {
	vars, err := kittehs.EncodeURLData(data)
	if err != nil {
		c.logger.Warn("Connection data encoding error", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError, "connection data encoding error")
		return
	}

	if _, err = sdktypes.StrictParseConnectionID(c.ConnectionID); err != nil {
		c.logger.Warn("Invalid connection ID")
		c.Abort("invalid connection ID")
		return
	}

	u := fmt.Sprintf("/connections/%s/postinit?vars=%s&origin=%s", c.ConnectionID, vars, c.Origin)
	http.Redirect(c.Writer, c.Request, u, http.StatusFound)
}
