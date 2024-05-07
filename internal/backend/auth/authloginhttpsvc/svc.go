package authloginhttpsvc

import (
	_ "embed"
	"fmt"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authloginhttpsvc/web"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Deps struct {
	fx.In

	Muxes    *muxes.Muxes
	Z        *zap.Logger
	Cfg      *Config
	DB       db.DB
	Sessions authsessions.Store
	Tokens   authtokens.Tokens
}

type svc struct {
	Deps
}

func Init(deps Deps) error {
	svc := &svc{
		Deps: deps,
	}

	return svc.registerRoutes(deps.Muxes)
}

func (a *svc) registerRoutes(muxes *muxes.Muxes) error {
	if a.Cfg.GoogleOAuth.Enabled {
		if err := registerGoogleOAuthRoutes(muxes.NoAuth, a.Deps.Cfg.GoogleOAuth, a.newSuccessLoginHandler); err != nil {
			return err
		}
	}

	if a.Cfg.GithubOAuth.Enabled {
		if err := registerGithubOAuthRoutes(muxes.NoAuth, a.Cfg.GithubOAuth, a.newSuccessLoginHandler); err != nil {
			return err
		}
	}

	if a.Cfg.Descope.Enabled {
		if a.Cfg.GithubOAuth.Enabled || a.Cfg.GoogleOAuth.Enabled {
			return fmt.Errorf("cannot enable descope with other providers enabled")
		}

		if err := registerDescopeRoutes(muxes.NoAuth, a.Cfg.Descope, a.newSuccessLoginHandler); err != nil {
			return err
		}
	}

	muxes.Auth.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		a.Deps.Sessions.Delete(w)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	muxes.NoAuth.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if err := web.LoginTemplate.Execute(w, a.Cfg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	muxes.Auth.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		u := authcontext.GetAuthnUser(r.Context())
		if !u.IsValid() {
			fmt.Fprint(w, "You are not logged in")
			return
		}

		bs, err := u.MarshalJSON()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		if _, err := w.Write(bs); err != nil {
			a.Z.Error("failed writing response", zap.Error(err))
		}
	})

	return nil
}

func (a *svc) newSuccessLoginHandler(user sdktypes.User) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if !user.IsValid() {
			a.Z.Warn("user not found")
			http.Error(w, "user not found", http.StatusForbidden)
			return
		}

		sd := authsessions.SessionData{User: user}

		if err := a.Deps.Sessions.Set(w, &sd); err != nil {
			a.Z.Error("failed storing session", zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, req, getRedirect(req), http.StatusFound)
	}

	return http.HandlerFunc(fn)
}
