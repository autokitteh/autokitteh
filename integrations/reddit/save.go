package reddit

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to save data from web form submissions as connections.
type handler struct {
	logger *zap.Logger
	vars   sdkservices.Vars
}

func NewHTTPHandler(l *zap.Logger, v sdkservices.Vars) http.Handler {
	return handler{logger: l, vars: v}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check the "Content-Type" header.
	if common.PostWithoutFormContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("save connection: unexpected POST content type", zap.String("content_type", ct))
		c.AbortBadRequest("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.AbortBadRequest("form parsing error")
		return
	}

	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		l.Warn("save connection: invalid connection ID", zap.Error(err))
		c.AbortBadRequest("invalid connection ID")
		return
	}

	common.SaveAuthType(r, h.vars, sdktypes.NewVarScopeID(cid))

	clientID := r.FormValue("client_id")
	if clientID == "" {
		l.Warn("save connection: missing client ID")
		c.AbortBadRequest("missing client ID")
		return
	}
	clientSecret := r.FormValue("client_secret")
	if clientSecret == "" {
		l.Warn("save connection: missing client secret")
		c.AbortBadRequest("missing client secret")
		return
	}
	userAgent := r.FormValue("user_agent")
	if userAgent == "" {
		l.Warn("save connection: missing user agent")
		c.AbortBadRequest("missing user agent")
		return
	}

	// username and password are optional.
	username := r.FormValue("username")
	password := r.FormValue("password")

	vs := sdktypes.NewVars(sdktypes.NewVar(clientIDVar).SetValue(clientID).SetSecret(true),
		sdktypes.NewVar(clientSecretVar).SetValue(clientSecret).SetSecret(true),
		sdktypes.NewVar(userAgentVar).SetValue(userAgent).SetSecret(true),
		sdktypes.NewVar(usernameVar).SetValue(username).SetSecret(true),
		sdktypes.NewVar(passwordVar).SetValue(password).SetSecret(true))

	if err := h.vars.Set(r.Context(), vs.WithScopeID(sdktypes.NewVarScopeID(cid))...); err != nil {
		l.Warn("failed to save vars", zap.Error(err))
		c.AbortServerError("failed to save connection variables")
	}
}
