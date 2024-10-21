package muxes

import "net/http"

type (
	MuxData struct {
		Addr         func() string // available only after start.
		Auth, NoAuth *http.ServeMux
	}

	Muxes struct {
		MainURL string

		Main MuxData
		Aux  MuxData
	}
)
