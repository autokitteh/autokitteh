package common

import (
	"net/http"
)

func HTTPError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
