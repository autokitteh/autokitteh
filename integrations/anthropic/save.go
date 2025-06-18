package anthropic

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

// ServeHTTP saves a new autokitteh connection with user-submitted data.
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

	vsid := sdktypes.NewVarScopeID(cid)
	common.SaveAuthType(r, h.vars, vsid)

	apiKey := r.FormValue("api_key")
	if apiKey == "" {
		l.Warn("save connection: missing API key")
		c.AbortBadRequest("missing API key")
		return
	}

	vs := sdktypes.NewVars(sdktypes.NewVar(common.ApiKeyVar).SetValue(apiKey).SetSecret(true))
	if err := h.vars.Set(r.Context(), vs.WithScopeID(vsid)...); err != nil {
		l.Warn("failed to save vars", zap.Error(err))
		c.AbortServerError("failed to save connection variables")
	}
}
