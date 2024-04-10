package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

// Extract path keys from path. This follows Go's convention for paths
// as described in https://pkg.go.dev/net/http#ServeMux.
// Essentially, extract the content of any curly braces in the path.
// If a key has a "..." suffix, it is removed.
// If a key is "$", it is ignored.
func extractPathKeys(s string) (keys []string, err error) {
	var (
		curr   strings.Builder
		inside bool
	)

	for _, r := range s {
		if inside {
			if r != '}' {
				curr.WriteRune(r)
			} else {
				key := strings.TrimSuffix(curr.String(), "...")
				if key == "" {
					return nil, sdkerrors.NewInvalidArgumentError("empty key")
				}
				if key != "$" {
					keys = append(keys, key)
				}
				curr.Reset()
				inside = false
			}
		} else if r == '{' {
			inside = true
		}
	}

	if inside {
		err = sdkerrors.NewInvalidArgumentError("unmatched '{'")
	}

	return
}

// Match request path against pattern. This uses go's http.ServeMux as described
// in https://pkg.go.dev/net/http#ServeMux. It will return a map of path keys
// to their values if the request path matches the pattern.
func MatchPattern(pattern string, req string) (vs map[string]string, err error) {
	err = sdkerrors.ErrNotFound

	var mux http.ServeMux

	mux.Handle(pattern, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		var keys []string
		if keys, err = extractPathKeys(pattern); err == nil {
			vs = kittehs.ListToMap(keys, func(k string) (string, string) {
				return k, r.PathValue(k)
			})
		}
	}))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, &http.Request{URL: &url.URL{Path: req}})
	if code := w.Result().StatusCode; code == http.StatusNotFound {
		err = sdkerrors.ErrNotFound
	} else if code != http.StatusOK {
		err = errors.New(w.Result().Status)
	}

	return
}
