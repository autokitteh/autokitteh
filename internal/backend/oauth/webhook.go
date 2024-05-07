package oauth

import (
	_ "embed"
	"fmt"
	"net/http"
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
	muxes.Auth.HandleFunc("/oauth/start/{intg}", s.start)
	muxes.NoAuth.HandleFunc("/oauth/redirect/{intg}", s.exchange)
}

func (s *svc) start(w http.ResponseWriter, r *http.Request) {
	intg := r.PathValue("intg")

	l := s.logger.With(zap.String("urlPath", r.URL.Path), zap.String("intg", intg))

	// Sanity check: the specified OAuth consumer is registered.
	o := s.svcs.OAuth()
	ctx := r.Context()
	if _, _, err := o.Get(ctx, r.PathValue("intg")); err != nil {
		l.Warn("OAuth config retrieval error", zap.Error(err))
		http.Error(w, "Bad request: unrecognized config ID", http.StatusBadRequest)
		return
	}

	cid, err := sdktypes.ParseConnectionID(r.URL.Query().Get("cid"))
	if err != nil {
		l.Warn("Failed to parse connection ID", zap.Error(err))
		http.Error(w, "Bad request: invalid connection ID", http.StatusBadRequest)
		return
	}

	startURL, err := o.StartFlow(ctx, intg, cid)
	if err != nil {
		l.Error("OAuth start flow error", zap.Error(err))
		http.Error(w, "Bad request: start flow error", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, startURL, http.StatusFound)
}

func (s *svc) exchange(w http.ResponseWriter, r *http.Request) {
	intg := r.PathValue("intg")

	l := s.logger.With(zap.String("url_path", r.URL.Path), zap.String("intg", intg))

	abort := func(msg string, err error, code int) {
		l.Warn(msg, zap.Error(err))
		http.Error(w, msg, code)
	}

	ctx := r.Context()

	// ensure the specified integration is registered.
	o := s.svcs.OAuth()
	if _, _, err := o.Get(ctx, r.PathValue("intg")); err != nil {
		abort("unrecognized integration name", err, http.StatusBadRequest)
		return
	}

	// Read the temporary authorization code from the redirected request.
	code := r.FormValue("code")
	if code == "" {
		abort("missing code parameter", nil, http.StatusBadRequest)
		return
	}

	state := r.FormValue("state")

	// Convert it into an OAuth refresh token / user access token.
	token, err := o.Exchange(ctx, intg, state, code)
	if err != nil {
		abort("OAuth exchange error", err, http.StatusBadRequest)
		return
	}

	l.Info("OAuth token exchange successful")

	oauthData, err := sdkintegrations.OAuthData{Token: token, Params: r.URL.Query()}.Encode()
	if err != nil {
		abort("failed to encode OAuth token", err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/oauth?oauth=%s&cid=%s", intg, oauthData, state), http.StatusFound)
}
