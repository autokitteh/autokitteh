package auth0

import (
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type handler struct {
	logger *zap.Logger
	vars   sdkservices.Vars
}

func NewHTTPHandler(l *zap.Logger, vars sdkservices.Vars) handler {
	return handler{logger: l, vars: vars}
}

func (h handler) handleOAuth(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Handle errors (e.g. the user didn't authorize us) based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirect request reported an error", zap.Error(errors.New(e)))
		c.AbortBadRequest(e)
		return
	}

	_, data, err := sdkintegrations.GetOAuthDataFromURL(r.URL)
	if err != nil {
		l.Warn("Invalid data in OAuth redirect request", zap.Error(err))
		c.AbortBadRequest("invalid data parameter")
		return
	}

	oauthToken := data.Token
	if oauthToken == nil {
		l.Warn("Missing token in OAuth redirect request", zap.Any("data", data))
		c.AbortBadRequest("missing OAuth token")
		return
	}

	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		l.Error("Failed to parse connection ID", zap.Error(err))
		c.AbortBadRequest("invalid connection ID")
		return
	}
	vs, err := h.vars.Get(r.Context(), sdktypes.NewVarScopeID(cid))
	if err != nil {
		l.Error("Failed to get Auth0 vars", zap.Error(err))
		c.AbortServerError("unknown Auth0 domain")
		return
	}

	// Test the OAuth token's usability and get authoritative installation details.
	d := vs.GetValueByString("auth0_domain")
	if d == "" {
		l.Error("Missing Auth0 domain in connection vars")
		c.AbortServerError("unknown Auth0 domain")
		return
	}

	// Tests Auth0's Management API.
	url := fmt.Sprintf("https://%s/api/v2/roles", d)
	auth := "Bearer " + oauthToken.AccessToken
	if _, err := common.HTTPGet(r.Context(), url, auth); err != nil {
		l.Warn("failed to test Auth0 OAuth token", zap.Error(err))
		c.AbortServerError("failed to get the OAuth token's roles")
		return
	}

	c.Finalize(sdktypes.NewVars(data.ToVars()...).
		Append(sdktypes.NewVar(common.AuthTypeVar).SetValue(integrations.OAuth)))
}
