package google

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// uiPath is the URL root path of a simple web UI to interact with users at
	// the beginning and the end of their 3-legged OAuth v2 flow with Google.
	uiPath = "/google/connect/"

	// oauthPath is the URL path for our handler to save new OAuth-based connections.
	oauthPath = "/google/oauth"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger *zap.Logger
	vars   sdkservices.Vars
	oauth  sdkservices.OAuth
}

func NewHTTPHandler(l *zap.Logger, vars sdkservices.Vars, o sdkservices.OAuth) handler {
	return handler{logger: l, vars: vars, oauth: o}
}

// HandleOAuth receives an inbound redirect request from autokitteh's OAuth
// management service. This request contains an OAuth token (if the OAuth
// flow was successful) and form parameters for debugging and validation
// (either way). If all is well, it saves a new autokitteh connection.
// Either way, it redirects the user to success or failure webpages.
func (h handler) HandleOAuth(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Handle errors (e.g. the user didn't authorize us) based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirect request reported an error", zap.Error(errors.New(e)))
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, e)
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	rawOAuthData, oauthData, err := sdkintegrations.GetOAuthDataFromURL(r.URL)
	if err != nil {
		l.Warn("Failed to decode OAuth data", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return

	}

	// Test the OAuth token's usability and get authoritative installation details.
	src := h.tokenSource(r.Context(), oauthData.Token)
	svc, err := googleoauth2.NewService(r.Context(), option.WithTokenSource(src))
	if err != nil {
		l.Warn("OAuth user token error", zap.Error(err))
		http.Error(w, "Internal Server Error: token source", http.StatusInternalServerError)
		return
	}

	uvars, err := h.getUserDetails(svc)
	if err != nil {
		l.Warn("OAuth user details error", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	initData := sdktypes.EncodeVars(&vars.Vars{OAuthData: rawOAuthData}).
		Append(oauthData.ToVars()...).
		Append(uvars...)

	sdkintegrations.FinalizeConnectionInit(w, r, integrationID, initData)
}

func (h handler) tokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	cfg, _, err := h.oauth.Get(ctx, "google")
	if err != nil {
		return nil
	}
	return cfg.TokenSource(ctx, t)
}

func (h handler) getUserDetails(svc *googleoauth2.Service) (sdktypes.Vars, error) {
	ui, err := svc.Userinfo.V2.Me.Get().Do()
	if err != nil {
		h.logger.Warn("OAuth user info retrieval error",
			zap.Error(err),
		)
		return nil, err
	}
	if ui.Email == "" || !*ui.VerifiedEmail {
		h.logger.Warn("OAuth user info is bad",
			zap.Any("userInfo", ui),
			zap.Error(err),
		)
		return nil, err
	}

	ti, err := svc.Tokeninfo().Do()
	if err != nil {
		h.logger.Warn("OAuth token info retrieval error",
			zap.Any("userInfo", ui),
			zap.Error(err),
		)
		return nil, err
	}
	if ti.Email != ui.Email {
		h.logger.Warn("OAuth token info is bad",
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
