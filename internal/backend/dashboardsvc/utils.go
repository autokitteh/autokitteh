package dashboardsvc

import (
	"net/http"
	"strconv"
)

var (
	getQueryNum  = mkGetQueryParam(strconv.Atoi)
	getQueryBool = mkGetQueryParam(strconv.ParseBool)
)

func mkGetQueryParam[T any](cvt func(string) (T, error)) func(r *http.Request, key string, def T) T {
	return func(r *http.Request, key string, def T) T {
		v := r.URL.Query().Get(key)
		if v == "" {
			return def
		}

		n, err := cvt(v)
		if err != nil {
			return def
		}
		return n
	}
}
