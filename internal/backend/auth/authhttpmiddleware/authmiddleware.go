package authhttpmiddleware

import (
	"errors"
	"fmt"
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

type (
	AuthMiddlewareDecorator func(http.Handler) http.Handler
	GetUserFromRequestFunc  func(*http.Request) (sdktypes.User, error)
)

type Deps struct {
	fx.In

	Logger   *zap.Logger
	Users    users.Users
	Sessions authsessions.Store `optional:"true"`
	Tokens   authtokens.Tokens  `optional:"true"`
}

type MiddlewareError struct {
	code int
	msg  string // must never contain sensitive data.
}

func (err *MiddlewareError) Error() string { return fmt.Sprintf("code %d: %s", err.code, err.msg) }

var (
	errInvalidAuthHeader = &MiddlewareError{http.StatusUnauthorized, "invalid authorization header"}
	errInvalidToken      = &MiddlewareError{http.StatusUnauthorized, "invalid token"}
	errUserIsNotActive   = &MiddlewareError{http.StatusUnauthorized, "user is not active"}
	errUnknownUser       = &MiddlewareError{http.StatusUnauthorized, "unknown user"}
)

// middlewareFn is a function that extracts a user ID from a request.
// It returns the user ID and an error if the request is invalid.
type middlewareFn func(*http.Request) (sdktypes.UserID, *MiddlewareError)

func newTokensMiddleware(tokens authtokens.Tokens) middlewareFn {
	return func(r *http.Request) (sdktypes.UserID, *MiddlewareError) {
		authHdr := r.Header.Get("Authorization")
		if authHdr == "" {
			return sdktypes.InvalidUserID, nil
		}

		kind, payload, _ := strings.Cut(authHdr, " ")

		if kind != "Bearer" {
			return sdktypes.InvalidUserID, errInvalidAuthHeader
		}

		u, err := tokens.Parse(payload)
		if err != nil {
			return sdktypes.InvalidUserID, errInvalidToken
		}

		return u.ID(), nil
	}
}

func newInternalTokensMiddleware(tokens authtokens.Tokens) middlewareFn {
	return func(r *http.Request) (sdktypes.UserID, *MiddlewareError) {
		authHdr := r.Header.Get("Authorization")
		if authHdr == "" {
			return sdktypes.InvalidUserID, nil
		}

		kind, payload, _ := strings.Cut(authHdr, " ")

		if kind != "Bearer" {
			return sdktypes.InvalidUserID, errInvalidAuthHeader
		}

		_, err := tokens.ParseInternal(payload)
		if err != nil {
			return sdktypes.InvalidUserID, errInvalidToken
		}

		return authusers.SystemUser.ID(), nil
	}
}

func newSessionsMiddleware(sessions authsessions.Store) middlewareFn {
	return func(r *http.Request) (sdktypes.UserID, *MiddlewareError) {
		user, err := sessions.Get(r)
		// Do not fail on error - graceful degradation in case of session structure changes.
		if err == nil && user.IsValid() {
			return user.ID(), nil
		}

		return sdktypes.InvalidUserID, nil
	}
}

func setDefaultUserMiddleware(r *http.Request) (sdktypes.UserID, *MiddlewareError) {
	return authusers.DefaultUser.ID(), nil
}

func New(deps Deps) (get GetUserFromRequestFunc, mw AuthMiddlewareDecorator) {
	sessions, users, tokens := deps.Sessions, deps.Users, deps.Tokens

	var mws []middlewareFn

	if tokens != nil {
		mws = append(mws, newTokensMiddleware(tokens))
		mws = append(mws, newInternalTokensMiddleware(tokens))
	}

	if sessions != nil {
		mws = append(mws, newSessionsMiddleware(sessions))
	}

	if deps.Users.HasDefaultUser() {
		mws = append(mws, setDefaultUserMiddleware)
	}

	get = func(r *http.Request) (u sdktypes.User, err error) {
		var uid sdktypes.UserID

		// Process all middlewares.
		for _, mw := range mws {
			// Stop processing middlewares if a valid user ID is found or there is an error.
			var tempErr *MiddlewareError
			uid, tempErr = mw(r)

			if err == nil && tempErr != nil {
				err = tempErr
			}

			if uid.IsValid() {
				err = nil // reset error if we found a valid user ID
				break
			}
		}

		if err != nil {
			return
		}

		// Check if authenticated.
		if !uid.IsValid() {
			// This is not an error.
			return
		}

		ctx := r.Context()

		if authusers.IsSystemUserID(uid) {
			u = authusers.SystemUser
		} else {
			u, err = users.Get(authcontext.SetAuthnSystemUser(ctx), uid, "")
			if err != nil {
				err = errUnknownUser
				return
			}

			if u.Status() != sdktypes.UserStatusActive {
				err = errUserIsNotActive
				return
			}
		}

		return
	}

	mw = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := deps.Logger

			u, err := get(r)
			if err != nil {
				var mwErr *MiddlewareError
				if ok := errors.As(err, &mwErr); ok {
					l.Info("authentication middleware error", zap.Error(mwErr))
					http.Error(w, mwErr.msg, mwErr.code)
					return
				}

				l.Error("unexpected error in authentication middleware", zap.Error(err))
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if !u.IsValid() {
				l.Info("unauthenticated request")
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				return
			}

			l.Debug("authenticated request", zap.String("user", u.String()))

			ctx := authcontext.SetAuthnUser(r.Context(), u)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}

	return
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
