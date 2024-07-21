package authhttpmiddleware

import (
	"net/http"
	"strings"

	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authloginhttpsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	cctx "go.autokitteh.dev/autokitteh/internal/context"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config struct {
	Required bool `koanf:"required"`
}

var Configs = configset.Set[Config]{
	Default: &Config{Required: true},
	Dev:     &Config{},
}

type AuthMiddlewareDecorator func(http.Handler) http.Handler

type Deps struct {
	fx.In

	Cfg      *Config
	Sessions authsessions.Store `optional:"true"`
	Tokens   authtokens.Tokens  `optional:"true"`
}

func newTokensMiddleware(next http.Handler, tokens authtokens.Tokens) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = cctx.WithRequestOrinator(ctx, cctx.Middleware)

		if user := authcontext.GetAuthnUser(ctx); !user.IsValid() {
			if auth := r.Header.Get("Authorization"); auth != "" {
				kind, payload, _ := strings.Cut(auth, " ")

				switch kind {
				case "Bearer":
					var err error
					if user, err = tokens.Parse(payload); err != nil {
						http.Error(w, "invalid token", http.StatusUnauthorized)
						return
					}
					ctx = authcontext.SetAuthnUser(ctx, user)

				default:
					http.Error(w, "invalid authorization header", http.StatusUnauthorized)
					return
				}
			}
		}
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}
}

func newSessionsMiddleware(next http.Handler, sessions authsessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if user := authcontext.GetAuthnUser(ctx); !user.IsValid() {
			session, err := sessions.Get(r)
			if err != nil {
				http.Error(w, "invalid session", http.StatusUnauthorized)
				return
			}

			if session != nil {
				r = r.WithContext(authcontext.SetAuthnUser(ctx, session.User))
			}
		}

		next.ServeHTTP(w, r)
	}
}

func New(deps Deps) AuthMiddlewareDecorator {
	return func(next http.Handler) http.Handler {
		f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if deps.Cfg.Required {
				if user := authcontext.GetAuthnUser(r.Context()); !user.IsValid() {
					authloginhttpsvc.RedirectToLogin(w, r, r.URL)
					return
				}
			}

			next.ServeHTTP(w, r)
		})

		if deps.Sessions != nil {
			f = newSessionsMiddleware(f, deps.Sessions)
		}

		if deps.Tokens != nil {
			f = newTokensMiddleware(f, deps.Tokens)
		}

		return f
	}
}

type AuthHeaderExtractor func(*http.Request) sdktypes.User

// Auth middleware is extracting, parsing and adding user to the request context (see newTokensMiddleware above)
// Unfortunately the main log is in interceptor, and auth middleware chained several hops after the interceptor
// handler, where httpRequest is logged. So this function duplicates partly the extraction logic of
// newTokensMiddleware in order to be applied earlier in chain and log the user passed in the request.
func AuthorizationHeaderExtractor(deps Deps) AuthHeaderExtractor {
	return func(r *http.Request) sdktypes.User {
		if deps.Tokens == nil {
			return sdktypes.InvalidUser
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			return sdktypes.InvalidUser
		}

		kind, payload, _ := strings.Cut(auth, " ")
		if kind != "Bearer" {
			return sdktypes.InvalidUser
		}

		user, err := deps.Tokens.Parse(payload)
		if err != nil {
			return sdktypes.InvalidUser
		}

		return user
	}
}
