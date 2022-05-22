package h

import (
	"errors"
	"fmt"
	"net/http"

	L "github.com/autokitteh/L"
)

type errHTTP struct {
	msg   string
	pairs []interface{}
	code  int
}

var _ error = &errHTTP{}

func (err *errHTTP) Code() int { return err.code }

func (err *errHTTP) Error() string {
	text := fmt.Sprintf("%s\n%s\n", http.StatusText(err.code), err.msg)

	pairs := err.pairs

	for len(pairs) > 1 {
		text += fmt.Sprintf("  %s: %v\n", pairs[0], pairs[1])
		pairs = pairs[2:]
	}

	if len(pairs) != 0 {
		text += "(invalid error: pairs not odd)"
	}

	return text
}

func (err *errHTTP) Respond(l L.L, w http.ResponseWriter) {
	l = L.N(l)

	f := l.Debug
	if err.code >= 500 {
		f = l.Error
	} else if err.code >= 400 {
		f = l.Warn
	}

	f(err.msg, err.pairs...)

	http.Error(w, err.Error(), err.code)
}

func GetHTTPError(err error) *errHTTP {
	var errh *errHTTP
	if errors.As(err, &errh) {
		return errh
	}
	return nil
}

func RespondOnHTTPError(l L.L, w http.ResponseWriter, err error) bool {
	if err := GetHTTPError(err); err != nil {
		err.Respond(l, w)
		return true
	}

	return false
}

func Respond(l L.L, w http.ResponseWriter, err error) {
	if RespondOnHTTPError(l, w, err) {
		return
	}

	NewError(http.StatusInternalServerError, "err", err.Error()).Respond(l, w)
}

func NewError(code int, msg string, pairs ...interface{}) *errHTTP {
	return &errHTTP{code: code, msg: msg, pairs: pairs}
}

func Error(l L.L, w http.ResponseWriter, code int, msg string, pairs ...interface{}) {
	NewError(code, msg, pairs...).Respond(l, w)
}
