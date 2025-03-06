package sdkintegrations

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

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

// AbortBadRequest is the same as [AbortWithStatus] with HTTP 400 (Bad Request).
func (c ConnectionInit) AbortBadRequest(err string) {
	c.AbortWithStatus(http.StatusBadRequest, err)
}

// AbortServerError is the same as [AbortWithStatus] with HTTP 500 (Internal Server Error).
func (c ConnectionInit) AbortServerError(err string) {
	c.AbortWithStatus(http.StatusInternalServerError, err)
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
	default: // SaaS web UI (non-OAuth connections) / local server ("cli", "dash", etc.)
		http.Error(c.Writer, err, status)
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
		c.logger.Error("Connection data encoding error", zap.Error(err))
		c.AbortServerError("connection data encoding error")
		return
	}

	if _, err = sdktypes.StrictParseConnectionID(c.ConnectionID); err != nil {
		c.logger.Warn("Invalid connection ID")
		c.AbortBadRequest("invalid connection ID")
		return
	}

	u := fmt.Sprintf("/connections/%s/postinit?vars=%s&origin=%s", c.ConnectionID, vars, c.Origin)
	http.Redirect(c.Writer, c.Request, u, http.StatusFound)
}

// FinalURL returns the final redirect URL at the end of the connection flow in
// each integration, based on the origin of the connection initialization request.
// TODO(INT-195): Deprecate "Finalize" above and the "postinit" webhook.
func (c ConnectionInit) FinalURL() (string, error) {
	// Security check: connection ID and origin must be alphanumeric
	// strings, to prevent path traversal attacks and other issues.
	re := regexp.MustCompile(`^[\w]+$`)
	if !re.MatchString(c.ConnectionID) || !re.MatchString(c.Origin) {
		return "", errors.New("invalid connection ID or origin")
	}

	var u string
	switch c.Origin {
	case "vscode":
		u = "vscode://autokitteh.autokitteh?cid=%s"
	case "dash":
		u = "/internal/dashboard/connections/%s?msg=Success"
	default:
		// Another redirect just to get rid of the secrets in the URL.
		u = "/connections/%s/success"
	}
	return fmt.Sprintf(u, c.ConnectionID), nil
}
