package oauth

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const exchangeTimeout = 3 * time.Second

type svc struct {
	logger *zap.Logger
	svcs   sdkservices.Services
}

func InitWebhook(l *zap.Logger, muxes *muxes.Muxes, c sdkservices.Services) {
	s := &svc{logger: l, svcs: c}
	muxes.Auth.HandleFunc("GET /oauth/start/{intg}", s.start)
	muxes.NoAuth.HandleFunc("GET /oauth/redirect/{intg}", s.exchange)
}

func (s *svc) start(w http.ResponseWriter, r *http.Request) {
	intg := r.PathValue("intg")
	origin := r.FormValue("origin")

	l := s.logger.With(
		zap.String("urlPath", r.URL.Path),
		zap.String("integration", intg),
		zap.String("origin", origin),
	)

	// Sanity check: the specified OAuth consumer is registered.
	o := s.svcs.OAuth()
	ctx := r.Context()
	if _, _, err := o.Get(ctx, intg); err != nil {
		l.Warn("OAuth config retrieval error", zap.Error(err))
		http.Error(w, "Bad request: unrecognized config ID", http.StatusBadRequest)
		return
	}

	cid, err := sdktypes.StrictParseConnectionID(r.URL.Query().Get("cid"))
	if err != nil {
		l.Warn("Failed to parse connection ID", zap.Error(err))
		http.Error(w, "Bad request: invalid connection ID", http.StatusBadRequest)
		return
	}

	startURL, err := o.StartFlow(ctx, intg, cid, origin)
	if err != nil {
		l.Error("OAuth start flow error", zap.Error(err))
		http.Error(w, "Bad request: start flow error", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, startURL, http.StatusFound)
}

func (s *svc) exchange(w http.ResponseWriter, r *http.Request) {
	intg := r.PathValue("intg")
	state := r.FormValue("state")
	ctx := r.Context()

	l := s.logger.With(
		zap.String("urlPath", r.URL.Path),
		zap.String("integration", intg),
		zap.String("state", state),
	)

	// Ensure the specified integration is registered.
	o := s.svcs.OAuth()
	if _, _, err := o.Get(ctx, intg); err != nil {
		l.Error("Unrecognized integration name")
		http.Error(w, "Unrecognized integration name", http.StatusBadRequest)
		return
	}

	// Ensure the state parameter is well-formed.
	sub := regexp.MustCompile(`^(.+)_([a-z]+)$`).FindStringSubmatch(state)
	if len(sub) != 3 {
		l.Error("Invalid state parameter")
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Report back OAuth errors.
	if err := r.FormValue("error_description"); err != "" {
		abort(w, r, intg, sub[1], sub[2], err)
		return
	}

	if err := r.FormValue("error"); err != "" {
		abort(w, r, intg, sub[1], sub[2], err)
		return
	}

	// Special case: GitHub gives us what we need for generating JWTs in the future
	// (i.e. the GitHub app's installation ID) without exchanging the code below.
	if intg == "github" {
		l = l.With(
			zap.String("setup_action", r.FormValue("setup_action")),
			zap.String("installation_id", r.FormValue("installation_id")),
		)

		paramData, err := sdkintegrations.OAuthData{Token: nil, Params: r.URL.Query()}.Encode()
		if err != nil {
			abort(w, r, intg, sub[1], sub[2], "URL parameters encoding error")
			return
		}

		l.Info("GitHub app flow successful")
		redirect(w, r, intg, sub[1], sub[2], "oauth", paramData)
		return
	}

	// Convert the OAuth code into a refresh token / user access token.
	code := r.FormValue("code")
	if code == "" {
		l.Warn("Missing code parameter in OAuth redirect", zap.String("url", r.RequestURI))
		abort(w, r, intg, sub[1], sub[2], "missing code parameter")
		return
	}

	cid, err := sdktypes.ParseConnectionID(transformState(sub[1]))
	if err != nil {
		l.Warn("Invalid connection ID in state parameter", zap.String("state", state))
		abort(w, r, intg, sub[1], sub[2], "invalid connection ID")
		return
	}

	token, err := o.Exchange(ctx, intg, cid, code)
	if err != nil {
		l.Warn("OAuth exchange error", zap.Error(err))
		abort(w, r, intg, sub[1], sub[2], "OAuth exchange error")
		return
	}

	l.Info("OAuth token exchange successful")

	// Pass the OAuth data back to the originating integration.
	oauthData, err := sdkintegrations.OAuthData{Token: token, Params: r.URL.Query()}.Encode()
	if err != nil {
		abort(w, r, intg, sub[1], sub[2], "OAuth token encoding error")
		return
	}

	redirect(w, r, intg, sub[1], sub[2], "oauth", oauthData)
}

func abort(w http.ResponseWriter, r *http.Request, intg, cid, origin, err string) {
	redirect(w, r, intg, cid, origin, "error", url.QueryEscape(err))
}

func redirect(w http.ResponseWriter, r *http.Request, intg, cid, origin, param, value string) {
	u := fmt.Sprintf("/%s/oauth?cid=con_%s&origin=%s&%s=%s", intg, cid, origin, param, value)
	http.Redirect(w, r, u, http.StatusFound)
}

func transformState(state string) string {
	if state == "" {
		return state
	}
	return "con_" + strings.Split(state, "_")[0]
}
