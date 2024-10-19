package authloginhttpsvc

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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

	loginMatcher func(string) bool
}

func Init(deps Deps) error {
	svc := &svc{
		Deps:         deps,
		loginMatcher: compileLoginMatchers(strings.Split(deps.Cfg.AllowedLogins, ",")),
	}

	return svc.registerRoutes(deps.Muxes)
}

func (a *svc) registerRoutes(muxes *muxes.Muxes) error {
	var loginPaths []string

	if a.Cfg.GoogleOAuth.Enabled {
		if err := registerGoogleOAuthRoutes(muxes.Main.NoAuth, a.Deps.Cfg.GoogleOAuth, a.newSuccessLoginHandler); err != nil {
			return err
		}

		loginPaths = append(loginPaths, googleLoginPath)
	}

	if a.Cfg.GithubOAuth.Enabled {
		if err := registerGithubOAuthRoutes(muxes.Main.NoAuth, a.Cfg.GithubOAuth, a.newSuccessLoginHandler); err != nil {
			return err
		}

		loginPaths = append(loginPaths, githubLoginPath)
	}

	if a.Cfg.Descope.Enabled {
		if a.Cfg.GithubOAuth.Enabled || a.Cfg.GoogleOAuth.Enabled {
			return fmt.Errorf("cannot enable descope with other providers enabled")
		}

		if err := registerDescopeRoutes(muxes.Main.NoAuth, a.Cfg.Descope, a.newSuccessLoginHandler); err != nil {
			return err
		}

		loginPaths = append(loginPaths, descopeLoginPath)
	}

	muxes.Main.Auth.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		a.Deps.Sessions.Delete(w)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	muxes.Main.NoAuth.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
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

	muxes.Main.NoAuth.HandleFunc("/auth/cli-login", func(w http.ResponseWriter, r *http.Request) {
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

	muxes.Main.Auth.HandleFunc("/auth/finish-cli-login", func(w http.ResponseWriter, r *http.Request) {
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

	muxes.Main.NoAuth.HandleFunc("/auth/vscode-login", func(w http.ResponseWriter, r *http.Request) {
		RedirectToLogin(w, r, &url.URL{Path: "/auth/finish-vscode-login"})
	})

	muxes.Main.Auth.HandleFunc("/auth/finish-vscode-login", func(w http.ResponseWriter, r *http.Request) {
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

	muxes.Main.Auth.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
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

		if !a.loginMatcher(user.Login()) {
			a.Z.Warn("user not allowed to login", zap.String("user", user.Login()))
			http.Error(w, "user not allowed to login", http.StatusForbidden)
			return
		}

		sd := authsessions.NewSessionData(user)

		if err := a.Deps.Sessions.Set(w, &sd); err != nil {
			a.Z.Error("failed storing session", zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, req, getRedirect(req), http.StatusFound)
	}

	return http.HandlerFunc(fn)
}
