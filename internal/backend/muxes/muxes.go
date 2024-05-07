package muxes

import "net/http"

type Muxes struct {
	Auth, NoAuth *http.ServeMux
}
