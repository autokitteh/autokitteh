package authz

import "net/http"

// Enrich the request context with the check function.
func HTTPInterceptor(checkFunc CheckFunc, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(ContextWithCheckFunc(r.Context(), checkFunc))
		h.ServeHTTP(w, r)
	})
}
