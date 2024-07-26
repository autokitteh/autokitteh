package google

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"go.autokitteh.dev/autokitteh/integrations/google/forms"
	"go.autokitteh.dev/autokitteh/integrations/google/gmail"
	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handleOAuth receives an incoming redirect request from AutoKitteh's OAuth
// management service. This request contains an OAuth token (if the OAuth
// flow was successful) and form parameters for debugging and validation
// (either way). If all is well, it saves a new AutoKitteh connection.
// Either way, it redirects the user to success or failure webpages.
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

	raw, data, err := sdkintegrations.GetOAuthDataFromURL(r.URL)
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

	// Test the OAuth token's usability and get authoritative installation details.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	src := h.tokenSource(ctx, oauthToken)
	svc, err := googleoauth2.NewService(ctx, option.WithTokenSource(src))
	if err != nil {
		l.Warn("OAuth user token error", zap.Error(err))
		c.AbortBadRequest("token source")
		return
	}

	user, err := h.getUserDetails(l, svc)
	if err != nil {
		l.Warn("OAuth user details error", zap.Error(err))
		c.AbortBadRequest("Google user details error")
		return
	}

	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		l.Warn("Invalid connection ID", zap.Error(err))
		c.AbortBadRequest("invalid connection ID")
		return
	}

	// Unique step for Google integrations (specifically for Gmail and Forms):
	// save the auth data before creating/updating event watches.
	vs := sdktypes.NewVars(sdktypes.NewVar(vars.OAuthData, raw, true)).
		Set(vars.JSON, "", true).Append(data.ToVars()...).Append(user...)

	vsl := kittehs.TransformMapToList(vs.ToMap(), func(_ sdktypes.Symbol, v sdktypes.Var) sdktypes.Var {
		return v.WithScopeID(sdktypes.NewVarScopeID(cid))
	})

	if err := h.vars.Set(ctx, vsl...); err != nil {
		l.Error("Connection data saving error", zap.Error(err))
		c.AbortServerError("connection data saving error")
		return
	}

	if err := forms.UpdateWatches(ctx, h.vars, cid); err != nil {
		l.Error("Google Forms watches creation error", zap.Error(err))
		c.AbortServerError("form watches creation error")
		return
	}

	if err := gmail.UpdateWatch(ctx, h.vars, cid); err != nil {
		l.Error("Gmail watch creation error", zap.Error(err))
		c.AbortServerError("Gmail watch creation error")
		return
	}

	// Encoding "OAuthData" and "JSON", but not "FormID", so we don't overwrite
	// the value that was already written there by the creds.go passthrough.
	c.Finalize(vs)
}

func (h handler) tokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	cfg, _, err := h.oauth.Get(ctx, "google")
	if err != nil {
		return nil
	}
	return cfg.TokenSource(ctx, t)
}

func (h handler) getUserDetails(l *zap.Logger, svc *googleoauth2.Service) (sdktypes.Vars, error) {
	ui, err := svc.Userinfo.V2.Me.Get().Do()
	if err != nil {
		l.Warn("OAuth user info retrieval error", zap.Error(err))
		return nil, err
	}
	if ui.Email == "" || !*ui.VerifiedEmail {
		l.Warn("OAuth user info is bad",
			zap.Any("userInfo", ui),
			zap.Error(err),
		)
		return nil, err
	}

	ti, err := svc.Tokeninfo().Do()
	if err != nil {
		l.Warn("OAuth token info retrieval error",
			zap.Any("userInfo", ui),
			zap.Error(err),
		)
		return nil, err
	}
	if ti.Email != ui.Email {
		l.Warn("OAuth token info is bad",
			zap.Any("userInfo", ui),
			zap.Any("tokenInfo", ti),
			zap.Error(err),
		)
		return nil, err
	}

	return sdktypes.NewVars().
		Set(sdktypes.NewSymbol("id"), ui.Id, false).
		Set(sdktypes.NewSymbol("email"), ui.Email, false).
		Set(sdktypes.NewSymbol("name"), ui.Name, false).
		Set(sdktypes.NewSymbol("given_name"), ui.GivenName, false).
		Set(sdktypes.NewSymbol("family_name"), ui.FamilyName, false).
		Set(sdktypes.NewSymbol("hd"), ui.Hd, false).
		Set(sdktypes.NewSymbol("scope"), ti.Scope, false).
		WithPrefix("user_"), nil
}
