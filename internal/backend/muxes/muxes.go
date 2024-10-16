package muxes

import "net/http"

type (
	muxes struct{ Auth, NoAuth *http.ServeMux }

	Muxes    muxes
	AuxMuxes muxes
)
