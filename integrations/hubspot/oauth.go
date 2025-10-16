package hubspot

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger   *zap.Logger
	vars     sdkservices.Vars
	oauth    *oauth.OAuth
	dispatch sdkservices.DispatchFunc
}

func NewHTTPHandler(l *zap.Logger, v sdkservices.Vars, o *oauth.OAuth, d sdkservices.DispatchFunc) handler {
	return handler{logger: l, oauth: o, vars: v, dispatch: d}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	// Test the OAuth token's usability and get authoritative connection details.
	url := "https://api.hubapi.com/crm/v3/owners/"
	auth := "Bearer " + oauthToken.AccessToken
	if _, err := common.HTTPGet(r.Context(), url, auth); err != nil {
		l.Warn("failed to test HubSpot OAuth token", zap.Error(err))
		c.AbortServerError("failed to test OAuth token")
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

	// Get portal ID
	accountURL := "https://api.hubapi.com/integrations/v1/me"
	accountResp, err := common.HTTPGet(r.Context(), accountURL, auth)
	if err != nil {
		l.Warn("failed to get HubSpot account details", zap.Error(err))
		c.AbortServerError("failed to get account details")
		return
	}

	// Parse response and extract portal ID.
	type HubSpotAccount struct {
		PortalId int64 `json:"portalId"`
	}

	var account HubSpotAccount
	if json.Unmarshal(accountResp, &account) == nil && account.PortalId != 0 {
		portalID := strconv.FormatInt(account.PortalId, 10)

		portal_id_var := sdktypes.NewVar(portalIDVar).SetValue(portalID)
		if err := h.vars.Set(r.Context(), portal_id_var.WithScopeID(vsid)); err != nil {
			l.Warn("failed to save portal ID", zap.Error(err))
			c.AbortServerError("failed to save portal ID")
			return
		}

		l.Info("saved HubSpot portal ID", zap.String("portalId", portalID))
	}

	c.Finalize(sdktypes.NewVars(data.ToVars()...))
}
