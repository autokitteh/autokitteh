package authloginhttpsvc

import (
	"context"
	"net/http"
)

const redirectCookieName = "ak_redirect"

func setRedirectCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if q := r.URL.Query(); q.Has("redirect") {
			http.SetCookie(w, &http.Cookie{
				Name:  redirectCookieName,
				Value: q.Get("redirect"),
			})
		}

		next.ServeHTTP(w, r)
	})
}

type redirectKey string

const redirectCtxKey = redirectKey("redirect")

func extractRedirectFromCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, _ := r.Cookie(redirectCookieName); cookie != nil {
			r = r.WithContext(
				context.WithValue(r.Context(), redirectCtxKey, cookie.Value),
			)
		}

		next.ServeHTTP(w, r)
	})
}
