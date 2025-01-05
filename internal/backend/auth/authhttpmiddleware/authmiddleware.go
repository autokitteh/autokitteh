package authhttpmiddleware

import (
	"context"
	"net/http"
	"strings"

	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config struct {
	// If set, no authn is required. If no other authn supplied,
	// the default user will be considered authenticated.
	UseDefaultUser bool `koanf:"use_default_user"`
}

var Configs = configset.Set[Config]{
	Default: &Config{},
	Dev:     &Config{UseDefaultUser: true},
}

// The middleware passes around internally the authenticated user ID,
// so that we can check only once in the AuthMiddlewareDecorator for
// the common stuff - that the user exists and that it is not disabled.
// The AuthMiddlewareDecorator will call eventually the authcontext.SetAuthnUser
// if the user is deemed worthy of authentication.
type userIDCtxKey string

var userIDContextKey = userIDCtxKey("authn_user_id")

func ctxWithUserID(ctx context.Context, id sdktypes.UserID) context.Context {
	return context.WithValue(ctx, userIDContextKey, id)
}

func getCtxUserID(ctx context.Context) sdktypes.UserID {
	v := ctx.Value(userIDContextKey)
	if v == nil {
		return sdktypes.InvalidUserID
	}

	return v.(sdktypes.UserID)
}

type AuthMiddlewareDecorator func(http.Handler) http.Handler

type Deps struct {
	fx.In

	Cfg      *Config
	Users    sdkservices.Users
	Sessions authsessions.Store `optional:"true"`
	Tokens   authtokens.Tokens  `optional:"true"`
}

// ifAuthenticated is a middleware that checks if the user is authenticated.
// If the user is authenticated, it calls the `yes` handler, otherwise it calls the `no` handler.
func ifAuthenticated(yes, no http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next := no

		if getCtxUserID(r.Context()).IsValid() {
			next = yes
		}

		next.ServeHTTP(w, r)
	})
}

func newTokensMiddleware(next http.Handler, tokens authtokens.Tokens) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHdr := r.Header.Get("Authorization")

		if authHdr == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()

		kind, payload, _ := strings.Cut(authHdr, " ")

		if kind != "Bearer" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		u, err := tokens.Parse(payload)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctxWithUserID(ctx, u.ID())))
	}
}

func newSessionsMiddleware(next http.Handler, sessions authsessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessions.Get(r)
		// Do not fail on error - graceful degradation in case of session structure changes.
		if err == nil && session != nil {
			r = r.WithContext(ctxWithUserID(r.Context(), session.UserID))
		}

		next.ServeHTTP(w, r)
	}
}

func newSetDefaultUserMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(ctxWithUserID(r.Context(), authusers.DefaultUser.ID())))
	})
}

func New(deps Deps) AuthMiddlewareDecorator {
	sessions, users, tokens := deps.Sessions, deps.Users, deps.Tokens

	return func(next http.Handler) http.Handler {
		// Order matters here!

		// Evaluated last.
		f := ifAuthenticated(
			/* authenticated: */ http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// make sure the user exists.
				ctx := r.Context()

				uid := getCtxUserID(ctx)

				u, err := users.Get(authcontext.SetAuthnSystemUser(ctx), uid, "")
				if err != nil {
					http.Error(w, "unknown user", http.StatusUnauthorized)
					return
				}

				if u.Status() != sdktypes.UserStatusActive {
					http.Error(w, "user is not active", http.StatusUnauthorized)
					return
				}

				next.ServeHTTP(w, r.WithContext(authcontext.SetAuthnUser(ctx, u)))
			}),
			/* not authenticated: */ http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
			}),
		)

		if deps.Cfg.UseDefaultUser {
			f = ifAuthenticated(f, newSetDefaultUserMiddleware(f))
		}

		if sessions != nil {
			f = ifAuthenticated(f, newSessionsMiddleware(f, sessions))
		}

		// Evaluated first.
		if tokens != nil {
			f = ifAuthenticated(f, newTokensMiddleware(f, tokens))
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

		u, err := deps.Tokens.Parse(payload)
		if err != nil {
			return sdktypes.InvalidUserID
		}

		return u.ID()
	}
}
