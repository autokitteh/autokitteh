// Taken from https://github.com/egtann/strip-wildcard-prefix/blob/main/strip_wildcard.go.
// See discussion in https://github.com/golang/go/issues/64909.
package kittehs

import (
	"net/http"
	"net/url"
	"strings"
)

func StripWildcardPrefix(prefix string, h http.Handler) http.Handler {
	staticSegments := []int{} // Indices of static segments in the prefix
	prefixSegments := strings.Split(prefix, "/")

	for i, segment := range prefixSegments {
		if !strings.Contains(segment, "{") || !strings.Contains(segment, "}") {
			staticSegments = append(staticSegments, i)
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathSegments := strings.Split(r.URL.Path, "/")
		matched := true

		for _, idx := range staticSegments {
			if idx >= len(pathSegments) || prefixSegments[idx] != pathSegments[idx] {
				matched = false
				break
			}
		}

		if !matched {
			http.NotFound(w, r)
			return
		}

		newPath := "/" + strings.Join(pathSegments[len(prefixSegments):], "/")

		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = newPath

		h.ServeHTTP(w, r2)
	})
}
