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
	muxes.AuthHandleFunc("/oauth/start/{intg}", s.start)
	muxes.HandleFunc("/oauth/redirect/{intg}", s.exchange)
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
	intgSymbol, err := sdktypes.ParseSymbol(intg)
	if err != nil {
		http.Error(w, "invalid integration name", http.StatusBadRequest)
		return
	}

	l := s.logger.With(zap.String("url_path", r.URL.Path), zap.String("intg", intg), zap.Any("form", r.Form))

	ctx := r.Context()
	i, err := s.svcs.Integrations().GetByName(ctx, intgSymbol)
	if err != nil {
		l.Warn("Failed to retrieve integration", zap.Error(err))
		http.Error(w, "failed to retrieve integration", http.StatusInternalServerError)
		return

	}

	conns, err := s.svcs.Connections().List(ctx, sdkservices.ListConnectionsFilter{
		IntegrationID: i.Get().ID(),
	})
	if err != nil {
		l.Warn("Failed to retrieve connections", zap.Error(err))
		http.Error(w, "failed to retrieve connections", http.StatusInternalServerError)
		return
	}

	// ensure the specified integration is registered.
	o := s.svcs.OAuth()
	if _, _, err := o.Get(ctx, r.PathValue("intg")); err != nil {
		l.Warn("OAuth config retrieval error", zap.Error(err))
		http.Error(w, "unrecognized integration name", http.StatusBadRequest)
		return
	}

	// Read the temporary authorization code from the redirected request.
	code := r.FormValue("code")
	if code == "" {
		l.Warn("OAuth redirect request without code parameter")
		http.Error(w, "missing code parameter", http.StatusBadRequest)
		return
	}

	state := r.FormValue("state")

	// Convert it into an OAuth refresh token / user access token.
	token, err := o.Exchange(ctx, intg, state, code)
	if err != nil {
		l.Warn("OAuth exchange error", zap.Error(err))
		http.Error(w, "exchange error", http.StatusBadRequest)
		return
	}

	l.Info("OAuth token exchange successful")

	oauthData, err := sdkintegrations.OAuthData{Token: token, Params: r.URL.Query()}.Encode()
	if err != nil {
		l.Warn("Failed to encode OAuth token", zap.Error(err))
		http.Error(w, "failed to encode OAuth token", http.StatusInternalServerError)
		return
	}

	type conn struct {
		ID          string
		Name        string
		ProjectName string
	}

	data := struct {
		Intg      string
		OAuthData string
		Conns     []conn
	}{
		Intg:      intg,
		OAuthData: oauthData,
	}

	pnames := make(map[sdktypes.ProjectID]string)

	for _, c := range conns {
		pname := pnames[c.ProjectID()]
		if pname == "" {
			l := l.With(zap.String("project_id", c.ProjectID().String()))

			p, err := s.svcs.Projects().GetByID(ctx, c.ProjectID())
			if err != nil {
				l.Warn("Failed to retrieve project", zap.Error(err))
				continue
			}

			if !p.IsValid() {
				l.Warn("project not found")
				continue
			}

			pname = p.Name().String()
		}

		data.Conns = append(data.Conns, conn{
			ID:          c.ID().String(),
			Name:        c.Name().String(),
			ProjectName: pname,
		})
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/oauth?oauth=%s&cid=%s", intg, oauthData, state), http.StatusFound)
}
