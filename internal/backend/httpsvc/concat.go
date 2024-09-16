package httpsvc

import "net/http"

type catch404ResponseWriter struct {
	http.ResponseWriter

	Code    int
	headers http.Header
}

func (w *catch404ResponseWriter) Header() http.Header {
	if w.headers == nil {
		w.headers = make(http.Header)
	}
	return w.headers
}

func (w *catch404ResponseWriter) WriteHeader(code int) {
	w.Code = code
	if code == http.StatusNotFound {
		return
	}

	for k := range w.headers {
		for _, v := range w.headers[k] {
			w.ResponseWriter.Header().Add(k, v)
		}
	}

	w.ResponseWriter.WriteHeader(code)
}

func (w *catch404ResponseWriter) Write(b []byte) (int, error) {
	if w.Code == http.StatusNotFound {
		return len(b), nil
	}

	return w.ResponseWriter.Write(b)
}

func NewConcatinatedHandler(hs ...http.Handler) http.Handler {
	if len(hs) == 0 {
		return http.NotFoundHandler()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for i, h := range hs {
			if i == len(hs)-1 {
				// last one will always just return.
				h.ServeHTTP(w, r)
				return
			}

			ww := &catch404ResponseWriter{ResponseWriter: w}

			h.ServeHTTP(ww, r)
			if ww.Code != http.StatusNotFound {
				return
			}
		}
	})
}
