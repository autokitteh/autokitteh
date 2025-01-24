package authhttpmiddleware

import (
	"errors"
	"net/http"
	"strings"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/users"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type AuthMiddlewareDecorator func(http.Handler) http.Handler

type Deps struct {
	fx.In

	Logger   *zap.Logger
	Users    users.Users
	Sessions authsessions.Store `optional:"true"`
	Tokens   authtokens.Tokens  `optional:"true"`
}

type middlewareError struct {
	code int
	msg  string // must never contain sensitive data.
}

func (m *middlewareError) apply(w http.ResponseWriter) { http.Error(w, m.msg, m.code) }

var (
	invalidAuthHeaderErr = &middlewareError{http.StatusUnauthorized, "invalid authorization header"}
	invalidTokenErr      = &middlewareError{http.StatusUnauthorized, "invalid token"}
)

// middlewareFn is a function that extracts a user ID from a request.
// It returns the user ID and an error if the request is invalid.
type middlewareFn func(*http.Request) (sdktypes.UserID, *middlewareError)

func newTokensMiddleware(tokens authtokens.Tokens) middlewareFn {
	return func(r *http.Request) (sdktypes.UserID, *middlewareError) {
		authHdr := r.Header.Get("Authorization")
		if authHdr == "" {
			return sdktypes.InvalidUserID, nil
		}

		kind, payload, _ := strings.Cut(authHdr, " ")

		if kind != "Bearer" {
			return sdktypes.InvalidUserID, invalidAuthHeaderErr
		}

		u, err := tokens.Parse(payload)
		if err != nil {
			return sdktypes.InvalidUserID, invalidTokenErr
		}

		return u.ID(), nil
	}
}

func newSessionsMiddleware(sessions authsessions.Store) middlewareFn {
	return func(r *http.Request) (sdktypes.UserID, *middlewareError) {
		session, err := sessions.Get(r)
		// Do not fail on error - graceful degradation in case of session structure changes.
		if err == nil && session != nil {
			return session.UserID, nil
		}

		return sdktypes.InvalidUserID, nil
	}
}

func setDefaultUserMiddleware(r *http.Request) (sdktypes.UserID, *middlewareError) {
	return authusers.DefaultUser.ID(), nil
}

func New(deps Deps) AuthMiddlewareDecorator {
	sessions, users, tokens := deps.Sessions, deps.Users, deps.Tokens

	var mws []middlewareFn

	if tokens != nil {
		mws = append(mws, newTokensMiddleware(tokens))
	}

	if sessions != nil {
		mws = append(mws, newSessionsMiddleware(sessions))
	}

	if deps.Users.HasDefaultUser() {
		mws = append(mws, setDefaultUserMiddleware)
	}

	return func(next http.Handler) http.Handler { // = AuthMiddlewareDecorator
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Process all middlewares.
			var (
				uid   sdktypes.UserID
				mwErr *middlewareError
			)

			for _, mw := range mws {
				// Stop processing middlewares if a valid user ID is found or there is an error.
				if uid, mwErr = mw(r); uid.IsValid() || mwErr != nil {
					break
				}
			}

			l := deps.Logger

			if mwErr != nil {
				l.Info("auth middleware error", zap.Error(errors.New(mwErr.msg)))
				mwErr.apply(w)
				return
			}

			// Check if authenticated.
			if !uid.IsValid() {
				l.Info("not authenticated")
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				return
			}

			// Hydrate user and check if it's active.
			l = l.With(zap.String("user_id", uid.String()))

			ctx := r.Context()

			u, err := users.Get(authcontext.SetAuthnSystemUser(ctx), uid, "")
			if err != nil {
				l.Warn("failed to get user", zap.Error(err))
				http.Error(w, "unknown user", http.StatusUnauthorized)
				return
			}

			if u.Status() != sdktypes.UserStatusActive {
				l.Info("user is not active")
				http.Error(w, "user is not active", http.StatusUnauthorized)
				return
			}

			ctx = authcontext.SetAuthnUser(ctx, u)

			// Propagate the request to the next handler.
			next.ServeHTTP(w, r.WithContext(ctx))
		})
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
