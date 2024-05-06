package auth

import (
	"context"
	"net/http"
)

type noAuthAuthenticator struct{}

func (*noAuthAuthenticator) Middleware(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		ctx = context.WithValue(ctx, AuthenticatedUserCtxKey, DefaultUser)

		n.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (*noAuthAuthenticator) Provider() ProviderDetails {
	return ProviderDetails{Name: AuthProviderNone, Config: map[string]string{}}
}
