package authhttpmiddleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config struct {
	Required bool `koanf:"required"`

	UseAnonymous bool `koanf:"use_anonymous"` // set user as "anonymous" if !required and no auth given.

	AllowFakeTestAuth bool `koanf:"allow_fake_test_auth"` // allow passing raw user names in auth header for testing.
}

var Configs = configset.Set[Config]{
	Default: &Config{
		Required: true,
	},
	Dev: &Config{},
	Test: &Config{
		AllowFakeTestAuth: true,
	},
}

type WrapFunc func(http.Handler) http.Handler

type Deps struct {
	fx.In

	Cfg      *Config
	Sessions authsessions.Store `optional:"true"`
	Tokens   authtokens.Tokens  `optional:"true"`
	Users    sdkservices.Users  `optional:"true"`
}

func newFakeTestAuth(next http.Handler, users sdkservices.Users) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if userID := authcontext.GetAuthnUserID(ctx); !userID.IsValid() {
			if auth := r.Header.Get("Authorization"); auth != "" {
				h, err := sdktypes.Strict(sdktypes.ParseSymbol(auth))
				if err != nil {
					http.Error(w, "invalid user name", http.StatusUnauthorized)
					return
				}

				user, err := users.GetByName(ctx, h)
				if err != nil {
					http.Error(w, fmt.Sprintf("get user error: %v", err), http.StatusInternalServerError)
					return
				}

				if !user.IsValid() {
					http.Error(w, "user not found", http.StatusUnauthorized)
					return
				}

				r = r.WithContext(authcontext.SetAuthnUserID(ctx, user.ID()))
			}
		}

		next.ServeHTTP(w, r)
	}
}

func newTokensMiddleware(next http.Handler, tokens authtokens.Tokens) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if userID := authcontext.GetAuthnUserID(ctx); !userID.IsValid() {
			if auth := r.Header.Get("Authorization"); auth != "" {
				kind, payload, _ := strings.Cut(auth, " ")

				switch kind {
				case "Bearer":
					var err error
					if userID, err = tokens.Parse(payload); err != nil {
						http.Error(w, "invalid token", http.StatusUnauthorized)
						return
					}
				default:
					http.Error(w, "invalid authorization header", http.StatusUnauthorized)
					return
				}
			}

			r = r.WithContext(authcontext.SetAuthnUserID(ctx, userID))
		}

		next.ServeHTTP(w, r)
	}
}

func newSessionsMiddleware(next http.Handler, sessions authsessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if userID := authcontext.GetAuthnUserID(ctx); !userID.IsValid() {
			session, err := sessions.Get(r)
			if err != nil {
				http.Redirect(w, r, fmt.Sprintf("/error.html?err=%s", url.QueryEscape(err.Error())), http.StatusTemporaryRedirect)
				return
			}

			if session != nil {
				r = r.WithContext(authcontext.SetAuthnUserID(ctx, userID))
			}
		}

		next.ServeHTTP(w, r)
	}
}

func New(deps Deps) WrapFunc {
	return func(next http.Handler) http.Handler {
		f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID := authcontext.GetAuthnUserID(ctx)

			if !userID.IsValid() {
				if deps.Cfg.Required {
					http.Redirect(w, r, "/login.html", http.StatusTemporaryRedirect)
					return
				}

				if deps.Cfg.UseAnonymous {
					userID = fixtures.AutokittehAnonymousUserID
				}
			}

			if userID.IsValid() {
				ctx = authcontext.SetAuthnUserID(ctx, userID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})

		if deps.Sessions != nil {
			f = newSessionsMiddleware(f, deps.Sessions)
		}

		if deps.Tokens != nil {
			f = newTokensMiddleware(f, deps.Tokens)
		}

		if deps.Cfg.AllowFakeTestAuth && deps.Users != nil {
			f = newFakeTestAuth(f, deps.Users)
		}

		return f
	}
}
