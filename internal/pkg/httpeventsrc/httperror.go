package httpeventsrc

import (
	"fmt"
	"net/http"
)

type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string { return fmt.Sprintf("%d: %s", e.Code, e.Message) }

func (e *HTTPError) Write(w http.ResponseWriter) {
	w.WriteHeader(e.Code)
	_, _ = w.Write([]byte(e.Message))
}

func httpError(code int, f string, vs ...interface{}) *HTTPError {
	return &HTTPError{Code: code, Message: fmt.Sprintf(f, vs...)}
}

var _ error = &HTTPError{}
