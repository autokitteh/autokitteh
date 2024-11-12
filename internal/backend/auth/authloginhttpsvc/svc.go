package authloginhttpsvc

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authloginhttpsvc/web"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Deps struct {
	fx.In

	Muxes    *muxes.Muxes
	L        *zap.Logger
	Cfg      *Config
	DB       db.DB
	Sessions authsessions.Store
	Tokens   authtokens.Tokens
	Users    authusers.Users
}

type svc struct{ Deps }

func Init(deps Deps) error {
	svc := &svc{Deps: deps}
	return svc.registerRoutes(deps.Muxes)
}

type loginData struct {
	Email        string
	DisplayName  string
	ProviderName string
}

func (a *svc) registerRoutes(muxes *muxes.Muxes) error {
	var loginPaths []string

	if a.Cfg.GoogleOAuth.Enabled {
		if err := registerGoogleOAuthRoutes(muxes.NoAuth, a.Deps.Cfg.GoogleOAuth, a.newSuccessLoginHandler); err != nil {
			return err
		}

		loginPaths = append(loginPaths, googleLoginPath)
	}

	if a.Cfg.GithubOAuth.Enabled {
		if err := registerGithubOAuthRoutes(muxes.NoAuth, a.Cfg.GithubOAuth, a.newSuccessLoginHandler); err != nil {
			return err
		}

		loginPaths = append(loginPaths, githubLoginPath)
	}

	if a.Cfg.Descope.Enabled {
		if a.Cfg.GithubOAuth.Enabled || a.Cfg.GoogleOAuth.Enabled {
			return fmt.Errorf("cannot enable descope with other providers enabled")
		}

		if err := registerDescopeRoutes(muxes.NoAuth, a.Cfg.Descope, a.newSuccessLoginHandler); err != nil {
			return err
		}

		loginPaths = append(loginPaths, descopeLoginPath)
	}

	muxes.Auth.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		a.Deps.Sessions.Delete(w)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	muxes.NoAuth.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if len(loginPaths) == 0 {
			http.Error(w, "login is not supported", http.StatusForbidden)
			return
		}

		if len(loginPaths) == 1 {
			http.Redirect(w, r, loginPaths[0], http.StatusFound)
			return
		}

		if err := web.LoginTemplate.Execute(w, a.Cfg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	muxes.NoAuth.HandleFunc("/auth/cli-login", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get("p")
		if _, err := strconv.ParseUint(p, 10, 16); err != nil {
			http.Error(w, "invalid port", http.StatusBadRequest)
			return
		}

		url := &url.URL{
			Path:     "/auth/finish-cli-login",
			RawQuery: "p=" + p,
		}

		RedirectToLogin(w, r, url)
	})

	muxes.Auth.HandleFunc("/auth/finish-cli-login", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get("p")
		if _, err := strconv.ParseUint(p, 10, 16); err != nil {
			http.Error(w, "invalid port", http.StatusBadRequest)
			return
		}

		u := authcontext.GetAuthnUser(r.Context())
		if !u.IsValid() {
			http.Error(w, "unable to identify user", http.StatusInternalServerError)
			return
		}

		token, err := a.Tokens.Create(u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("http://localhost:%s/?token=%s", p, token), http.StatusFound)
	})

	muxes.NoAuth.HandleFunc("/auth/vscode-login", func(w http.ResponseWriter, r *http.Request) {
		RedirectToLogin(w, r, &url.URL{Path: "/auth/finish-vscode-login"})
	})

	muxes.Auth.HandleFunc("/auth/finish-vscode-login", func(w http.ResponseWriter, r *http.Request) {
		u := authcontext.GetAuthnUser(r.Context())
		if !u.IsValid() {
			http.Error(w, "unable to identify user", http.StatusInternalServerError)
			return
		}

		token, err := a.Tokens.Create(u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("vscode://autokitteh.autokitteh/authenticate?token=%s", token), http.StatusFound)
	})

	muxes.Auth.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		u := authcontext.GetAuthnUser(r.Context())
		if !u.IsValid() {
			fmt.Fprint(w, "You are not logged in")
			return
		}

		w.Header().Add("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(u); err != nil {
			a.L.Error("failed writing response", zap.Error(err))
		}
	})

	return nil
}

func (a *svc) newSuccessLoginHandler(ld *loginData) http.Handler {
	sl := a.L.Sugar().With("login_data", ld)

	fn := func(w http.ResponseWriter, r *http.Request) {
		u, err := a.Users.GetByEmail(r.Context(), ld.Email)
		if err != nil && !errors.Is(err, sdkerrors.ErrNotFound) {
			sl.With("err", err).Errorf("failed getting user by email: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		uid := u.ID()

		if u.IsValid() {
			if u.Disabled() {
				sl.Infof("disabled user: %v", ld.Email)
				http.Error(w, "user is disabled", http.StatusForbidden)
				return
			}
		} else {
			// New user.

			if a.Cfg.RejectNewUsers {
				sl.Infof("unregistered user: %v", ld.Email)
				http.Error(w, "unregistered user", http.StatusForbidden)
				return
			}

			u := sdktypes.NewUser(ld.Email, ld.DisplayName)

			if a.Cfg.UseLegacyUserIDs {
				u = u.WithID(newLegacyUserIDFromUserData(ld))

				sl.With("legacy_user_id", u.ID()).Infof("using legacy user id: %v", u.ID())
			}

			if uid, err = a.Users.Create(r.Context(), u); err != nil {
				sl.With("err", err).Errorf("failed creating user: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			sl = sl.With("user_id", uid)

			sl.With("username", ld.Email).Infof("created user %v for %q", uid, ld.Email)
		}

		sd := authsessions.NewSessionData(uid)

		if err := a.Deps.Sessions.Set(w, &sd); err != nil {
			sl.With("err", err).Errorf("failed storing session: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, getRedirect(r), http.StatusFound)
	}

	return http.HandlerFunc(fn)
}
