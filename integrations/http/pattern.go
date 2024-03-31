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

func extractKeys(s string) (keys []string, err error) {
	var (
		curr   strings.Builder
		inside bool
	)

	for _, r := range s {
		if inside {
			if r != '}' {
				curr.WriteRune(r)
			} else {
				keys = append(keys, strings.TrimSuffix(curr.String(), "..."))
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

func MatchPattern(pattern string, req string) (vs map[string]string, err error) {
	err = sdkerrors.ErrNotFound

	var mux http.ServeMux

	mux.Handle(pattern, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		var keys []string
		if keys, err = extractKeys(pattern); err == nil {
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
