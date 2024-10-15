package authhttpmiddleware

import (
	"net/http"
	"strings"

	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config struct {
	Required         bool `koanf:"required"`
	AllowDefaultUser bool `koanf:"allow_default_user"`
}

var Configs = configset.Set[Config]{
	Default: &Config{Required: true},
	Dev:     &Config{AllowDefaultUser: true},
	Test:    &Config{AllowDefaultUser: true},
}

type AuthMiddlewareDecorator func(http.Handler) http.Handler

type Deps struct {
	fx.In

	Cfg      *Config
	Users    sdkservices.Users
	Sessions authsessions.Store `optional:"true"`
	Tokens   authtokens.Tokens  `optional:"true"`
}

func newTokensMiddleware(next http.Handler, deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := authcontext.GetAuthnUser(r.Context())
		authHdr := r.Header.Get("Authorization")

		if deps.Cfg.AllowDefaultUser {
			if !user.IsValid() || authHdr != "" {
				http.Error(w, "only default user is allowed", http.StatusUnauthorized)
				return
			}
			ctx = authcontext.SetAuthnUser(ctx, sdktypes.DefaultUser)
		} else {
			if !user.IsValid() && authHdr != "" {
				kind, payload, _ := strings.Cut(authHdr, " ")
				switch kind {
				case "Bearer":
					uid, err := deps.Tokens.Parse(payload)
					if err != nil {
						http.Error(w, "invalid token", http.StatusUnauthorized)
						return
					}

					u, err := deps.Users.GetByID(r.Context(), uid)
					if err != nil {
						http.Error(w, "invalid user", http.StatusUnauthorized)
						return
					}

					ctx = authcontext.SetAuthnUser(ctx, u)

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

func newSessionsMiddleware(next http.Handler, sessions authsessions.Store, users sdkservices.Users) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if u := authcontext.GetAuthnUser(ctx); !u.IsValid() {
			session, err := sessions.Get(r)
			if err != nil {
				http.Error(w, "invalid session", http.StatusUnauthorized)
				return
			}

			if session != nil {
				u, err := users.GetByID(r.Context(), session.UserID)
				if err != nil {
					http.Error(w, "invalid user", http.StatusUnauthorized)
					return
				}

				r = r.WithContext(authcontext.SetAuthnUser(ctx, u))
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
					http.Error(w, "unauthorized user", http.StatusUnauthorized)
					return
				}
			}

			next.ServeHTTP(w, r)
		})

		if deps.Sessions != nil {
			f = newSessionsMiddleware(f, deps.Sessions, deps.Users)
		}

		if deps.Tokens != nil {
			f = newTokensMiddleware(f, deps)
		}

		return f
	}
}

type AuthHeaderExtractor func(*http.Request) sdktypes.UserID

// Auth middleware is extracting, parsing and adding user to the request context (see newTokensMiddleware above)
// Unfortunately the main log is in interceptor, and auth middleware chained several hops after the interceptor
// handler, where httpRequest is logged. So this function duplicates partly the extraction logic of
// newTokensMiddleware in order to be applied earlier in chain and log the user passed in the request.
func AuthorizationHeaderExtractor(deps Deps) AuthHeaderExtractor {
	return func(r *http.Request) sdktypes.UserID {
		if deps.Tokens == nil {
			return sdktypes.InvalidUserID
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			return sdktypes.InvalidUserID
		}

		kind, payload, _ := strings.Cut(auth, " ")
		if kind != "Bearer" {
			return sdktypes.InvalidUserID
		}

		uid, err := deps.Tokens.Parse(payload)
		if err != nil {
			return sdktypes.InvalidUserID
		}

		return uid
	}
}
